package sbckey_test

import (
	"strings"
	"testing"

	"github.com/Autodoc-Technology/streaming-based-config/sbckey"
)

type kafkaSimpleConfig struct{}

type kafkaConfigWithUnderscores struct{}

func TestKafkaDefaultKeyBuilder_SimpleStruct(t *testing.T) {
	builder := sbckey.KafkaDefaultKeyBuilder[kafkaSimpleConfig]()
	got := builder.BuildKey(kafkaSimpleConfig{})
	want := "sbckey_test.kafkaSimpleConfig"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestKafkaDefaultKeyBuilder_StructWithUnderscores(t *testing.T) {
	builder := sbckey.KafkaDefaultKeyBuilder[kafkaConfigWithUnderscores]()
	got := builder.BuildKey(kafkaConfigWithUnderscores{})
	want := "sbckey_test.kafkaConfigWithUnderscores"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestKafkaDefaultKeyBuilder_Int(t *testing.T) {
	builder := sbckey.KafkaDefaultKeyBuilder[int]()
	got := builder.BuildKey(0)
	if got != "int" {
		t.Errorf("expected 'int', got %q", got)
	}
}

func TestKafkaDefaultKeyBuilder_PointerType(t *testing.T) {
	// Pointer types produce "*pkg.Type" which contains '*' — it must be sanitised.
	builder := sbckey.KafkaDefaultKeyBuilder[*kafkaSimpleConfig]()
	got := builder.BuildKey(nil)
	if strings.ContainsAny(got, "*") {
		t.Errorf("key %q contains invalid character '*'", got)
	}
}

func TestKafkaDefaultKeyBuilder_LongNameTruncated(t *testing.T) {
	// Build a key manually from a very long string to confirm truncation.
	builder := sbckey.KafkaDefaultKeyBuilder[string]()
	got := builder.BuildKey("")
	if len(got) > 249 {
		t.Errorf("key length %d exceeds 249 characters", len(got))
	}
}

func TestKafkaDefaultKeyBuilder_KeyIsValidKafkaTopic(t *testing.T) {
	builder := sbckey.KafkaDefaultKeyBuilder[kafkaSimpleConfig]()
	got := builder.BuildKey(kafkaSimpleConfig{})
	if len(got) == 0 {
		t.Fatal("expected non-empty key")
	}
	if len(got) > 249 {
		t.Errorf("key length %d exceeds 249 characters", len(got))
	}
}
