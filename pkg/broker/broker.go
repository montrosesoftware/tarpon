package broker

import (
	"sync"

	"github.com/montrosesoftware/tarpon/pkg/messaging"
)

type Subscriber interface {
	Write(m messaging.Message)
	ID() string
}

type Broker interface {
	Send(room string, message messaging.Message)
	Register(room string, s Subscriber)
	Unregister(room string, s Subscriber)
}

type InMemoryBroker struct {
	subscribers map[string][]Subscriber
	mutex       sync.RWMutex
}

func NewBroker() *InMemoryBroker {
	return &InMemoryBroker{subscribers: make(map[string][]Subscriber)}
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
}

func (b *InMemoryBroker) Unregister(room string, s Subscriber) bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	roomSubs := b.subscribers[room]
	for i, subscriber := range roomSubs {
		if subscriber == s {
			roomSubs[i] = roomSubs[len(roomSubs)-1]
			b.subscribers[room] = roomSubs[:len(roomSubs)-1]
			return true
		}
	}
	return false
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
