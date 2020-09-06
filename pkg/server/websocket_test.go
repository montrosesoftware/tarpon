package server_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
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

type SpyPeerHandler struct {
	handled []struct {
		peer messaging.Peer
		room string
	}
	mutex sync.Mutex
}

func (s *SpyPeerHandler) handlePeer(p messaging.Peer, room string, conn *websocket.Conn) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.handled = append(s.handled, struct {
		peer messaging.Peer
		room string
	}{p, room})
}

func (s *SpyPeerHandler) assertPeerHandled(t *testing.T, peer messaging.Peer, room string) {
	t.Helper()
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if len(s.handled) == 1 {
		if s.handled[0].peer != peer {
			t.Errorf("did not handle right peer, got %v, want %v", s.handled[0].peer, peer)
		}
		if s.handled[0].room != room {
			t.Errorf("did not join the right room, got %q, want %q", s.handled[0].room, room)
		}
	} else {
		t.Errorf("didn't handle 1 peer, got %d", len(s.handled))
	}
}

func TestJoinRoomRequest(t *testing.T) {
	cases := map[string]struct {
		room           string
		secret         string
		useSubprotocol bool
		wantStatus     int
		wantMessage    string
	}{
		"returns error when invalid room": {
			room:        "invalid",
			wantStatus:  404,
			wantMessage: "Room not found\n",
		},
		"returns error when no room provided": {
			room:        "",
			wantStatus:  400,
			wantMessage: "Room not provided\n",
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
		"upgrades to websocket": {
			room:       myRoomUID,
			secret:     mySecret,
			wantStatus: 101,
		},
		"upgrades to websocket when secret provided as subprotocol": {
			room:           myRoomUID,
			secret:         mySecret,
			useSubprotocol: true,
			wantStatus:     101,
		},
		"return error when empty secret provided as subprotocol": {
			room:           myRoomUID,
			secret:         "",
			useSubprotocol: true,
			wantStatus:     401,
			wantMessage:    "Unauthorized\n",
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			store := &StubRoomStore{}
			ph := &SpyPeerHandler{}
			server := httptest.NewServer(server.NewRoomServer(store, ph.handlePeer))
			defer server.Close()

			ws, response, err := joinRoom(server, tt.room, tt.secret, tt.useSubprotocol)
			if err == nil {
				defer ws.Close()
			}
			assertResponseStatus(t, response, tt.wantStatus)
			if tt.wantMessage != "" {
				assertResponseMessage(t, response, tt.wantMessage)
			}

			// End the test if it isn't expected to open a websocket
			if tt.wantStatus != 101 {
				return
			}

			ph.assertPeerHandled(t, messaging.Peer{UID: myPeer}, tt.room)

			if err != nil {
				t.Fatalf("could not open a ws connection: %v", err)
			}
		})
	}
}

func joinRoom(server *httptest.Server, room string, secret string, useSubprotocol bool) (*websocket.Conn, *http.Response, error) {
	wsURL := "ws://" + server.Listener.Addr().String() + "/rooms/ws?room=" + room

	var header http.Header
	if secret != "" {
		if !useSubprotocol {
			header = http.Header{"Authorization": {"Bearer " + secret}}
		} else {
			header = http.Header{"Sec-WebSocket-Protocol": {"access_token," + secret}}
		}
	}

	return websocket.DefaultDialer.Dial(wsURL, header)
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
