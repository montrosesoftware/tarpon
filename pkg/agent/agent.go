package agent

import (
	"encoding/json"
	"io"
	"log"

	"github.com/gorilla/websocket"
	"github.com/montrosesoftware/tarpon/pkg/broker"
	"github.com/montrosesoftware/tarpon/pkg/messaging"
)

// Agent handles websocket communication between peers and the broker.
type Agent struct {
	peer      messaging.Peer
	room      string
	conn      *websocket.Conn
	broker    broker.Broker
	writeChan chan messaging.Message
}

func New(p messaging.Peer, r string, b broker.Broker) *Agent {
	return &Agent{peer: p, room: r, broker: b, writeChan: make(chan messaging.Message)}
}

func HandlePeer(p messaging.Peer, room string, conn *websocket.Conn) {
	agent := New(p, room, nil) // FIXME: broker is nil
	agent.Start(conn)
}

func (a *Agent) Write(m messaging.Message) {
	a.writeChan <- m
}

func (a *Agent) Start(c *websocket.Conn) {
	a.conn = c
	go a.readPump()
	go a.writePump()
}

func (a *Agent) ID() string {
	return a.peer.UID
}

// readPump handles messages coming from the peer
func (a *Agent) readPump() {
	a.broker.Register(a.room, a)
	defer func() {
		a.broker.Unregister(a.room, a)
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
}

// writePump handles messages coming from the broker
func (a *Agent) writePump() {
	defer func() {
		a.conn.Close()
	}()

	for m := range a.writeChan {
		err := a.conn.WriteJSON(m)
		if err != nil {
			log.Printf("error writing to websocket: %v", err)
			break
		}
	}
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
