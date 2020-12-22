package messaging

import (
	"encoding/json"

	"github.com/montrosesoftware/tarpon/pkg/logging"
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

func NewPeerDisconnected(logger logging.Logger, peerUID string) Message {
	payload := controlPayload{
		Type: ctrlDisconnected,
		Peer: peerUID,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		logger.Error("failed to marshal control payload", logging.Fields{"type": ctrlDisconnected, "peer": peerUID})
	}

	return Message{
		From:    ServerUID,
		To:      "",
		Payload: jsonPayload,
	}
}

func NewPeerConnected(logger logging.Logger, peerUID string) Message {
	payload := controlPayload{
		Type: ctrlConnected,
		Peer: peerUID,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		logger.Error("failed to marshal control payload", logging.Fields{"type": ctrlConnected, "peer": peerUID})
	}

	return Message{
		From:    ServerUID,
		To:      "",
		Payload: jsonPayload,
	}
}
