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

func newMockHandler(broker agent.Broker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("error upgradeing connetion to websocket: %v", err)
			return
		}
		agent := agent.New(messaging.Peer{UID: myPeer}, myRoomUID, c, broker)
		agent.Start()
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
	if len(b.messages) != len(messages) {
		t.Fatalf("got %d messages, but want %d", len(b.messages), len(messages))
	}
	for i := range messages {
		if !reflect.DeepEqual(b.messages[i], messages[i]) {
			t.Errorf("at index %d got message %v, but wanted %v", i, b.messages[i], messages[i])
			break
		}
	}
}

func TestSendMessage(t *testing.T) {
	broker := &SpyBroker{}
	s := httptest.NewServer(newMockHandler(broker))
	defer s.Close()

	wsURL := "ws" + strings.TrimPrefix(s.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("can't open WS connection: %v", err)
	}
	defer ws.Close()

	var messages []messaging.Message
	for i := 0; i < 100; i++ {
		messages = append(messages, generateMessage(i))
	}

	writeIncorrectJSON(t, ws)
	for _, msg := range messages {
		err = ws.WriteJSON(msg)
		if err != nil {
			t.Fatalf("error writing to WS: %v", err)
		}
	}
	writeIncorrectJSON(t, ws)

	// wait until server processes all messages
	time.Sleep(time.Millisecond * 500)

	broker.assertMessages(t, messages)
}

func writeIncorrectJSON(t *testing.T, conn *websocket.Conn) {
	t.Helper()
	if err := conn.WriteMessage(websocket.BinaryMessage, []byte("{")); err != nil {
		t.Fatalf("error writing incorrect JSON to WS: %v", err)
	}
}

func generateMessage(i int) messaging.Message {
	var to string
	if i%2 == 0 {
		to = "another-peer"
	}
	return messaging.Message{
		From:    myPeer, // this is ignored by the agent, but we also use this array to check received messages
		To:      to,
		Payload: json.RawMessage(`"my message"`), // needs to be a valid JSON
	}
}
