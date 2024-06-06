package sbckey_test

import (
	"github.com/Autodoc-Technology/streaming-based-config/sbckey"
	"testing"
)

func TestConsulDefaultKeyBuilderWithInt(t *testing.T) {
	builder := sbckey.ConsulDefaultKeyBuilder[int]()
	result := builder.BuildKey(123)
	if result != "int" {
		t.Errorf("Expected 'int', got '%s'", result)
	}
}

func TestConsulDefaultKeyBuilderWithString(t *testing.T) {
	builder := sbckey.ConsulDefaultKeyBuilder[string]()
	result := builder.BuildKey("test")
	if result != "string" {
		t.Errorf("Expected 'string', got '%s'", result)
	}
}

func TestConsulDefaultKeyBuilderWithStruct(t *testing.T) {
	type TestStruct struct {
		Field1 string
		Field2 int
	}
	builder := sbckey.ConsulDefaultKeyBuilder[TestStruct]()
	result := builder.BuildKey(TestStruct{"value", 123})
	if result != "sbckey_test.TestStruct" {
		t.Errorf("Expected 'sbckey_test.TestStruct', got '%s'", result)
	}
}

func TestConsulDefaultKeyBuilderWithNil(t *testing.T) {
	builder := sbckey.ConsulDefaultKeyBuilder[*int]()
	result := builder.BuildKey(nil)
	if result != "*int" {
		t.Errorf("Expected '*int', got '%s'", result)
	}
}
