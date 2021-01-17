package broker

import (
	"sync"

	"github.com/montrosesoftware/tarpon/pkg/logging"
	"github.com/montrosesoftware/tarpon/pkg/messaging"
)

type Subscriber interface {
	Write(m messaging.Message)
	ID() string
}

type Broker interface {
	Send(room string, message messaging.Message)
	Register(room string, s Subscriber)
	Unregister(room string, s Subscriber) bool
}

type InMemoryBroker struct {
	subscribers map[string][]Subscriber
	mutex       sync.RWMutex
	logger      logging.Logger
}

func NewBroker(l logging.Logger) *InMemoryBroker {
	return &InMemoryBroker{subscribers: make(map[string][]Subscriber), logger: l}
}

func (b *InMemoryBroker) Send(room string, message messaging.Message) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	if message.IsBroadcast() {
		b.broadcast(message, b.subscribers[room])
	} else {
		b.sendDirect(message, b.subscribers[room])
	}
}

func (b *InMemoryBroker) Register(room string, s Subscriber) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.subscribers[room] = append(b.subscribers[room], s)
	b.logger.Info("subscriber registered", logging.Fields{"room": room, "subscriber": s.ID(), "subscribers_count": len(b.subscribers[room])})
}

func (b *InMemoryBroker) Unregister(room string, s Subscriber) bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	roomSubs := b.subscribers[room]
	for i, subscriber := range roomSubs {
		if subscriber == s {
			if len(roomSubs) == 1 {
				delete(b.subscribers, room)
			} else {
				roomSubs[i] = roomSubs[len(roomSubs)-1]
				b.subscribers[room] = roomSubs[:len(roomSubs)-1]
			}
			b.logger.Info("subscriber unregistered", logging.Fields{"room": room, "subscriber": s.ID(), "subscribers_count": len(b.subscribers[room])})
			return true
		}
	}
	b.logger.Warn("tried to unregister subscriber, but it seems not registered", logging.Fields{"room": room, "subscriber": s.ID()})
	return false
}

func (b *InMemoryBroker) RoomsCount() int {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	return len(b.subscribers)
}

func (b *InMemoryBroker) broadcast(message messaging.Message, subscribers []Subscriber) {
	for _, subscriber := range subscribers {
		subscriber.Write(message)
	}
}

func (b *InMemoryBroker) sendDirect(message messaging.Message, subscribers []Subscriber) {
	for _, subscriber := range subscribers {
		if subscriber.ID() == message.To {
			subscriber.Write(message)
		}
	}
}
