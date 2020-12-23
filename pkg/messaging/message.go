package messaging

import (
	"encoding/json"
)

const (
	ServerUID        = "tarpon"
	ctrlDisconnected = "peer_disconnected"
	ctrlConnected    = "peer_connected"
)

type Message struct {
	From    string          `json:"from"`
	To      string          `json:"to"`
	Payload json.RawMessage `json:"payload"`
}

func (m *Message) IsBroadcast() bool {
	return m.To == ""
}

type controlPayload struct {
	Type string `json:"type"`
	Peer string `json:"peer"`
}

func NewPeerDisconnected(peerUID string) (*Message, error) {
	payload := controlPayload{
		Type: ctrlDisconnected,
		Peer: peerUID,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &Message{
		From:    ServerUID,
		To:      "",
		Payload: jsonPayload,
	}, nil
}

func NewPeerConnected(peerUID string) (*Message, error) {
	payload := controlPayload{
		Type: ctrlConnected,
		Peer: peerUID,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &Message{
		From:    ServerUID,
		To:      "",
		Payload: jsonPayload,
	}, nil
}
