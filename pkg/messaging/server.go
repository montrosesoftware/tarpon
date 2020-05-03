package messaging

import "net/http"

type RoomStore interface {
	CreateRoom(uid string)
}

type RoomServer struct {
	Store RoomStore
}

func (s *RoomServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Store.CreateRoom("123")
}
