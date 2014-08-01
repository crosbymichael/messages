package messages

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
)

type Mailbox interface {
	Send(*Message) error
	Wait() (*Message, error)
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
	conn := mbox.pool.Get()
	defer conn.Close()

	reply, err := redis.MultiBulk(conn.Do("BLPOP", fmt.Sprintf("mailbox:%s", mbox.name), mbox.defaultWaitTimeout))
	if err != nil {
		return nil, err
	}

	id := string(reply[1].([]byte))
	result, err := redis.MultiBulk(conn.Do("HGETALL", fmt.Sprintf("messages:%s", id)))
	if err != nil {
		return nil, err
	}

	args := make([][]byte, len(result))
	for i := 0; i < len(result); i++ {
		args[i] = result[i].([]byte)
	}

	hash := argsToMap(args)

	return &Message{
		ID:      id,
		Created: string(hash["created"]),
		Body:    hash["body"],
	}, nil
}

// Delete the message after n seconds
func (mbox *mailbox) DestroyAfter(m *Message, seconds int) error {
	var (
		key  = fmt.Sprintf("messages:%s", m.ID)
		conn = mbox.pool.Get()
	)
	defer conn.Close()

	if seconds < 1 {
		if _, err := conn.Do("DEL", key); err != nil {
			return err
		}

		return nil
	}

	if _, err := conn.Do("EXPIRE", key, seconds); err != nil {
		return err
	}

	return nil
}
