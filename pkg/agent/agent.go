package agent

import (
	"encoding/json"
	"io"
	"log"

	"github.com/gorilla/websocket"
	"github.com/montrosesoftware/tarpon/pkg/messaging"
)

type Broker interface {
	Send(room string, message messaging.Message)
}

// Agent handles websocket communication between peers and the broker.
type Agent struct {
	peer   messaging.Peer
	room   string
	conn   *websocket.Conn
	broker Broker
}

func New(p messaging.Peer, r string, c *websocket.Conn, b Broker) *Agent {
	return &Agent{peer: p, room: r, conn: c, broker: b}
}

func HandlePeer(p messaging.Peer, room string, conn *websocket.Conn) {
	agent := New(p, room, conn, nil)
	agent.Start()
}

func (a *Agent) Start() {
	go func() {
		defer func() {
			a.conn.Close()
		}()

		a.conn.SetReadLimit(2048)

		for {
			_, r, err := a.conn.NextReader()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("error reading from websocket: %v", err)
				}
				break
			}
			a.handleClientMessage(r)
		}
	}()
}

type ClientMessage struct {
	To      string          `json:"to"`
	Payload json.RawMessage `json:"payload"`
}

func (a *Agent) handleClientMessage(r io.Reader) {
	var msgReq ClientMessage
	if err := json.NewDecoder(r).Decode(&msgReq); err != nil {
		log.Printf("error decoding message: %v", err)
		return
	}
	a.broker.Send(a.room, messaging.Message{
		From:    a.peer.UID,
		To:      msgReq.To,
		Payload: msgReq.Payload,
	})
}
