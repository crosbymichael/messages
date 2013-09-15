package messages

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"io"
	"os"
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

// Place to send and receive messages
type Mailbox struct {
	Name               string // Name of the mailbox
	DefaultWaitTimeout int    // How long
	pool               *redis.Pool
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

// Create a new named mailbox to send and receive
// messages on.
func NewMailbox(name string) *Mailbox {
	return &Mailbox{
		Name:               name,
		DefaultWaitTimeout: 0, // Default to 0 so that the mailbox blocks forever
		pool:               newPool(),
	}
}

func newPool() *redis.Pool {
	return redis.NewPool(func() (redis.Conn, error) {
		proto := defaultEnv("REDIS_PROTO", "tcp")
		addr := defaultEnv("REDIS_ADDR", ":6379")

		return redis.Dial(proto, addr)
	}, DefaultPoolSize)
}

// Create a new message from the current Mailbox
func (mbox *Mailbox) NewMessage() *Message {
	buff := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, buff); err != nil {
		panic(err)
	}
	id := fmt.Sprintf("message:%s:%s", mbox.Name, hex.EncodeToString(buff))

	return &Message{
		ID:      id,
		Created: time.Now().Format(time.UnixDate),
		Mailbox: mbox.Name,
	}
}

// Send the message
func (mbox *Mailbox) Send(m *Message) error {
	conn := mbox.pool.Get()
	defer conn.Close()

	if err := conn.Send("MULTI"); err != nil {
		return err
	}

	args := []interface{}{m.ID}
	args = append(args,
		"mailbox", m.Mailbox,
		"created", m.Created,
		"body", m.Body)

	if err := conn.Send("HMSET", args...); err != nil {
		return err
	}

	if err := conn.Send("RPUSH", fmt.Sprintf("%s:messages", mbox.key()), m.ID); err != nil {
		return err
	}

	if _, err := conn.Do("EXEC"); err != nil {
		return err
	}
	return nil
}

// Wait for a new message to be delivered using the
// timeout specified for the mailbox
func (mbox *Mailbox) Wait() (*Message, error) {
	reply, err := redis.MultiBulk(mbox.send("BLPOP", fmt.Sprintf("%s:messages", mbox.key()), mbox.DefaultWaitTimeout))
	if err != nil {
		return nil, err
	}

	id := string(reply[1].([]byte))

	conn := mbox.pool.Get()
	defer conn.Close()

	return newMessageFromId(id, conn)
}

// Delete the message from the transport NOW
func (mbox *Mailbox) Destroy(m *Message) error {
	return mbox.DestoryAfter(m, 0)
}

// Delete the message after n seconds
func (mbox *Mailbox) DestoryAfter(m *Message, seconds int) error {
	if seconds <= 0 {
		if _, err := mbox.send("DEL", m.ID); err != nil {
			return err
		}
		return nil
	}

	if _, err := mbox.send("EXPIRE", m.ID, seconds); err != nil {
		return err
	}
	return nil
}

// Return the number of Messages in the Mailbox
func (mbox *Mailbox) Len() int64 {
	l, _ := redis.Int64(mbox.send("LLEN", fmt.Sprintf("%s:messages", mbox.key())))
	return l
}

func (mbox *Mailbox) send(cmd string, args ...interface{}) (interface{}, error) {
	conn := mbox.pool.Get()
	defer conn.Close()
	return conn.Do(cmd, args...)
}

func (mbox *Mailbox) key() string {
	return fmt.Sprintf("mailbox:%s", mbox.Name)
}

func newMessageFromId(id string, conn redis.Conn) (*Message, error) {
	result, err := redis.MultiBulk(conn.Do("HGETALL", id))
	if err != nil {
		return nil, err
	}
	args := make([][]byte, len(result))

	for i := 0; i < len(result); i++ {
		data := result[i].([]byte)
		args[i] = data
	}

	hash := argsToMap(args)

	return &Message{
		ID:      id,
		Created: string(hash["created"]),
		Body:    hash["body"],
		Mailbox: string(hash["mailbox"]),
	}, nil
}

func argsToMap(args [][]byte) map[string][]byte {
	result := make(map[string][]byte, len(args)/2)
	for i := 0; i < len(args); i++ {
		key := string(args[i])
		i++
		result[key] = args[i]
	}
	return result
}

// Get a value from the machines environment or use the
// default value
func defaultEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		value = defaultValue
	}
	return value
}
