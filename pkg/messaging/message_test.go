package messaging_test

import (
	"reflect"
	"testing"

	"github.com/montrosesoftware/face-poker-messaging/pkg/messaging"
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
		t.Errorf("expected msg %v but got %v", m, msg)
	}

	_, err = messaging.Decode([]byte("msg"))
	if err == nil {
		t.Error("expected error but got nil")
	}
}
