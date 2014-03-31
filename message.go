package messages

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"time"
)

var (
	DefaultPoolSize = 10 // Number of redis connections to keep in the pool
)

// Message to a recipient
type Message struct {
	ID      string `json:"id"`      // UUID of a message
	Created string `json:"created"` // When the message was created
	Body    []byte `json:"body"`    // Body of the message
}

// Create a new message
func NewMessage() *Message {
	var (
		created = time.Now().Format(time.UnixDate)
		buff    = make([]byte, 32)
		hash    = md5.New()
	)

	if _, err := io.ReadFull(rand.Reader, buff); err != nil {
		panic(err)
	}
	hash.Write(buff)
	hash.Write([]byte(created))

	return &Message{
		ID:      hex.EncodeToString(hash.Sum(nil)),
		Created: created,
	}
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
