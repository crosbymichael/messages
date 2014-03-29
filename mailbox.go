package messages

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/crosbymichael/messages/stats"
	"github.com/garyburd/redigo/redis"
	"io"
	"time"
)

type Mailbox interface {
	NewMessage() *Message
	Send(*Message) error
	Wait() (*Message, error)
	Destroy(*Message) error
	DestoryAfter(*Message, int) error
}

// Place to send and receive messages
type mailbox struct {
	name               string // name of the mailbox
	defaultWaitTimeout int    // How long to wait
	pool               *redis.Pool
}

// Create a new named mailbox to send and receive
// messages on.
func NewMailbox(name, proto, addr, password string) Mailbox {
	return &mailbox{
		name:               name,
		defaultWaitTimeout: 0, // Default to 0 so that the mailbox blocks forever
		pool:               newPool(proto, addr, password),
	}
}

// Create a new message from the current Mailbox
func (mbox *mailbox) NewMessage() *Message {
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
		ID:      fmt.Sprintf("message:%s:%s", mbox.name, hex.EncodeToString(hash.Sum(nil))),
		Created: created,
		Mailbox: mbox.name,
	}
}

// Send the message
func (mbox *mailbox) Send(m *Message) error {
	stats.MessageCount.Inc(1)

	conn := mbox.pool.Get()
	defer conn.Close()

	args := []interface{}{
		m.ID,
		"mailbox", m.Mailbox,
		"created", m.Created,
		"body", m.Body,
	}

	if err := conn.Send("MULTI"); err != nil {
		return err
	}
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
func (mbox *mailbox) Wait() (*Message, error) {
	reply, err := redis.MultiBulk(mbox.send("BLPOP", fmt.Sprintf("%s:messages", mbox.key()), mbox.defaultWaitTimeout))
	if err != nil {
		return nil, err
	}
	return mbox.messageFromId(string(reply[1].([]byte)))
}

// Delete the message from the transport NOW
func (mbox *mailbox) Destroy(m *Message) error {
	return mbox.DestoryAfter(m, 0)
}

// Delete the message after n seconds
func (mbox *mailbox) DestoryAfter(m *Message, seconds int) error {
	if seconds < 1 {
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

func (mbox *mailbox) send(cmd string, args ...interface{}) (interface{}, error) {
	conn := mbox.pool.Get()
	defer conn.Close()
	return conn.Do(cmd, args...)
}

func (mbox *mailbox) key() string {
	return fmt.Sprintf("mailbox:%s", mbox.name)
}

func (mbox *mailbox) messageFromId(id string) (*Message, error) {
	result, err := redis.MultiBulk(mbox.send("HGETALL", id))
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
