package messaging_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/montrosesoftware/tarpon/pkg/messaging"
)

const (
	tooLongUID = "0123456789-0123456789-0123456789-0123456789"
)

type SpyRoomStore struct {
	rooms []string
	peers []messaging.Peer
	t     *testing.T
}

func (s *SpyRoomStore) CreateRoom(uid string) bool {
	if uid == "duplicate" {
		return false
	}

	s.rooms = append(s.rooms, uid)
	return true
}

func (s *SpyRoomStore) RegisterPeer(room string, peer messaging.Peer) bool {
	if room != "room-123" {
		s.t.Errorf("unexpected room %q passed to register peer", room)
		return false
	}
	if peer.UID == "duplicate" {
		return false
	}
	s.peers = append(s.peers, peer)
	return true
}

func TestCreateRoom(t *testing.T) {
	cases := map[string]struct {
		uid         string
		wantStatus  int
		wantRoom    bool
		wantMessage string
	}{
		"creates empty room with given id": {
			uid:         "123-abc",
			wantStatus:  201,
			wantRoom:    true,
			wantMessage: "Created\n",
		},
		"returns error when UID too long": {
			uid:        tooLongUID,
			wantStatus: 400,
			wantRoom:   false,
		}, "returns error when invalid request": {
			uid:        "",
			wantStatus: 400,
			wantRoom:   false,
		}, "returns error when already exists": {
			uid:        "duplicate",
			wantStatus: 409,
			wantRoom:   false,
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			store := &SpyRoomStore{}
			server := messaging.NewRoomServer(store)

			request := newCreateRoomRequest(t, tt.uid)
			response := httptest.NewRecorder()

			server.ServeHTTP(response, request)

			assertStatus(t, response, tt.wantStatus)
			assertRoomCreated(t, store, tt.wantRoom, tt.uid)
			if tt.wantMessage != "" {
				assertMessage(t, response, tt.wantMessage)
			}
		})
	}
}
func TestRegisterPeer(t *testing.T) {
	cases := map[string]struct {
		peer        *messaging.Peer
		wantStatus  int
		wantPeer    bool
		wantMessage string
	}{
		"creates given peer": {
			peer:        &messaging.Peer{UID: "peer-abc", Secret: "secret"},
			wantStatus:  201,
			wantPeer:    true,
			wantMessage: "Created\n",
		},
		"returns error when peer UID too long": {
			peer:       &messaging.Peer{UID: tooLongUID, Secret: "secret"},
			wantStatus: 400,
			wantPeer:   false,
		}, "returns error when invalid request": {
			peer:       nil,
			wantStatus: 400,
			wantPeer:   false,
		}, "returns 200 when already registered": {
			peer:        &messaging.Peer{UID: "duplicate", Secret: "secret"},
			wantStatus:  200,
			wantPeer:    false,
			wantMessage: "OK\n",
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			store := &SpyRoomStore{t: t}
			server := messaging.NewRoomServer(store)

			request := newRegisterPeerRequest(t, tt.peer)
			response := httptest.NewRecorder()

			server.ServeHTTP(response, request)

			assertStatus(t, response, tt.wantStatus)
			assertPeerRegistered(t, store, tt.wantPeer, tt.peer)
			if tt.wantMessage != "" {
				assertMessage(t, response, tt.wantMessage)
			}
		})
	}
}

func TestInvalidRequests(t *testing.T) {
	cases := []struct {
		url    string
		method string
		status int
	}{
		{"/", "GET", 404},
		{"/abc", "GET", 404},
		{"/rooms", "GET", 405},
		{"/rooms/abc/test", "POST", 404},
		{"/rooms/abc/peers", "GET", 405},
		{"/rooms/abc/peers/abc", "POST", 404},
	}
	for _, tt := range cases {
		t.Run("check "+tt.url, func(t *testing.T) {
			store := &SpyRoomStore{}
			server := messaging.NewRoomServer(store)
			req, err := http.NewRequest(tt.method, tt.url, nil)
			if err != nil {
				t.Fatalf("could not instantiate request: %v", err)
			}
			response := httptest.NewRecorder()
			server.ServeHTTP(response, req)
			assertStatus(t, response, tt.status)
			if tt.status == 404 {
				assertMessage(t, response, "Not Found\n")
			} else {
				assertMessage(t, response, "Method Not Allowed\n")
			}
		})
	}
}

// Creates invalid body request when uid is empty
func newCreateRoomRequest(t *testing.T, uid string) *http.Request {
	t.Helper()
	var body io.Reader
	if uid == "" {
		body = bytes.NewBuffer([]byte{0})
	} else {
		req := messaging.CreateRoomReq{UID: uid}
		b, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("could not marshal create room request body: %v", err)
		}
		body = bytes.NewBuffer(b)
	}
	req, err := http.NewRequest("POST", "/rooms", body)
	if err != nil {
		t.Fatalf("could not instantiate create room request: %v", err)
	}
	return req
}

func newRegisterPeerRequest(t *testing.T, peer *messaging.Peer) *http.Request {
	t.Helper()
	var body io.Reader
	if peer == nil {
		body = bytes.NewBuffer([]byte{0})
	} else {
		req := messaging.RegisterPeerReq(*peer)
		b, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("could not marshal register peer request body: %v", err)
		}
		body = bytes.NewBuffer(b)
	}
	req, err := http.NewRequest("POST", "/rooms/room-123/peers", body)
	if err != nil {
		t.Fatalf("could not instantiate register peer request: %v", err)
	}
	return req
}

func assertStatus(t *testing.T, got *httptest.ResponseRecorder, want int) {
	t.Helper()
	if got.Code != want {
		t.Errorf("did not get correct status, got %d, want %d", got.Code, want)
	}
}

func assertRoomCreated(t *testing.T, got *SpyRoomStore, wantRoom bool, uid string) {
	t.Helper()
	if !wantRoom {
		if len(got.rooms) != 0 {
			t.Errorf("did create rooms when it should not, got %d, want 0", len(got.rooms))
		}
		return
	}

	if len(got.rooms) == 1 {
		if got.rooms[0] != uid {
			t.Errorf("did not create right room, got %q, want %q", got.rooms[0], uid)
		}
	} else {
		t.Errorf("did not create correct number of rooms, got %d, want 1", len(got.rooms))
	}
}

func assertPeerRegistered(t *testing.T, got *SpyRoomStore, wantPeer bool, peer *messaging.Peer) {
	t.Helper()
	if !wantPeer {
		if len(got.peers) != 0 {
			t.Errorf("did register peer when it should not, got %d, want 0", len(got.peers))
		}
		return
	}

	if len(got.peers) == 1 {
		if got.peers[0] != *peer {
			t.Errorf("did not register right peer, got %+v, want %+v", got.peers[0], *peer)
		}
	} else {
		t.Errorf("did not register correct number of peers, got %d, want 1", len(got.peers))
	}
}

func assertMessage(t *testing.T, got *httptest.ResponseRecorder, want string) {
	t.Helper()
	if got.Body.String() != want {
		t.Errorf("did not return correct message body, got %q, want %q", got.Body.String(), want)
	}
}
