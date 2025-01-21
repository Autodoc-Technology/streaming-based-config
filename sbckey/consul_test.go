package sbckey_test

import (
	"testing"

	"github.com/Autodoc-Technology/streaming-based-config/sbckey"
)

func TestConsulDefaultKeyBuilderWithInt(t *testing.T) {
	builder := sbckey.ConsulDefaultKeyBuilder[int](sbckey.NoPrefix)
	result := builder.BuildKey(123)
	if result != "int" {
		t.Errorf("Expected 'int', got '%s'", result)
	}
}

func TestConsulDefaultKeyBuilderWithString(t *testing.T) {
	builder := sbckey.ConsulDefaultKeyBuilder[string](sbckey.NoPrefix)
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
	builder := sbckey.ConsulDefaultKeyBuilder[TestStruct](sbckey.NoPrefix)
	result := builder.BuildKey(TestStruct{"value", 123})
	if result != "sbckey_test.TestStruct" {
		t.Errorf("Expected 'sbckey_test.TestStruct', got '%s'", result)
	}
}

func TestConsulDefaultKeyBuilderWithNil(t *testing.T) {
	builder := sbckey.ConsulDefaultKeyBuilder[*int](sbckey.NoPrefix)
	result := builder.BuildKey(nil)
	if result != "*int" {
		t.Errorf("Expected '*int', got '%s'", result)
	}
}

func TestConsulDefaultKeyBuilderWithPrefix(t *testing.T) {
	builder := sbckey.ConsulDefaultKeyBuilder[int]("prefix/")
	result := builder.BuildKey(123)
	if result != "prefix/int" {
		t.Errorf("Expected 'prefix/int', got '%s'", result)
	}
}
