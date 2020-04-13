package messaging

// Subscriber represents entities allowed to send and receive messages.
type Subscriber struct {
	UID    string `json:"uid"`
	Secret string `json:"secret"`
}
