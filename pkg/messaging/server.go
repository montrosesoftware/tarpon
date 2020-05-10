package messaging

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/montrosesoftware/tarpon/pkg/msv"
)

type RoomStore interface {
	CreateRoom(uid string) bool
	RegisterPeer(room string, peer Peer) bool
}

type RoomServer struct {
	store RoomStore
}

func NewRoomServer(store RoomStore) *RoomServer {
	return &RoomServer{store}
}

func (s *RoomServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	head, tail := msv.ShiftPath(r.URL.Path)

	if head == "rooms" {
		head, _ := msv.ShiftPath(tail)
		if head == "" {
			if checkMethod(w, r, http.MethodPost) {
				s.CreateRoom(w, r)
			}
		} else {
			if checkMethod(w, r, http.MethodPost) {
				s.RegisterPeer(w, r)
			}
		}
		return
	}

	http.Error(w, "Not Found", http.StatusNotFound)
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

	if !checkLength(w, req.UID, 1, 40, "uid") {
		return
	}

	created := s.store.CreateRoom(req.UID)
	if created {
		w.WriteHeader(http.StatusCreated)
		logger(w.Write([]byte("Created\n")))
	} else {
		http.Error(w, "uid: already exists", http.StatusConflict)
	}
}

type RegisterPeerReq struct {
	UID    string `json:"uid"`
	Secret string `json:"secret"`
}

func (s *RoomServer) RegisterPeer(w http.ResponseWriter, r *http.Request) {
	room, tail := msv.ShiftPathN(r.URL.Path, 2)

	if !checkLength(w, room, 1, 40, "room uid") {
		return
	}

	if tail != "/peers" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	var req RegisterPeerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "decoding json failed", http.StatusBadRequest)
		return
	}

	if !checkLength(w, req.UID, 1, 40, "uid") {
		return
	}

	if !checkLength(w, req.Secret, 24, 100, "secret") {
		return
	}

	p := Peer(req)
	if s.store.RegisterPeer(room, p) {
		w.WriteHeader(http.StatusCreated)
		logger(w.Write([]byte("Created\n")))
	} else {
		w.WriteHeader(http.StatusOK)
		logger(w.Write([]byte("OK\n")))
	}
}

func logger(n int, err error) {
	if err != nil {
		log.Printf("Response write failed: %v", err)
	}
}

func checkLength(w http.ResponseWriter, val string, lower int, upper int, name string) bool {
	if len(val) < lower || len(val) > upper {
		http.Error(w, fmt.Sprint(name, ": must be between ", lower, " and ", upper, " characters"), http.StatusBadRequest)
		return false
	}
	return true
}

func checkMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}
