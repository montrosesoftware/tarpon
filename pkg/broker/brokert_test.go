package broker_test

import (
	"reflect"
	"sync"
	"testing"

	"github.com/montrosesoftware/tarpon/pkg/broker"
	"github.com/montrosesoftware/tarpon/pkg/messaging"
)

var (
	room1 = "room-123"
	room2 = "room-xyz"
	peer1 = "peer-1"
	peer2 = "peer-2"
	peer3 = "peer-3"
)

type SpySubscriber struct {
	id       string
	messages []messaging.Message
	mutex    sync.Mutex
}

func (s *SpySubscriber) ID() string {
	return s.id
}

func (s *SpySubscriber) Write(m messaging.Message) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.messages = append(s.messages, m)
}

func (s *SpySubscriber) assertMessages(t *testing.T, ms []messaging.Message) {
	if !reflect.DeepEqual(s.messages, ms) {
		t.Errorf("%q got %v messages, but expected %v", s.id, s.messages, ms)
	}
}

func TestRegistration(t *testing.T) {
	broker := broker.NewBroker()
	subscriber := &SpySubscriber{id: peer1}

	broker.Register(room1, subscriber)
	broker.Register(room1, subscriber)

	if !broker.Unregister(room1, subscriber) {
		t.Errorf("subscriber %q not unregistered during 1st attempt", subscriber)
	}

	if !broker.Unregister(room1, subscriber) {
		t.Errorf("subscriber %q not unregistered during 2nd attempt", subscriber)
	}

	if broker.Unregister(room1, subscriber) {
		t.Errorf("subscriber %q unregistered, but it shouldn't", subscriber)
	}

	if broker.Unregister("invalid room", subscriber) {
		t.Errorf("subscriber %q unregistered, but room doesn't exist", subscriber)
	}
}

func TestSendingMessages(t *testing.T) {
	broker := broker.NewBroker()
	subscriber1 := &SpySubscriber{id: peer1}
	subscriber2 := &SpySubscriber{id: peer2}
	subscriber3 := &SpySubscriber{id: peer3}
	subscriber33 := &SpySubscriber{id: peer3} // second instance of peer3

	broker.Register(room1, subscriber1)
	broker.Register(room1, subscriber2)
	broker.Register(room2, subscriber3)
	broker.Register(room2, subscriber33)

	m1 := messaging.Message{To: subscriber1.id}
	broker.Send(room1, m1)
	subscriber1.assertMessages(t, []messaging.Message{m1})
	subscriber2.assertMessages(t, nil)
	subscriber3.assertMessages(t, nil)
	subscriber33.assertMessages(t, nil)

	m2 := messaging.Message{To: ""} // broadcast message
	broker.Send(room1, m2)
	subscriber1.assertMessages(t, []messaging.Message{m1, m2})
	subscriber2.assertMessages(t, []messaging.Message{m2})
	subscriber3.assertMessages(t, nil)
	subscriber33.assertMessages(t, nil)

	m3 := messaging.Message{To: subscriber3.id}
	broker.Send(room2, m3)
	subscriber1.assertMessages(t, []messaging.Message{m1, m2})
	subscriber2.assertMessages(t, []messaging.Message{m2})
	subscriber3.assertMessages(t, []messaging.Message{m3})
	subscriber33.assertMessages(t, []messaging.Message{m3})

	broker.Send("invalid room", m3)
	subscriber1.assertMessages(t, []messaging.Message{m1, m2})
	subscriber2.assertMessages(t, []messaging.Message{m2})
	subscriber3.assertMessages(t, []messaging.Message{m3})
	subscriber33.assertMessages(t, []messaging.Message{m3})
}

func TestConcurrentSends(t *testing.T) {
	broker := broker.NewBroker()
	subscriber1 := &SpySubscriber{id: peer1}
	subscriber2 := &SpySubscriber{id: peer2}

	broker.Register(room1, subscriber1)
	broker.Register(room1, subscriber2)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		m := messaging.Message{To: subscriber1.id}
		for i := 0; i < 1000; i++ {
			broker.Send(room1, m)
		}
		wg.Done()
	}()

	go func() {
		m := messaging.Message{To: ""}
		for i := 0; i < 1000; i++ {
			broker.Send(room1, m)
		}
		wg.Done()
	}()

	wg.Wait()

	if len(subscriber1.messages) != 2000 {
		t.Errorf("subscriber 1 received %d messages, but wanted 2000", len(subscriber1.messages))
	}

	if len(subscriber2.messages) != 1000 {
		t.Errorf("subscriber 2 received %d messages, but wanted 1000", len(subscriber1.messages))
	}
}
