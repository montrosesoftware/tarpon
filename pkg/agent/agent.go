package agent

import (
	"encoding/json"
	"io"
	"log"

	"github.com/gorilla/websocket"
	"github.com/montrosesoftware/tarpon/pkg/messaging"
)

type Agent struct {
	peer messaging.Peer
	room string
	conn *websocket.Conn
}

func New(p messaging.Peer, r string, c *websocket.Conn) *Agent {
	return &Agent{peer: p, room: r, conn: c}
}

func (a *Agent) Start() {
	go func() {
		for {
			_, r, err := a.conn.NextReader()
			if err != nil {
				log.Printf("error reading from websocket: %v", err)
				break
			}
			a.handleClientMessage(r)
		}
	}()
}

func (a *Agent) handleClientMessage(r io.Reader) {
	var message struct {
		To      string          `json:"to"`
		Payload json.RawMessage `json:"payload"`
	}

	if err := json.NewDecoder(r).Decode(&message); err != nil {
		log.Printf("error decoding message: %v", err)
		return
	}

}
