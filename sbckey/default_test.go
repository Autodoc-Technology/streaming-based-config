package sbckey_test

import (
	"fmt"
	"testing"

	sbc "github.com/Autodoc-Technology/streaming-based-config"
	"github.com/Autodoc-Technology/streaming-based-config/sbckey"
)

func TestKeyBuilderFuncWithInt(t *testing.T) {
	builder := sbc.KeyBuilderFunc[int](func(t int) string {
		return fmt.Sprintf("%d", t)
	})
	result := builder.BuildKey(123)
	if result != "123" {
		t.Errorf("Expected '123', got '%s'", result)
	}
}

func TestKeyBuilderFuncWithString(t *testing.T) {
	builder := sbc.KeyBuilderFunc[string](func(t string) string {
		return t
	})
	result := builder.BuildKey("test")
	if result != "test" {
		t.Errorf("Expected 'test', got '%s'", result)
	}
}

func TestKeyBuilderFuncWithStruct(t *testing.T) {
	type TestStruct struct {
		Field1 string
		Field2 int
	}
	builder := sbc.KeyBuilderFunc[TestStruct](func(t TestStruct) string {
		return fmt.Sprintf("%s%d", t.Field1, t.Field2)
	})
	result := builder.BuildKey(TestStruct{"value", 123})
	if result != "value123" {
		t.Errorf("Expected 'value123', got '%s'", result)
	}
}

func TestKeyBuilderFuncWithNil(t *testing.T) {
	builder := sbc.KeyBuilderFunc[*int](func(t *int) string {
		if t == nil {
			return "nil"
		}
		return fmt.Sprintf("%d", *t)
	})
	result := builder.BuildKey(nil)
	if result != "nil" {
		t.Errorf("Expected 'nil', got '%s'", result)
	}
}

func TestDefaultKeyBuilderWithInt(t *testing.T) {
	builder := sbckey.DefaultKeyBuilder[int]()
	result := builder.BuildKey(123)
	if result != "int" {
		t.Errorf("Expected 'int', got '%s'", result)
	}
}

func TestDefaultKeyBuilderWithString(t *testing.T) {
	builder := sbckey.DefaultKeyBuilder[string]()
	result := builder.BuildKey("test")
	if result != "string" {
		t.Errorf("Expected 'string', got '%s'", result)
	}
}

func TestDefaultKeyBuilderWithStruct(t *testing.T) {
	type TestStruct struct {
		Field1 string
		Field2 int
	}
	builder := sbckey.DefaultKeyBuilder[TestStruct]()
	result := builder.BuildKey(TestStruct{"value", 123})
	if result != "sbckey_test.TestStruct" {
		t.Errorf("Expected 'sbckey_test.TestStruct', got '%s'", result)
	}
}
