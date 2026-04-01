package sbctransport

import (
	"context"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
)

// ---- helpers for ConsumerGroupSession and ConsumerGroupClaim mocks ----

type testConsumerGroupSession struct {
	ctx context.Context
}

func (s *testConsumerGroupSession) Claims() map[string][]int32                       { return nil }
func (s *testConsumerGroupSession) MemberID() string                                 { return "test-member" }
func (s *testConsumerGroupSession) GenerationID() int32                              { return 1 }
func (s *testConsumerGroupSession) MarkOffset(_ string, _ int32, _ int64, _ string)  {}
func (s *testConsumerGroupSession) Commit()                                          {}
func (s *testConsumerGroupSession) ResetOffset(_ string, _ int32, _ int64, _ string) {}
func (s *testConsumerGroupSession) MarkMessage(_ *sarama.ConsumerMessage, _ string)  {}
func (s *testConsumerGroupSession) Context() context.Context                         { return s.ctx }

type testConsumerGroupClaim struct {
	topic     string
	partition int32
	msgs      chan *sarama.ConsumerMessage
}

func (c *testConsumerGroupClaim) Topic() string                            { return c.topic }
func (c *testConsumerGroupClaim) Partition() int32                         { return c.partition }
func (c *testConsumerGroupClaim) InitialOffset() int64                     { return 0 }
func (c *testConsumerGroupClaim) HighWaterMarkOffset() int64               { return 0 }
func (c *testConsumerGroupClaim) Messages() <-chan *sarama.ConsumerMessage { return c.msgs }

// ---- testConsumerGroup: minimal sarama.ConsumerGroup for Updates tests ----

type testConsumerGroup struct {
	// consumeFn is called once per Consume call; after it returns the group loop retries.
	consumeFn func(ctx context.Context, handler sarama.ConsumerGroupHandler) error
	closed    chan struct{}
}

func newTestConsumerGroup(consumeFn func(context.Context, sarama.ConsumerGroupHandler) error) *testConsumerGroup {
	return &testConsumerGroup{consumeFn: consumeFn, closed: make(chan struct{})}
}

func (g *testConsumerGroup) Consume(ctx context.Context, _ []string, handler sarama.ConsumerGroupHandler) error {
	return g.consumeFn(ctx, handler)
}
func (g *testConsumerGroup) Errors() <-chan error { return nil }
func (g *testConsumerGroup) Close() error {
	select {
	case <-g.closed:
	default:
		close(g.closed)
	}
	return nil
}
func (g *testConsumerGroup) Pause(_ map[string][]int32)  {}
func (g *testConsumerGroup) Resume(_ map[string][]int32) {}
func (g *testConsumerGroup) PauseAll()                   {}
func (g *testConsumerGroup) ResumeAll()                  {}

// ---- tests ----

func TestKafkaConsumerGroupHandler(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan []byte, 2)
	handler := &kafkaConsumerGroupHandler{ch: ch, ctx: ctx}

	session := &testConsumerGroupSession{ctx: ctx}
	msgs := make(chan *sarama.ConsumerMessage, 2)
	claim := &testConsumerGroupClaim{topic: "test-topic", partition: 0, msgs: msgs}

	msgs <- &sarama.ConsumerMessage{Value: []byte("msg-1")}
	msgs <- &sarama.ConsumerMessage{Value: []byte("msg-2")}
	close(msgs)

	if err := handler.ConsumeClaim(session, claim); err != nil {
		t.Fatalf("ConsumeClaim returned error: %v", err)
	}

	for _, want := range []string{"msg-1", "msg-2"} {
		select {
		case got := <-ch:
			if string(got) != want {
				t.Errorf("expected %q, got %q", want, string(got))
			}
		case <-time.After(time.Second):
			t.Fatalf("timeout waiting for message %q", want)
		}
	}
}

func TestKafkaConsumerGroupHandler_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	ch := make(chan []byte)
	handler := &kafkaConsumerGroupHandler{ch: ch, ctx: ctx}

	session := &testConsumerGroupSession{ctx: ctx}
	msgs := make(chan *sarama.ConsumerMessage)
	claim := &testConsumerGroupClaim{topic: "test-topic", partition: 0, msgs: msgs}

	done := make(chan error, 1)
	go func() {
		done <- handler.ConsumeClaim(session, claim)
	}()

	// Cancel context; handler should exit cleanly.
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("expected nil error on context cancel, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("ConsumeClaim did not return after context cancel")
	}
}

