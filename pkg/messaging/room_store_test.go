package messaging_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/montrosesoftware/tarpon/pkg/messaging"
)

var myRoom = "room-123"

func TestCreateEmptyRoom(t *testing.T) {
	store := messaging.NewRoomStore()

	assertNoRoom(t, store, myRoom)

	if !store.CreateRoom(myRoom) {
		t.Errorf("did not return true when creating room %q", myRoom)
	}

	assertRoom(t, store, myRoom, &messaging.Room{})
}

func TestRejectRecreatingRoom(t *testing.T) {
	store := messaging.NewRoomStore()
	room := createRoomWithPeer(t, store)

	assertRoom(t, store, myRoom, room)

	if store.CreateRoom(myRoom) {
		t.Error("recreated room when it should not")
	}

	assertRoom(t, store, myRoom, room)
}

func TestRegisterPeerInRoom(t *testing.T) {
	store := messaging.NewRoomStore()
	room := createEmptyRoom(t, store)
	p := messaging.Peer{}

	if !store.RegisterPeer(myRoom, p) {
		t.Errorf("did not return true when registering peer %+v", p)
	}

	assertPeer(t, room, p)
}

func TestCreateRoomWhenRegisteringPeer(t *testing.T) {
	store := messaging.NewRoomStore()
	p := messaging.Peer{}

	assertNoRoom(t, store, myRoom)

	if !store.RegisterPeer(myRoom, p) {
		t.Errorf("did not return true when registering peer %+v", p)
	}

	r := store.GetRoom(myRoom)
	if r == nil {
		t.Fatalf("no room created")
	}
	assertPeer(t, r, p)
}

func TestRegisterPeerConcurrently(t *testing.T) {
	store := messaging.NewRoomStore()
	for i := 0; i < 2; i++ {
		go func(uid string) {
			p := messaging.Peer{UID: uid}
			store.RegisterPeer(myRoom, p)
			r := store.RoomsCount()
			if r != 1 {
				t.Errorf("got %d rooms, but want 1", r)
			}
		}(strconv.Itoa(i))
	}
}

func assertNoRoom(t *testing.T, s *messaging.MemoryRoomStore, uid string) {
	t.Helper()
	r := s.GetRoom(uid)
	if r != nil {
		t.Errorf("found room %+v, want none", r)
	}
	if s.RoomsCount() != 0 {
		t.Errorf("store should be empty, but has %d rooms", s.RoomsCount())
	}
}

func assertRoom(t *testing.T, s *messaging.MemoryRoomStore, uid string, want *messaging.Room) {
	t.Helper()
	r := s.GetRoom(uid)
	if r == nil {
		t.Errorf("could not find room with uid %q", uid)
		return
	}
	if !reflect.DeepEqual(r, want) {
		t.Errorf("got room %+v, want %+v", r, want)
	}
	if s.RoomsCount() != 1 {
		t.Errorf("store should contain 1 room, but has %d rooms", s.RoomsCount())
	}
}

func createEmptyRoom(t *testing.T, s *messaging.MemoryRoomStore) *messaging.Room {
	t.Helper()
	s.CreateRoom(myRoom)
	r := s.GetRoom(myRoom)
	if r == nil {
		t.Fatalf("could not create empty room")
	}
	return r
}

func createRoomWithPeer(t *testing.T, s *messaging.MemoryRoomStore) *messaging.Room {
	t.Helper()
	r := createEmptyRoom(t, s)
	if !r.RegisterPeer(messaging.Peer{}) {
		t.Fatalf("could not register peer in room")
	}
	return r
}
