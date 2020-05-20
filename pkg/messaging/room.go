package messaging

import (
	"sync"
)

type Room struct {
	peers []Peer
	mutex sync.RWMutex
}

func (r *Room) RegisterPeer(peer Peer) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for i, p := range r.peers {
		if p.UID == peer.UID {
			r.peers[i] = peer
			return false
		}
	}
	r.peers = append(r.peers, peer)
	return true
}

func (r *Room) GetPeer(uid string) (Peer, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, p := range r.peers {
		if p.UID == uid {
			return p, true
		}
	}
	return Peer{}, false
}

func (r *Room) PeersCount() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return len(r.peers)
}

func (r *Room) Join(secret string) (Peer, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, p := range r.peers {
		if p.Secret == secret {
			return p, nil
		}
	}
	return Peer{}, ErrUnauthorized
}
