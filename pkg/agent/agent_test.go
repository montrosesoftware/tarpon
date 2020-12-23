package agent_test

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/montrosesoftware/tarpon/pkg/agent"
	"github.com/montrosesoftware/tarpon/pkg/broker"
	"github.com/montrosesoftware/tarpon/pkg/logging"
	"github.com/montrosesoftware/tarpon/pkg/messaging"
)

var (
	myRoomUID = "room-123"
	myPeer    = "peer-abc"
)

func newMockHandler(agent *agent.Agent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("error upgrading connetion to websocket: %v", err)
			return
		}
		agent.Start(c)
	}
}

type SpyBroker struct {
	messages    []messaging.Message
	subscribers []broker.Subscriber
	mutex       sync.Mutex
}

func (b *SpyBroker) Send(room string, m messaging.Message) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if room == myRoomUID {
		b.messages = append(b.messages, m)
	}
}

func (b *SpyBroker) Register(room string, s broker.Subscriber) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if room == myRoomUID {
		b.subscribers = append(b.subscribers, s)
	}
}

func (b *SpyBroker) Unregister(room string, s broker.Subscriber) bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if room == myRoomUID && b.subscribers[0] == s {
		b.subscribers = b.subscribers[1:]
		return true
	}
	return false
}

func (b *SpyBroker) assertMessages(t *testing.T, messages []messaging.Message) {
	t.Helper()
	b.mutex.Lock()
	defer b.mutex.Unlock()
	assertSameMessages(t, b.messages, messages)
}

func (b *SpyBroker) assertSubscriber(t *testing.T, id string) {
	t.Helper()
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if len(b.subscribers) != 1 || b.subscribers[0].ID() != id {
		t.Errorf("got %v subscribers, but wanted one with id %q", b.subscribers, id)
	}
}

func (b *SpyBroker) assertNoSubscriber(t *testing.T) {
	t.Helper()
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if len(b.subscribers) != 0 {
		t.Errorf("got %v subscribers, but wanted none", b.subscribers)
	}
}

func TestSubsciptionToBroker(t *testing.T) {
	broker := &SpyBroker{}
	agent := agent.New(messaging.Peer{UID: myPeer}, myRoomUID, broker, logging.NoopLogger{})
	s := httptest.NewServer(newMockHandler(agent))
	defer s.Close()

	broker.assertNoSubscriber(t)
	ws := openWS(t, s)
	defer ws.Close()
	// wait for the server to register the agent
	time.Sleep(time.Millisecond * 100)
	broker.assertSubscriber(t, myPeer)
	if err := ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(1000, "closing")); err != nil {
		t.Errorf("error when closing websocket: %v", err)
	}
	// wait until server cleans up
	time.Sleep(time.Millisecond * 100)
	broker.assertNoSubscriber(t)
}

func TestSendMessageToBroker(t *testing.T) {
	broker := &SpyBroker{}
	agent := agent.New(messaging.Peer{UID: myPeer}, myRoomUID, broker, logging.NoopLogger{})
	s := httptest.NewServer(newMockHandler(agent))
	defer s.Close()

	ws := openWS(t, s)
	defer ws.Close()

	var ctrlMessages []messaging.Message
	msg, err := messaging.NewPeerConnected(myPeer)
	if err != nil {
		t.Fatalf("error creating control message: %v", err)
	}
	ctrlMessages = append(ctrlMessages, *msg)

	var messages []messaging.Message
	for i := 0; i < 100; i++ {
		messages = append(messages, generateMessage(i))
	}

	writeMessageWithoutPayload(t, ws)
	writeIncorrectJSON(t, ws)
	for _, msg := range messages {
		err := ws.WriteJSON(msg)
		if err != nil {
			t.Fatalf("error writing to WS: %v", err)
		}
	}
	writeIncorrectJSON(t, ws)

	// wait until server processes all messages
	time.Sleep(time.Millisecond * 100)

	broker.assertMessages(t, append(ctrlMessages, messages...))
}

func TestWriteControlMessages(t *testing.T) {
	broker := &SpyBroker{}
	agent := agent.New(messaging.Peer{UID: myPeer}, myRoomUID, broker, logging.NoopLogger{})
	s := httptest.NewServer(newMockHandler(agent))
	defer s.Close()

	ws := openWS(t, s)
	ws.Close()

	var messages []messaging.Message

	msg1, err1 := messaging.NewPeerConnected(myPeer)
	if err1 != nil {
		t.Fatalf("error creating control message: %v", err1)
	}
	messages = append(messages, *msg1)

	msg2, err2 := messaging.NewPeerDisconnected(myPeer)
	if err2 != nil {
		t.Fatalf("error creating control message: %v", err2)
	}
	messages = append(messages, *msg2)

	// wait for a server to process peers connection
	time.Sleep(time.Millisecond * 100)

	broker.assertMessages(t, messages)
}

func TestWriteMessageToPeer(t *testing.T) {
	broker := &SpyBroker{}
	agent := agent.New(messaging.Peer{UID: myPeer}, myRoomUID, broker, logging.NoopLogger{})
	s := httptest.NewServer(newMockHandler(agent))
	defer s.Close()

	ws := openWS(t, s)
	defer ws.Close()

	var messages []messaging.Message
	for i := 0; i < 100; i++ {
		messages = append(messages, generateMessage(i))
	}

	for _, msg := range messages {
		agent.Write(msg)
	}

	var receivedMessages []messaging.Message
	for i := 0; i < 100; i++ {
		_ = ws.SetReadDeadline(time.Now().Add(time.Second * 1))
		var msg messaging.Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			t.Fatalf("error reading from WS: %v", err)
		}
		receivedMessages = append(receivedMessages, msg)
	}

	assertSameMessages(t, receivedMessages, messages)

	// check if writing to closed WS is handled
	if err := ws.Close(); err != nil {
		t.Errorf("error closing WS: %v", err)
	}
	agent.Write(generateMessage(0))
}

func assertSameMessages(t *testing.T, got []messaging.Message, want []messaging.Message) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d messages, but want %d", len(got), len(want))
	}
	for i := range want {
		if !reflect.DeepEqual(got[i], want[i]) {
			t.Errorf("at index %d got message %v, but wanted %v", i, got[i], want[i])
			break
		}
	}
}

func openWS(t *testing.T, s *httptest.Server) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(s.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("can't open WS connection: %v", err)
	}
	return ws
}

func writeIncorrectJSON(t *testing.T, conn *websocket.Conn) {
	t.Helper()
	if err := conn.WriteMessage(websocket.BinaryMessage, []byte("{")); err != nil {
		t.Errorf("error writing incorrect JSON to WS: %v", err)
	}
}

func writeMessageWithoutPayload(t *testing.T, conn *websocket.Conn) {
	t.Helper()
	if err := conn.WriteJSON(messaging.Message{To: "another-peer"}); err != nil {
		t.Errorf("error writing empty payload message to WS: %v", err)
	}
}

func generateMessage(i int) messaging.Message {
	var to string
	if i%2 == 0 {
		to = "another-peer"
	}
	return messaging.Message{
		From:    myPeer, // this is ignored by the agent when sent as a client message, but we use it for assertions
		To:      to,
		Payload: json.RawMessage(`"my message"`), // needs to be a valid JSON
	}
}
