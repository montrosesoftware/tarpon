package messaging

import (
	"github.com/vmihailenco/msgpack/v4"
)

// Message represents data sent between subscribers.
type Message struct {
	Payload string `json:"payload"`
	FromUID string `json:"from_uid"`
}

// Decode converts binary representation to Message.
func Decode(b []byte) (*Message, error) {
	var msg Message
	err := msgpack.Unmarshal(b, &msg)
	return &msg, err
}

// Encode converts Message to binary representation.
func (m *Message) Encode() ([]byte, error) {
	return msgpack.Marshal(m)
}
