package sbcencoder_test

import (
	"testing"

	"github.com/Autodoc-Technology/streaming-based-config/sbcencoder"
)

func TestJsonEncoderEncodeValidData(t *testing.T) {
	encoder := sbcencoder.NewJsonEncoder()
	data, err := encoder.Encode(map[string]string{"key": "value"})
	if err != nil || string(data) != `{"key":"value"}` {
		t.Fail()
	}
}

func TestJsonEncoderEncodeInvalidData(t *testing.T) {
	encoder := sbcencoder.NewJsonEncoder()
	_, err := encoder.Encode(make(chan int))
	if err == nil {
		t.Fail()
	}
}

func TestJsonEncoderDecodeValidData(t *testing.T) {
	encoder := sbcencoder.NewJsonEncoder()
	var v map[string]string
	err := encoder.Decode([]byte(`{"key":"value"}`), &v)
	if err != nil || v["key"] != "value" {
		t.Fail()
	}
}

func TestJsonEncoderDecodeInvalidData(t *testing.T) {
	encoder := sbcencoder.NewJsonEncoder()
	var v map[string]string
	err := encoder.Decode([]byte(`{"key":`), &v)
	if err == nil {
		t.Fail()
	}
}
