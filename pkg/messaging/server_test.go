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

type SpyRoomStore struct {
	rooms []string
}

func (s *SpyRoomStore) CreateRoom(uid string) bool {
	if uid == "duplicate" {
		return false
	}

	s.rooms = append(s.rooms, uid)
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
			wantMessage: "Created",
		},
		"returns error when UID too long": {
			uid:        "0123456789-0123456789-0123456789-0123456789",
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

func assertStatus(t *testing.T, got *httptest.ResponseRecorder, want int) {
	t.Helper()
	if got.Code != want {
		t.Errorf("did not get correct status, got %d, want %d", got.Code, want)
	}
}

func assertRoomCreated(t *testing.T, got *SpyRoomStore, wantRoom bool, uid string) {
	t.Helper()
	if wantRoom {
		if len(got.rooms) == 1 {
			if got.rooms[0] != uid {
				t.Errorf("did not create right room, got %q, want %q", got.rooms[0], uid)
			}
		} else {
			t.Errorf("did not create correct number of rooms, got %d, want 1", len(got.rooms))
		}
	} else {
		if len(got.rooms) != 0 {
			t.Errorf("did create rooms when it should not, got %d, want 0", len(got.rooms))
		}
	}
}

func assertMessage(t *testing.T, got *httptest.ResponseRecorder, want string) {
	t.Helper()
	if got.Body.String() != want {
		t.Errorf("did not return correct message body, got %q, want %q", got.Body.String(), want)
	}
}
