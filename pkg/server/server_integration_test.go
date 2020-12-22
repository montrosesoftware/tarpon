package server_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/montrosesoftware/tarpon/pkg/agent"
	"github.com/montrosesoftware/tarpon/pkg/broker"
	"github.com/montrosesoftware/tarpon/pkg/logging"
	"github.com/montrosesoftware/tarpon/pkg/messaging"
	"github.com/montrosesoftware/tarpon/pkg/server"
)

func TestSendingMessagesBetweenPeers(t *testing.T) {
	store := messaging.NewRoomStore()
	broker := broker.NewBroker(logging.NoopLogger{})
	httpServer := httptest.NewServer(server.NewRoomServer(store, agent.PeerHandler(broker, logging.NoopLogger{}), logging.NoopLogger{}))
	defer httpServer.Close()

	room := "aaa3ff11-9ff3-44b8-ab95-b2f339fb9765"
	peer1 := "p1-74cbdcda-bdc3-4fe3-8602-fbaac01689cc"
	peerSecret1 := "4FAAA42E3DEB4C4F0AD20CC9A2A441F400B0A3DD0E57C7FB33EA73D7BFA966BB"
	peer2 := "p2-af868c84-ab5a-4835-8503-93f295068f98"
	peerSecret2 := "88BDA59097E5840A25C2E7B442E88C7790C508F4C759E82047F9637DA6ACB2C5"

	// this normally happens on the backend
	registerPeer(t, httpServer, server.RegisterPeerReq{UID: peer1, Secret: peerSecret1}, room)
	registerPeer(t, httpServer, server.RegisterPeerReq{UID: peer2, Secret: peerSecret2}, room)

	ws1 := peerJoinsRoom(t, httpServer, room, peerSecret1)
	defer ws1.Close()
	ws2 := peerJoinsRoom(t, httpServer, room, peerSecret2)
	defer ws2.Close()

	readMessage(t, ws1) // skip 'peer_connected'

	m1 := agent.ClientMessage{To: peer2, Payload: json.RawMessage(`"ping"`)}
	sendMessage(t, ws1, m1)
	recv1 := readMessage(t, ws2)
	assertSameClientMessages(t, peer1, m1, recv1)

	m2 := agent.ClientMessage{To: peer1, Payload: json.RawMessage(`"pong"`)}
	sendMessage(t, ws2, m2)
	recv2 := readMessage(t, ws1)
	assertSameClientMessages(t, peer2, m2, recv2)
}

func TestSendingControlMessages(t *testing.T) {
	store := messaging.NewRoomStore()
	broker := broker.NewBroker(logging.NoopLogger{})
	httpServer := httptest.NewServer(server.NewRoomServer(store, agent.PeerHandler(broker, logging.NoopLogger{}), logging.NoopLogger{}))
	defer httpServer.Close()

	room := "aaa3ff11-9ff3-44b8-ab95-b2f339fb9765"
	peer1 := "p1-74cbdcda-bdc3-4fe3-8602-fbaac01689cc"
	peerSecret1 := "4FAAA42E3DEB4C4F0AD20CC9A2A441F400B0A3DD0E57C7FB33EA73D7BFA966BB"
	peer2 := "p2-af868c84-ab5a-4835-8503-93f295068f98"
	peerSecret2 := "88BDA59097E5840A25C2E7B442E88C7790C508F4C759E82047F9637DA6ACB2C5"

	// this normally happens on the backend
	registerPeer(t, httpServer, server.RegisterPeerReq{UID: peer1, Secret: peerSecret1}, room)
	registerPeer(t, httpServer, server.RegisterPeerReq{UID: peer2, Secret: peerSecret2}, room)

	ws1 := peerJoinsRoom(t, httpServer, room, peerSecret1)
	defer ws1.Close()
	ws2 := peerJoinsRoom(t, httpServer, room, peerSecret2)

	// ws1 <- ws2 is connected
	c1 := messaging.NewPeerConnected(logging.NoopLogger{}, peer2)
	recv1 := readMessage(t, ws1)
	assertSameMessages(t, c1, recv1)

	ws2.Close()

	// ws1 <- ws2 is disconnected
	c2 := messaging.NewPeerDisconnected(logging.NoopLogger{}, peer2)
	recv2 := readMessage(t, ws1)
	assertSameMessages(t, c2, recv2)
}

func peerJoinsRoom(t *testing.T, s *httptest.Server, room string, secret string) *websocket.Conn {
	ws, _, err := joinRoom(s, room, secret, false)
	if err != nil {
		t.Fatalf("could not open websocket: %v", err)
	}
	return ws
}

func registerPeer(t *testing.T, s *httptest.Server, req server.RegisterPeerReq, room string) {
	t.Helper()
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("could not marshal register peer request body: %v", err)
	}
	body := bytes.NewBuffer(b)
	res, err := s.Client().Post("http://"+s.Listener.Addr().String()+"/rooms/"+room+"/peers", "application/json", body)
	if err != nil {
		t.Fatalf("register peer request failed: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 201 {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("can't read register peer response, status %q: %v", res.Status, err)
		}
		t.Fatalf("register peer returned %q status and %q body, but wanted 201", res.Status, string(bodyBytes))
	}
}

func sendMessage(t *testing.T, ws *websocket.Conn, m agent.ClientMessage) {
	if err := ws.WriteJSON(m); err != nil {
		t.Fatalf("can't send message %v: %v", m, err)
	}
}

func readMessage(t *testing.T, ws *websocket.Conn) messaging.Message {
	var m messaging.Message
	if err := ws.ReadJSON(&m); err != nil {
		t.Fatalf("can't read message %v", err)
	}
	return m
}

func assertSameClientMessages(t *testing.T, from string, sent agent.ClientMessage, recv messaging.Message) {
	if from != recv.From {
		t.Errorf("received message sent from %q, but wanted from %s", recv.From, from)
	}
	if sent.To != recv.To {
		t.Errorf("received message sent to %q, but wanted to %s", recv.To, sent.To)
	}
	if !reflect.DeepEqual(sent.Payload, recv.Payload) {
		t.Errorf("received message body is %v, but wanted %v", recv.Payload, sent.Payload)
	}
}

func assertSameMessages(t *testing.T, sent messaging.Message, recv messaging.Message) {
	if sent.From != recv.From {
		t.Errorf("received message sent from %q, but wanted from %s", recv.From, sent.From)
	}
	if sent.To != recv.To {
		t.Errorf("received message sent to %q, but wanted to %s", recv.To, sent.To)
	}
	if !reflect.DeepEqual(sent.Payload, recv.Payload) {
		t.Errorf("received message body is %v, but wanted %v", recv.Payload, sent.Payload)
	}
}
