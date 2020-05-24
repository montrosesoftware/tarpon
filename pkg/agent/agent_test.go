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
			log.Printf("error upgradeing connetion to websocket: %v", err)
			return
		}
		agent.Start(c)
	}
}

type SpyBroker struct {
	messages []messaging.Message
	mutex    sync.Mutex
}

func (b *SpyBroker) Send(room string, m messaging.Message) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if room == myRoomUID {
		b.messages = append(b.messages, m)
	}
}

func (b *SpyBroker) assertMessages(t *testing.T, messages []messaging.Message) {
	t.Helper()
	b.mutex.Lock()
	defer b.mutex.Unlock()
	assertSameMessages(t, b.messages, messages)
}

func TestSendMessageToBroker(t *testing.T) {
	broker := &SpyBroker{}
	agent := agent.New(messaging.Peer{UID: myPeer}, myRoomUID, broker)
	s := httptest.NewServer(newMockHandler(agent))
	defer s.Close()

	ws := openWS(t, s)
	defer ws.Close()

	var messages []messaging.Message
	for i := 0; i < 100; i++ {
		messages = append(messages, generateMessage(i))
	}

	writeIncorrectJSON(t, ws)
	for _, msg := range messages {
		err := ws.WriteJSON(msg)
		if err != nil {
			t.Fatalf("error writing to WS: %v", err)
		}
	}
	writeIncorrectJSON(t, ws)

	// wait until server processes all messages
	time.Sleep(time.Millisecond * 500)

	broker.assertMessages(t, messages)
}

func TestWriteMessageToPeer(t *testing.T) {
	agent := agent.New(messaging.Peer{UID: myPeer}, myRoomUID, nil)
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
