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
	writeWait       = 15 * time.Second
	pongWait        = 60 * time.Second
	pingPeriod      = (pongWait * 9) / 10
	maxMessageSize  = 32768
	messagesBufSize = 64
)

// Agent handles websocket communication between peers and the broker.
type Agent struct {
	peer      messaging.Peer
	room      string
	conn      *websocket.Conn
	broker    broker.Broker
	writeChan chan messaging.Message
	stopChan  chan struct{}
	logger    logging.Logger
}

func New(p messaging.Peer, r string, b broker.Broker, l logging.Logger) *Agent {
	return &Agent{peer: p, room: r, broker: b, writeChan: make(chan messaging.Message, messagesBufSize), stopChan: make(chan struct{}), logger: l}
}

func PeerHandler(b broker.Broker, l logging.Logger) server.PeerHandlerFunc {
	return func(p messaging.Peer, room string, conn *websocket.Conn) {
		agent := New(p, room, b, l)
		agent.Start(conn)
	}
}

func (a *Agent) Write(m messaging.Message) {
	a.logMessage("adding message to the write channel...", m)
	select {
	case a.writeChan <- m:
		a.logger.Debug("added message to the write channel", logging.Fields{"room": a.room, "peer": a.peer.UID})
	default:
		a.logger.Warn("message dropped, agent write channel buffer is full", logging.Fields{"room": a.room, "peer": a.peer.UID})
	}
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
	a.logger.Info("agent started", logging.Fields{"room": a.room, "peer": a.peer.UID})
}

func (a *Agent) ID() string {
	return a.peer.UID
}

func (a *Agent) sendControlMessage(msgFactory func(a string) (*messaging.Message, error)) {
	msg, err := msgFactory(a.ID())
	if err != nil {
		a.logger.Error("failed to create control message", logging.Fields{"room": a.room, "peer": a.peer.UID, "error": err})
	} else {
		a.broker.Send(a.room, *msg)
	}
}

// readPump handles messages coming from the peer
func (a *Agent) readPump() {
	a.sendControlMessage(messaging.NewPeerConnected)
	a.broker.Register(a.room, a)

	defer func() {
		a.broker.Unregister(a.room, a)
		a.sendControlMessage(messaging.NewPeerDisconnected)

		close(a.stopChan)
		a.logger.Debug("agent read pump stopped", logging.Fields{"room": a.room, "peer": a.peer.UID})
	}()

	a.conn.SetReadLimit(maxMessageSize)
	if err := a.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		a.logger.Error("error setting read deadline on socket", logging.Fields{"room": a.room, "peer": a.peer.UID, "error": err})
		return
	}
	a.conn.SetPongHandler(func(string) error {
		a.logger.Debug("received pong from peer", logging.Fields{"room": a.room, "peer": a.peer.UID})
		return a.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, r, err := a.conn.NextReader()
		if err != nil {
			a.logWSError(err)
			break
		}
		a.logger.Debug("received data from peer", logging.Fields{"room": a.room, "peer": a.peer.UID})
		a.handleClientMessage(r)
	}
}

// writePump handles messages coming from the broker
func (a *Agent) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		a.logger.Debug("closing websocket", logging.Fields{"room": a.room, "peer": a.peer.UID})
		if err := a.conn.Close(); err != nil {
			a.logger.Warn("error while closing websocket", logging.Fields{"room": a.room, "peer": a.peer.UID, "error": err})
		}
		a.logger.Debug("agent write pump stopped", logging.Fields{"room": a.room, "peer": a.peer.UID})
	}()

	for {
		select {
		case m, more := <-a.writeChan:
			if !more {
				a.logger.Error("agent write channel closed which shouldn't happen", logging.Fields{"room": a.room, "peer": a.peer.UID})
				return
			}

			if err := a.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				a.logger.Error("error setting write deadline for message", logging.Fields{"room": a.room, "peer": a.peer.UID})
			}
			a.logMessage("sending message to peer", m)
			err := a.conn.WriteJSON(m)
			if err != nil {
				a.logWSError(err)
				return
			}
		case <-ticker.C:
			if err := a.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				a.logger.Error("error setting write deadline for ping", logging.Fields{"room": a.room, "peer": a.peer.UID})
			}
			a.logger.Debug("sending ping to peer", logging.Fields{"room": a.room, "peer": a.peer.UID})
			if err := a.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				a.logWSError(err)
				return
			}
		case <-a.stopChan:
			a.logger.Debug("agent write pump received stop signal", logging.Fields{"room": a.room, "peer": a.peer.UID})
			return
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
		a.logger.Debug("no payload, dropping message", logging.Fields{"room": a.room, "peer": a.peer.UID})
		return
	}
	a.logMessage("received message from peer", msgReq)
	a.broker.Send(a.room, messaging.Message{
		From:    a.peer.UID,
		To:      msgReq.To,
		Payload: msgReq.Payload,
	})
}

func (a *Agent) logMessage(t string, o interface{}) {
	if !a.logger.IsDebug() {
		return
	}
	jsonMessage, err := json.Marshal(o)
	if err != nil {
		a.logger.Error("can't marshal object to json", logging.Fields{"room": a.room, "peer": a.peer.UID, "object": o})
		return
	}
	s := string(jsonMessage)
	if len(s) > 1000 {
		s = s[:1000] + "..."
	}
	a.logger.Debug(t, logging.Fields{"room": a.room, "peer": a.peer.UID, "message": s})
}
