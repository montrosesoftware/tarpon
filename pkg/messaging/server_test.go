package messaging_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/montrosesoftware/tarpon/pkg/messaging"
	"github.com/stretchr/testify/assert"
)

type SpyRoomStore struct {
	rooms []string
}

func (s *SpyRoomStore) CreateRoom(uid string) {
	s.rooms = append(s.rooms, uid)
}

func TestCreateRoom(t *testing.T) {
	t.Run("creates empty room with given id", func(t *testing.T) {
		req := messaging.CreateRoomReq{UID: "123-abc"}
		body, _ := json.Marshal(req)
		store := &SpyRoomStore{}
		server := messaging.NewRoomServer(store)
		request, _ := http.NewRequest(http.MethodPost, "/rooms", bytes.NewBuffer(body))
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assert.Equal(t, http.StatusOK, response.Result().StatusCode, "should return success")
		if assert.Equal(t, 1, len(store.rooms), "should create single room") {
			assert.Equal(t, req.UID, store.rooms[0], "room id should match")
		}
	})

	t.Run("returns error when UID too long", func(t *testing.T) {
		req := messaging.CreateRoomReq{UID: "0123456789-0123456789-0123456789-0123456789"}
		body, _ := json.Marshal(req)
		store := &SpyRoomStore{}
		server := messaging.NewRoomServer(store)
		request, _ := http.NewRequest(http.MethodPost, "/rooms", bytes.NewBuffer(body))
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assert.Equal(t, http.StatusBadRequest, response.Result().StatusCode, "should return bad request")
		assert.Equal(t, 0, len(store.rooms), "should not create any rooms")
	})

	t.Run("returns error when invalid request", func(t *testing.T) {
		store := &SpyRoomStore{}
		server := messaging.NewRoomServer(store)
		request, _ := http.NewRequest(http.MethodPost, "/rooms", bytes.NewBuffer([]byte{0}))
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assert.Equal(t, http.StatusBadRequest, response.Result().StatusCode, "should return bad request")
		assert.Equal(t, 0, len(store.rooms), "should not create any rooms")
	})

}
