package sbc

import (
	"testing"
)

func TestKeyBuilderFuncBuildKeyWithInt(t *testing.T) {
	builder := KeyBuilderFunc[int](func(i int) string { return "int" })
	result := builder.BuildKey(123)
	if result != "int" {
		t.Errorf("Expected 'int', got '%s'", result)
	}
}

func TestKeyBuilderFuncBuildKeyWithString(t *testing.T) {
	builder := KeyBuilderFunc[string](func(s string) string { return "string" })
	result := builder.BuildKey("test")
	if result != "string" {
		t.Errorf("Expected 'string', got '%s'", result)
	}
}

func TestKeyBuilderFuncBuildKeyWithStruct(t *testing.T) {
	type TestStruct struct {
		Field1 string
		Field2 int
	}
	builder := KeyBuilderFunc[TestStruct](func(ts TestStruct) string { return "TestStruct" })
	result := builder.BuildKey(TestStruct{"value", 123})
	if result != "TestStruct" {
		t.Errorf("Expected 'TestStruct', got '%s'", result)
	}
}
