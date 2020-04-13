package messaging

// Realm represents subsribers that can send messages to each other.
type Realm struct {
	UID         string                 `json:"uid"`
	Subscribers map[string]*Subscriber `json:"subscribers"`
}
