package messaging

import "encoding/json"

type Message struct {
	From    string          `json:"from"`
	To      string          `json:"to"`
	Payload json.RawMessage `json:"payload"`
}

func (m *Message) IsBroadcast() bool {
	return m.To == ""
}

type ControlPayload struct {
	Type string `json:"type"`
	Peer string `json:"peer"`
}

func NewPeerDisconnected(peerUID string) *ControlPayload {
	return &ControlPayload{
		Type: "peer_disconnected",
		Peer: peerUID,
	}
}

func NewPeerConnected(peerUID string) *ControlPayload {
	return &ControlPayload{
		Type: "peer_connected",
		Peer: peerUID,
	}
}
