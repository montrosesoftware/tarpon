package server_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/montrosesoftware/tarpon/pkg/messaging"
	"github.com/montrosesoftware/tarpon/pkg/server"
)

type StubRoomStore struct {
	server.RoomStore
}

func (s *StubRoomStore) JoinRoom(room string, secret string) (messaging.Peer, error) {
	if room != myRoomUID {
		return messaging.Peer{}, messaging.ErrRoomNotFound
	}

	if secret != mySecret {
		return messaging.Peer{}, messaging.ErrUnauthorized
	}

	return messaging.Peer{UID: myPeer}, nil
}

func TestJoinRoomRequest(t *testing.T) {
	cases := map[string]struct {
		room        string
		secret      string
		wantStatus  int
		wantMessage string
	}{
		"returns error when invalid room": {
			room:        "invalid",
			wantStatus:  404,
			wantMessage: "Room not found\n",
		},
		"returns error when no secret": {
			room:        myRoomUID,
			wantStatus:  401,
			wantMessage: "Unauthorized\n",
		},
		"returns error when bad secret": {
			room:        myRoomUID,
			secret:      "bad",
			wantStatus:  401,
			wantMessage: "Unauthorized\n",
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			store := &StubRoomStore{}
			server := httptest.NewServer(server.NewRoomServer(store))
			defer server.Close()

			wsURL := "ws://" + server.Listener.Addr().String() + "/rooms/" + tt.room + "/ws"

			ws, response, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err == nil {
				defer ws.Close()
			}
			assertResponseStatus(t, response, tt.wantStatus)
			if tt.wantMessage != "" {
				assertResponseMessage(t, response, tt.wantMessage)
			}

			if tt.wantStatus != 101 {
				return
			}

			if err != nil {
				t.Fatalf("could not open a ws connection on %s: %v", wsURL, err)
			}
		})
	}
}

func assertResponseStatus(t *testing.T, got *http.Response, want int) {
	t.Helper()
	if got.StatusCode != want {
		t.Errorf("did not get correct status, got %d, want %d", got.StatusCode, want)
	}
}

func assertResponseMessage(t *testing.T, got *http.Response, want string) {
	t.Helper()
	body := getResponseBody(t, got)
	if body != want {
		t.Errorf("did not return correct message body, got %q, want %q", body, want)
	}
}

func getResponseBody(t *testing.T, r *http.Response) string {
	t.Helper()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Errorf("could not read response body: %v", err)
		return ""
	}
	return string(bodyBytes)
}
