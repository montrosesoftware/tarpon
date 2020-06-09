package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/montrosesoftware/tarpon/pkg/messaging"
	"github.com/montrosesoftware/tarpon/pkg/msv"
)

type RoomStore interface {
	CreateRoom(uid string) bool
	RegisterPeer(room string, peer messaging.Peer) bool
	JoinRoom(room string, secret string) (messaging.Peer, error)
}

type PeerHandlerFunc func(p messaging.Peer, room string, conn *websocket.Conn)

type RoomServer struct {
	store       RoomStore
	peerHandler PeerHandlerFunc
}

func NewRoomServer(store RoomStore, ph PeerHandlerFunc) *RoomServer {
	return &RoomServer{store, ph}
}

func (s *RoomServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	head, tail := msv.ShiftPath(r.URL.Path)

	if head == "rooms" {
		head, tail := msv.ShiftPath(tail)
		if head == "" {
			if checkMethod(w, r, http.MethodPost) {
				s.CreateRoom(w, r)
			}
			return
		}
		{
			head, _ := msv.ShiftPath(tail)
			if head == "ws" {
				if checkMethod(w, r, http.MethodGet) {
					s.JoinRoom(w, r)
				}
				return
			}
			if head == "peers" {
				if checkMethod(w, r, http.MethodPost) {
					s.RegisterPeer(w, r)
				}
				return
			}
		}
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

	p := messaging.Peer(req)
	if s.store.RegisterPeer(room, p) {
		w.WriteHeader(http.StatusCreated)
		logger(w.Write([]byte("Created\n")))
	} else {
		w.WriteHeader(http.StatusOK)
		logger(w.Write([]byte("OK\n")))
	}
}

func (s *RoomServer) JoinRoom(w http.ResponseWriter, r *http.Request) {
	room, tail := msv.ShiftPathN(r.URL.Path, 2)

	if tail != "/ws" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	secret := getSecret(r)
	peer, err := s.store.JoinRoom(room, secret)

	if err != nil {
		switch err {
		case messaging.ErrRoomNotFound:
			http.Error(w, "Room not found", http.StatusNotFound)
		case messaging.ErrUnauthorized:
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		default:
			log.Printf("unknown error when joining room: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		Subprotocols: []string{"tarpon"},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("couldn't upgrade to websocket: %v", err)
		return
	}

	s.peerHandler(peer, room, conn)
}

func logger(n int, err error) {
	if err != nil {
		log.Printf("response write failed: %v", err)
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

func getSecret(r *http.Request) string {
	h := r.Header.Get("Authorization")

	if h == "" {
		return getSecretFromSubprotocols(r)
	}

	return strings.TrimSpace(strings.Replace(h, "Bearer", "", 1))
}

func getSecretFromSubprotocols(r *http.Request) string {
	subprotocols := websocket.Subprotocols(r)
	for i, s := range subprotocols {
		if s == "access_token" {
			if i == len(subprotocols)-1 {
				return ""
			}
			return subprotocols[i+1]
		}
	}
	return ""
}
