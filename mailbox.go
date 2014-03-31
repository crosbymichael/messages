package messages

import (
	"fmt"
	"github.com/crosbymichael/messages/stats"
	"github.com/garyburd/redigo/redis"
)

type Mailbox interface {
	Send(*Message) error
	Wait() (*Message, error)
	Destroy(*Message) error
	DestroyAfter(*Message, int) error
	Close() error
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

func (mbox *mailbox) Close() error {
	return mbox.pool.Close()
}

// Send the message
func (mbox *mailbox) Send(m *Message) error {
	stats.MessageCount.Inc(1)

	conn := mbox.pool.Get()
	defer conn.Close()

	args := []interface{}{
		fmt.Sprintf("messages:%s", m.ID),
		"created", m.Created,
		"body", m.Body,
	}

	if err := conn.Send("MULTI"); err != nil {
		return err
	}
	if err := conn.Send("HMSET", args...); err != nil {
		return err
	}
	if err := conn.Send("RPUSH", fmt.Sprintf("mailbox:%s", mbox.name), m.ID); err != nil {
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
	reply, err := redis.MultiBulk(mbox.send("BLPOP", fmt.Sprintf("mailbox:%s", mbox.name), mbox.defaultWaitTimeout))
	if err != nil {
		return nil, err
	}
	return mbox.messageFromId(string(reply[1].([]byte)))
}

// Delete the message from the transport NOW
func (mbox *mailbox) Destroy(m *Message) error {
	return mbox.DestroyAfter(m, 0)
}

// Delete the message after n seconds
func (mbox *mailbox) DestroyAfter(m *Message, seconds int) error {
	if seconds < 1 {
		if _, err := mbox.send("DEL", fmt.Sprintf("messages:%s", m.ID)); err != nil {
			return err
		}
		return nil
	}
	if _, err := mbox.send("EXPIRE", fmt.Sprintf("messages:%s", m.ID), seconds); err != nil {
		return err
	}
	return nil
}

func (mbox *mailbox) send(cmd string, args ...interface{}) (interface{}, error) {
	conn := mbox.pool.Get()
	defer conn.Close()
	return conn.Do(cmd, args...)
}

func (mbox *mailbox) messageFromId(id string) (*Message, error) {
	result, err := redis.MultiBulk(mbox.send("HGETALL", fmt.Sprintf("messages:%s", id)))
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
	}, nil
}
