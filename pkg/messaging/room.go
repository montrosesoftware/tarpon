package messaging

type Room struct {
	peers []Peer
}

func (r *Room) RegisterPeer(peer Peer) bool {
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
	for _, p := range r.peers {
		if p.UID == uid {
			return p, true
		}
	}
	return Peer{}, false
}

func (r *Room) PeersCount() int {
	return len(r.peers)
}
