package messages

import (
	"encoding/json"
	"time"
)

var (
	DefaultPoolSize = 10 // Number of redis connections to keep in the pool
)

// Message to a recipient
type Message struct {
	ID      string // UUID of a message
	Created string // When the message was created
	Mailbox string // Name of the mailbox used to send the message
	Body    []byte // Body of the message
}

// Return a time.Time for the created timestamp
func (m *Message) Time() (time.Time, error) {
	return time.Parse(time.UnixDate, m.Created)
}

// Unmarshal the body of the message into a type
func (m *Message) Unmarshal(v interface{}) error {
	return json.Unmarshal(m.Body, v)
}

// Marshal a type as the body of the message
func (m *Message) Marshal(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	m.Body = data
	return nil
}
