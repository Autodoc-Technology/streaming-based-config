package sbckey_test

import (
	"github.com/Autodoc-Technology/streaming-based-config/sbckey"
	"testing"
)

func TestNatsDefaultKeyBuilderWithInt(t *testing.T) {
	builder := sbckey.NatsDefaultKeyBuilder[int]()
	result := builder.BuildKey(123)
	if result != "int" {
		t.Errorf("Expected 'int', got '%s'", result)
	}
}

func TestNatsDefaultKeyBuilderWithString(t *testing.T) {
	builder := sbckey.NatsDefaultKeyBuilder[string]()
	result := builder.BuildKey("test")
	if result != "string" {
		t.Errorf("Expected 'string', got '%s'", result)
	}
}

func TestNatsDefaultKeyBuilderWithStruct(t *testing.T) {
	type TestStruct struct {
		Field1 string
		Field2 int
	}
	builder := sbckey.NatsDefaultKeyBuilder[TestStruct]()
	result := builder.BuildKey(TestStruct{"value", 123})
	if result != "sbckey_test.TestStruct" {
		t.Errorf("Expected 'sbckey_test.TestStruct', got '%s'", result)
	}
}

func TestNatsDefaultKeyBuilderWithInvalidCharacters(t *testing.T) {
	builder := sbckey.NatsDefaultKeyBuilder[string]()
	result := builder.BuildKey("test*")
	if result != "string" {
		t.Errorf("Expected 'string', got '%s'", result)
	}
}

func TestNatsDefaultKeyBuilderWithNil(t *testing.T) {
	builder := sbckey.NatsDefaultKeyBuilder[*int]()
	result := builder.BuildKey(nil)
	if result != "_int" {
		t.Errorf("Expected '_int', got '%s'", result)
	}
}
