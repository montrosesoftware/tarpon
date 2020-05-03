package messaging_test

import (
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
		room := "123"
		store := SpyRoomStore{}
		server := messaging.RoomServer{&store}
		request, _ := http.NewRequest(http.MethodPost, "/rooms/"+room, nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assert.Equal(t, http.StatusOK, response.Result().StatusCode, "should return status success")
		assert.Equal(t, 1, len(store.rooms), "should create single room")
		assert.Equal(t, room, store.rooms[0], "room id should match")
	})

}
