package messaging

import (
	"errors"
	"sync"
)

var (
	ErrRoomNotFound = errors.New("room not found")
	ErrUnauthorized = errors.New("not authorized")
)

type MemoryRoomStore struct {
	rooms map[string]*Room
	mutex sync.RWMutex
}

func NewRoomStore() *MemoryRoomStore {
	s := MemoryRoomStore{}
	s.rooms = make(map[string]*Room)
	return &s
}

func (s *MemoryRoomStore) CreateRoom(uid string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.ensureRoom(uid)
}

func (s *MemoryRoomStore) GetRoom(uid string) *Room {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.rooms[uid]
}

func (s *MemoryRoomStore) RoomsCount() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return len(s.rooms)
}

func (s *MemoryRoomStore) RegisterPeer(room string, p Peer) bool {
	s.mutex.Lock()
	s.ensureRoom(room)
	r := s.rooms[room]
	s.mutex.Unlock()
	return r.RegisterPeer(p)
}

func (s *MemoryRoomStore) JoinRoom(room string, secret string) (Peer, error) {
	panic("Not implemented")
}

// ensureRoom assumes a lock is held
func (s *MemoryRoomStore) ensureRoom(uid string) bool {
	if _, ok := s.rooms[uid]; ok {
		return false
	}
	s.rooms[uid] = &Room{}
	return true
}