func TestKafkaTransport_Current(t *testing.T) {
	const (
		topic        = "config-topic"
		partition    = int32(0)
		newestOffset = int64(5)
		wantPayload  = `{"env":"prod"}`
	)

	mockConsumer := mocks.NewConsumer(t, nil)
	pc := mockConsumer.ExpectConsumePartition(topic, partition, newestOffset-1)
	pc.YieldMessage(&sarama.ConsumerMessage{
		Topic:     topic,
		Partition: partition,
		Offset:    newestOffset - 1,
		Value:     []byte(wantPayload),
	})

	transport := newKafkaTransportWithClient(nil)
	transport.getOffsetFn = func(_ string, _ int32) (int64, error) {
		return newestOffset, nil
	}
	transport.consumerFactory = func(_ sarama.Client) (sarama.Consumer, error) {
		return mockConsumer, nil
	}

	got, err := transport.Current(context.Background(), topic)
	if err != nil {
		t.Fatalf("Current returned error: %v", err)
	}
	if string(got) != wantPayload {
		t.Errorf("expected %q, got %q", wantPayload, string(got))
	}
}

func TestKafkaTransport_Current_NoMessages(t *testing.T) {
	transport := newKafkaTransportWithClient(nil)
	transport.getOffsetFn = func(_ string, _ int32) (int64, error) {
		return 0, nil // offset 0 means no messages
	}

	_, err := transport.Current(context.Background(), "empty-topic")
	if err == nil {
		t.Fatal("expected error for topic with no messages, got nil")
	}
}

func TestKafkaTransport_Current_ContextCancel(t *testing.T) {
	// Consumer that never delivers a message — context cancel should unblock Current.
	mockConsumer := mocks.NewConsumer(t, nil)
	mockConsumer.ExpectConsumePartition("my-topic", 0, 4)
	// Do NOT yield a message; context cancel must free Current.

	ctx, cancel := context.WithCancel(context.Background())

	transport := newKafkaTransportWithClient(nil)
	transport.getOffsetFn = func(_ string, _ int32) (int64, error) { return 5, nil }
	transport.consumerFactory = func(_ sarama.Client) (sarama.Consumer, error) {
		return mockConsumer, nil
	}

	done := make(chan error, 1)
	go func() {
		_, err := transport.Current(ctx, "my-topic")
		done <- err
	}()

	cancel()

	select {
	case err := <-done:
		if err == nil {
			t.Error("expected context error, got nil")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Current did not return after context cancel")
	}
}

func TestKafkaTransport_Updates(t *testing.T) {
	const wantPayload = `{"feature":"on"}`

	callCount := 0
	group := newTestConsumerGroup(func(ctx context.Context, handler sarama.ConsumerGroupHandler) error {
		callCount++
		if callCount > 1 {
			// After the first consume cycle, block until context is done so the
			// reconnect loop exits cleanly.
			<-ctx.Done()
			return nil
		}

		session := &testConsumerGroupSession{ctx: ctx}
		msgs := make(chan *sarama.ConsumerMessage, 1)
		msgs <- &sarama.ConsumerMessage{Value: []byte(wantPayload)}
		close(msgs)
		claim := &testConsumerGroupClaim{topic: "cfg", msgs: msgs}

		_ = handler.Setup(session)
		_ = handler.ConsumeClaim(session, claim)
		_ = handler.Cleanup(session)
		return nil
	})

	transport := newKafkaTransportWithClient(nil)
	transport.groupFactory = func(_ string, _ sarama.Client) (sarama.ConsumerGroup, error) {
		return group, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	updCh, err := transport.Updates(ctx, "cfg")
	if err != nil {
		t.Fatalf("Updates returned error: %v", err)
	}

	select {
	case got := <-updCh:
		if string(got) != wantPayload {
			t.Errorf("expected %q, got %q", wantPayload, string(got))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for update message")
	}

	// Cancel and ensure channel is closed cleanly.
	cancel()
	select {
	case _, open := <-updCh:
		if open {
			t.Error("expected updates channel to be closed after context cancel")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("updates channel not closed after context cancel")
	}
}

func TestKafkaTransportOpt_WithKafkaGroupID(t *testing.T) {
	transport := newKafkaTransportWithClient(nil, WithKafkaGroupID("my-group"))
	if transport.groupID != "my-group" {
		t.Errorf("expected groupID 'my-group', got %q", transport.groupID)
	}
}

func TestKafkaTransportOpt_WithKafkaGroupID_Empty(t *testing.T) {
	transport := newKafkaTransportWithClient(nil, WithKafkaGroupID(""))
	if transport.groupID != "streaming-based-config" {
		t.Errorf("expected default groupID, got %q", transport.groupID)
	}
}

func TestKafkaTransportOpt_WithKafkaPartition(t *testing.T) {
	transport := newKafkaTransportWithClient(nil, WithKafkaPartition(2))
	if transport.partition != 2 {
		t.Errorf("expected partition 2, got %d", transport.partition)
	}
}
