package messaging

// NewRealm creates a new communication realm.
func NewRealm(UID string) *Realm {
	r := Realm{
		UID:         UID,
		Subscribers: make(map[string]*Subscriber),
	}

	return &r
}

// Realm represents subsribers that can send messages to each other.
type Realm struct {
	UID         string                 `json:"uid"`
	Subscribers map[string]*Subscriber `json:"subscribers"`
}

var ()

func (r *Realm) Register(s *Subscriber) {

}
