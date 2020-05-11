package messaging_test

import (
	"testing"

	"github.com/montrosesoftware/tarpon/pkg/messaging"
)

func TestRegisterNewPeer(t *testing.T) {
	room := messaging.Room{}
	peer := messaging.Peer{UID: "peer-123", Secret: "secret"}

	assertNoPeer(t, room, peer)

	if !room.RegisterPeer(peer) {
		t.Errorf("did not return true when registering peer %+v", peer)
	}

	assertPeer(t, room, peer)
}

func TestUpdateAlreadyRegisteredPeer(t *testing.T) {
	room := messaging.Room{}
	peer := messaging.Peer{UID: "peer-123", Secret: "secret"}
	room.RegisterPeer(peer)

	assertPeer(t, room, peer)

	peer.Secret = "newsecret"
	if room.RegisterPeer(peer) {
		t.Errorf("did not return false when updating peer %+v", peer)
	}

	assertPeer(t, room, peer)
}

func assertNoPeer(t *testing.T, r messaging.Room, p messaging.Peer) {
	t.Helper()
	if _, ok := r.GetPeer(p.UID); ok {
		t.Errorf("peer already exists %+v", p)
	}
	if r.PeersCount() != 0 {
		t.Errorf("room should be empty, but has %d peers", r.PeersCount())
	}
}

func assertPeer(t *testing.T, r messaging.Room, want messaging.Peer) {
	t.Helper()
	p, ok := r.GetPeer(want.UID)
	if !ok {
		t.Errorf("could not find peer with uid %q", want.UID)
	}
	if ok && p != want {
		t.Errorf("got %+v peer, want %+v", p, want)
	}
	if r.PeersCount() != 1 {
		t.Errorf("room should contain 1 peer, but has %d peers", r.PeersCount())
	}
}
