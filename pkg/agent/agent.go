package agent

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/gorilla/websocket"
	"github.com/montrosesoftware/tarpon/pkg/broker"
	"github.com/montrosesoftware/tarpon/pkg/logging"
	"github.com/montrosesoftware/tarpon/pkg/messaging"
	"github.com/montrosesoftware/tarpon/pkg/server"
)

const (
	writeWait       = 20 * time.Second
	pongWait        = 60 * time.Second
	pingPeriod      = (pongWait * 9) / 10
	maxMessageSize  = 65536
	messagesBufSize = 16
)

// Agent handles websocket communication between peers and the broker.
type Agent struct {
	peer      messaging.Peer
	room      string
	conn      *websocket.Conn
	broker    broker.Broker
	writeChan chan messaging.Message
	logger    logging.Logger
}

func New(p messaging.Peer, r string, b broker.Broker, l logging.Logger) *Agent {
	return &Agent{peer: p, room: r, broker: b, writeChan: make(chan messaging.Message, messagesBufSize), logger: l}
}

func PeerHandler(b broker.Broker, l logging.Logger) server.PeerHandlerFunc {
	return func(p messaging.Peer, room string, conn *websocket.Conn) {
		agent := New(p, room, b, l)
		agent.Start(conn)
	}
}

func (a *Agent) Write(m messaging.Message) {
	a.writeChan <- m
}

func (a *Agent) logWSError(err error) {
	if websocket.IsUnexpectedCloseError(err,
		websocket.CloseNormalClosure,
		websocket.CloseGoingAway,
		websocket.CloseAbnormalClosure) {
		a.logger.Error("error from websocket", logging.Fields{"room": a.room, "peer": a.peer.UID, "error": err})
	} else {
		a.logger.Info("websocket closed", logging.Fields{"room": a.room, "peer": a.peer.UID, "error": err})
	}
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

	a.conn.SetReadLimit(maxMessageSize)
	if err := a.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		a.logger.Error("error setting read deadline on socket", logging.Fields{"room": a.room, "peer": a.peer.UID, "error": err})
		return
	}
	a.conn.SetPongHandler(func(string) error { return a.conn.SetReadDeadline(time.Now().Add(pongWait)) })

	for {
		_, r, err := a.conn.NextReader()
		if err != nil {
			a.logWSError(err)
			break
		}
		a.handleClientMessage(r)
	}
}

// writePump handles messages coming from the broker
func (a *Agent) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		a.conn.Close()
	}()

	for {
		select {
		case m := <-a.writeChan:
			_ = a.conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := a.conn.WriteJSON(m)
			if err != nil {
				a.logWSError(err)
				return
			}
		case <-ticker.C:
			_ = a.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := a.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				a.logWSError(err)
				return
			}
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
		a.logger.Error("error decoding message:", logging.Fields{"room": a.room, "peer": a.peer.UID, "error": err})
		return
	}
	if msgReq.Payload == nil || bytes.Equal(msgReq.Payload, []byte("null")) {
		a.logger.Info("no payload, dropping message", logging.Fields{"room": a.room, "peer": a.peer.UID})
		return
	}
	a.broker.Send(a.room, messaging.Message{
		From:    a.peer.UID,
		To:      msgReq.To,
		Payload: msgReq.Payload,
	})
}
