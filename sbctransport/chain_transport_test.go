package sbctransport

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Autodoc-Technology/streaming-based-config/sbcencoder"
)

// ManualTransport allows manual control over Current value and emitting Updates.
type ManualTransport struct {
	current []byte
	ch      chan []byte
}

func NewManualTransport(initial string) *ManualTransport {
	return &ManualTransport{
		current: []byte(initial),
		ch:      make(chan []byte),
	}
}

func (m *ManualTransport) Current(ctx context.Context, key string) ([]byte, error) {
	return m.current, nil
}

func (m *ManualTransport) Updates(ctx context.Context, key string) (<-chan []byte, error) {
	return m.ch, nil
}

func (m *ManualTransport) Emit(val string) {
	m.current = []byte(val)
	m.ch <- m.current
}

func TestChainTransport_Current(t *testing.T) {
	enc := sbcencoder.NewJsonEncoder()
	
	t1 := NewManualTransport(`{"a": 1, "b": 2}`)
	t2 := NewManualTransport(`{"b": 3, "c": 4}`)

	chain := NewChainTransport(enc, t1, t2)

	val, err := chain.Current(context.Background(), "test")
	if err != nil {
		t.Fatalf("Current failed: %v", err)
	}

	var res map[string]interface{}
	if err := json.Unmarshal(val, &res); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Expect a=1 (from t1), b=3 (from t2 override), c=4 (from t2)
	assertVal(t, res, "a", 1.0)
	assertVal(t, res, "b", 3.0)
	assertVal(t, res, "c", 4.0)
}

func TestChainTransport_Updates(t *testing.T) {
	enc := sbcencoder.NewJsonEncoder()

	t1 := NewManualTransport(`{"a": 1, "b": 2}`)
	t2 := NewManualTransport(`{"b": 3, "c": 4}`)

	chain := NewChainTransport(enc, t1, t2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	updCh, err := chain.Updates(ctx, "test")
	if err != nil {
		t.Fatalf("Updates failed: %v", err)
	}

	// Helper to wait for update
	expectUpdate := func(expectedMap map[string]interface{}) {
		select {
		case val := <-updCh:
			var res map[string]interface{}
			if err := json.Unmarshal(val, &res); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}
			for k, v := range expectedMap {
				assertVal(t, res, k, v)
			}
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for update")
		}
	}

	// Update t1: a->10. Expect {a:10, b:3, c:4}
	go t1.Emit(`{"a": 10, "b": 2}`)
	expectUpdate(map[string]interface{}{"a": 10.0, "b": 3.0, "c": 4.0})

	// Update t2: b->30. Expect {a:10, b:30}. c is gone.
	go t2.Emit(`{"b": 30}`)
	// We handle the check manually to verify "c" is missing
	select {
	case val := <-updCh:
		var res map[string]interface{}
		if err := json.Unmarshal(val, &res); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		assertVal(t, res, "a", 10.0)
		assertVal(t, res, "b", 30.0)
		assertMissing(t, res, "c")
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for update")
	}
}

func assertVal(t *testing.T, m map[string]interface{}, key string, expected interface{}) {
	val, ok := m[key]
	if !ok {
		t.Errorf("Key %s missing", key)
		return
	}
	if val != expected {
		t.Errorf("Key %s: expected %v, got %v", key, expected, val)
	}
}

func assertMissing(t *testing.T, m map[string]interface{}, key string) {
	if _, ok := m[key]; ok {
		t.Errorf("Key %s present, expected missing", key)
	}
}
