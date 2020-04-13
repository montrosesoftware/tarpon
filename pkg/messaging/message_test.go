package messaging_test

import (
	"reflect"
	"testing"

	"github.com/montrosesoftware/tarpon/pkg/messaging"
)

func TestEncoding(t *testing.T) {
	m := &messaging.Message{Payload: `{type:"A",content:"data"}`, FromUID: "ABC"}
	bts, err := m.Encode()
	if err != nil {
		t.Errorf("encoding failed: %v", err)
	}
	msg, err := messaging.Decode(bts)
	if err != nil {
		t.Errorf("decoding failed: %v", err)
	}
	if !reflect.DeepEqual(m, msg) {
		t.Errorf("got %v but expceted %v", msg, m)
	}

	_, err = messaging.Decode([]byte("msg"))
	if err == nil {
		t.Error("got nil but expected error")
	}
}
