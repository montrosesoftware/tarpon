package messaging_test

import (
	"testing"

	"github.com/montrosesoftware/tarpon/pkg/messaging"
)

func TestRegisterNewPeer(t *testing.T) {
	room := messaging.Room{}
	peer := messaging.Peer{UID: "peer-123", Secret: "secret"}

	if _, ok := room.GetPeer(peer.UID); ok {
		t.Errorf("peer already exists %+v", peer)
	}
	if room.PeersCount() != 0 {
		t.Errorf("room should be empty, but has %d peers", room.PeersCount())
	}

	if !room.RegisterPeer(peer) {
		t.Errorf("did not return true when registering peer %+v", peer)
	}

	p, ok := room.GetPeer(peer.UID)
	if !ok {
		t.Errorf("could not find peer with uid %q", peer.UID)
	}
	if ok && p != peer {
		t.Errorf("got %+v peer, want %+v", p, peer)
	}

	if room.PeersCount() != 1 {
		t.Errorf("room should contain 1 peer, but has %d peers", room.PeersCount())
	}
}

func TestUpdateAlreadyRegisteredPeer(t *testing.T) {
	room := messaging.Room{}
	peer := messaging.Peer{UID: "peer-123", Secret: "secret"}
	room.RegisterPeer(peer)

	if room.PeersCount() != 1 {
		t.Errorf("room should contain 1 peer, but has %d peers", room.PeersCount())
	}

	peer.Secret = "newsecret"
	if room.RegisterPeer(peer) {
		t.Errorf("did not return false when updating peer %+v", peer)
	}

	if room.PeersCount() != 1 {
		t.Errorf("room should contain 1 peer, but has %d peers", room.PeersCount())
	}
	p, ok := room.GetPeer(peer.UID)
	if !ok {
		t.Errorf("could not find peer with uid %q", peer.UID)
	}
	if ok && p != peer {
		t.Errorf("peer wasn't updated, got %+v, want %+v", p, peer)
	}
}
