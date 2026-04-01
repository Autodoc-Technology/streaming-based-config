package sbctransport

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
)

// ErrNoMessages is returned by Current when the topic has no messages yet.
var ErrNoMessages = fmt.Errorf("topic has no messages")

// KafkaTransport is a transport that reads configuration from a Kafka topic.
// The key argument passed to Current and Updates is used as the Kafka topic name.
type KafkaTransport struct {
	client          sarama.Client
	groupID         string
	partition       int32
	consumerFactory func(sarama.Client) (sarama.Consumer, error)
	groupFactory    func(string, sarama.Client) (sarama.ConsumerGroup, error)
	// getOffsetFn retrieves the newest offset for a topic/partition.
	// Exposed as a field so tests can inject a stub without a real broker.
	getOffsetFn func(topic string, partition int32) (int64, error)
}

// KafkaTransportOpt is a functional option for KafkaTransport.
type KafkaTransportOpt func(*KafkaTransport)

// WithKafkaGroupID sets the consumer group ID used by Updates.
// Default is "streaming-based-config".
func WithKafkaGroupID(groupID string) KafkaTransportOpt {
	return func(k *KafkaTransport) {
		if groupID != "" {
			k.groupID = groupID
		}
	}
}

// WithKafkaPartition sets the partition used by Current to fetch the latest message.
// Default is 0.
func WithKafkaPartition(partition int32) KafkaTransportOpt {
	return func(k *KafkaTransport) {
		k.partition = partition
	}
}

// NewKafkaTransport creates a KafkaTransport connecting to the given brokers.
func NewKafkaTransport(brokers []string, opts ...KafkaTransportOpt) (*KafkaTransport, error) {
	cfg := sarama.NewConfig()
	cfg.Consumer.Return.Errors = true
	cfg.Version = sarama.V2_0_0_0

	client, err := sarama.NewClient(brokers, cfg)
	if err != nil {
		return nil, fmt.Errorf("creating kafka client: %w", err)
	}

	return newKafkaTransportWithClient(client, opts...), nil
}

// newKafkaTransportWithClient creates a KafkaTransport from an existing sarama.Client.
// Used in tests to inject mock clients/stubs without a real broker.
func newKafkaTransportWithClient(client sarama.Client, opts ...KafkaTransportOpt) *KafkaTransport {
	t := &KafkaTransport{
		client:          client,
		groupID:         "streaming-based-config",
		partition:       0,
		consumerFactory: sarama.NewConsumerFromClient,
		groupFactory: func(groupID string, c sarama.Client) (sarama.ConsumerGroup, error) {
			return sarama.NewConsumerGroupFromClient(groupID, c)
		},
	}
	t.getOffsetFn = func(topic string, partition int32) (int64, error) {
		return t.client.GetOffset(topic, partition, sarama.OffsetNewest)
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// Current retrieves the latest message value for the given topic (key).
// It fetches the newest offset and reads the message at that offset.
func (k *KafkaTransport) Current(ctx context.Context, key string) ([]byte, error) {
	newestOffset, err := k.getOffsetFn(key, k.partition)
	if err != nil {
		return nil, fmt.Errorf("getting newest offset for topic '%s': %w", key, err)
	}
	if newestOffset == 0 {
		return nil, fmt.Errorf("topic '%s': %w", key, ErrNoMessages)
	}

	consumer, err := k.consumerFactory(k.client)
	if err != nil {
		return nil, fmt.Errorf("creating kafka consumer for topic '%s': %w", key, err)
	}
	defer consumer.Close() //nolint:errcheck

	pc, err := consumer.ConsumePartition(key, k.partition, newestOffset-1)
	if err != nil {
		return nil, fmt.Errorf("consuming partition for topic '%s': %w", key, err)
	}
	defer pc.AsyncClose()

	select {
	case msg := <-pc.Messages():
		return msg.Value, nil
	case consumerErr := <-pc.Errors():
		return nil, fmt.Errorf("partition consumer error for topic '%s': %w", key, consumerErr.Err)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Updates returns a channel that streams new messages from the given topic (key)
// using a Kafka consumer group. The channel is closed when the context is cancelled.
// If the consumer group session ends and the context is still active it reconnects automatically.
func (k *KafkaTransport) Updates(ctx context.Context, key string) (<-chan []byte, error) {
	group, err := k.groupFactory(k.groupID, k.client)
	if err != nil {
		return nil, fmt.Errorf("creating kafka consumer group for topic '%s': %w", key, err)
	}

	ch := make(chan []byte)
	handler := &kafkaConsumerGroupHandler{ch: ch, ctx: ctx}

	go func() {
		defer close(ch)
		defer group.Close() //nolint:errcheck
		for {
			if err := group.Consume(ctx, []string{key}, handler); err != nil {
				return
			}
			// If context is done, stop the reconnect loop.
			if ctx.Err() != nil {
				return
			}
		}
	}()

	return ch, nil
}

// kafkaConsumerGroupHandler implements sarama.ConsumerGroupHandler and forwards
// consumed message values to the channel.
type kafkaConsumerGroupHandler struct {
	ch  chan []byte
	ctx context.Context
}

func (h *kafkaConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *kafkaConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *kafkaConsumerGroupHandler) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			session.MarkMessage(msg, "")
			select {
			case h.ch <- msg.Value:
			case <-h.ctx.Done():
				return nil
			}
		case <-session.Context().Done():
			return nil
		}
	}
}
