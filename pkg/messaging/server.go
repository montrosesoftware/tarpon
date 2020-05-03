package messaging

import (
	"net/http"

	"github.com/gorilla/mux"
)

type RoomStore interface {
	CreateRoom(uid string)
}

type RoomServer struct {
	store RoomStore
}

func NewRoomServer(store RoomStore) *RoomServer {
	return &RoomServer{store}
}

func (s *RoomServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := mux.NewRouter()
	router.HandleFunc("/rooms/{uid:.{1,40}}", s.CreateRoom)
	router.ServeHTTP(w, r)
}

func (s *RoomServer) CreateRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	s.store.CreateRoom(vars["uid"])
}
