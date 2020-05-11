package messaging

type MemoryRoomStore struct {
	rooms map[string]*Room
}

func NewRoomStore() *MemoryRoomStore {
	s := MemoryRoomStore{}
	s.rooms = make(map[string]*Room)
	return &s
}

func (s *MemoryRoomStore) CreateRoom(uid string) bool {
	if _, ok := s.rooms[uid]; ok {
		return false
	}
	s.rooms[uid] = &Room{}
	return true
}

func (s *MemoryRoomStore) GetRoom(uid string) *Room {
	return s.rooms[uid]
}

func (s *MemoryRoomStore) RoomsCount() int {
	return len(s.rooms)
}

func (s *MemoryRoomStore) RegisterPeer(room string, p Peer) bool {
	s.CreateRoom(room)
	return s.rooms[room].RegisterPeer(p)
}
