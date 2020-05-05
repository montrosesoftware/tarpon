package messaging

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type RoomStore interface {
	CreateRoom(uid string) bool
}

type RoomServer struct {
	store RoomStore
}

func NewRoomServer(store RoomStore) *RoomServer {
	return &RoomServer{store}
}

func (s *RoomServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router := mux.NewRouter()
	router.HandleFunc("/rooms", s.CreateRoom).Methods("POST")
	router.ServeHTTP(w, r)
}

type CreateRoomReq struct {
	UID string `json:"uid"`
}

func (s *RoomServer) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req CreateRoomReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "decoding json failed", http.StatusBadRequest)
		return
	}

	if len(req.UID) < 1 || len(req.UID) > 40 {
		http.Error(w, "uid: must be between 1 and 40 characters", http.StatusBadRequest)
		return
	}

	created := s.store.CreateRoom(req.UID)
	if created {
		w.WriteHeader(http.StatusCreated)
		logger(w.Write([]byte("Created")))
	} else {
		http.Error(w, "uid: already exists", http.StatusConflict)
	}
}

func logger(n int, err error) {
	if err != nil {
		log.Printf("Response write failed: %v", err)
	}
}
