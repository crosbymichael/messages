package messages

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"testing"
	"time"
)

type TestBody struct {
	Name string
	Age  int
}

type doHandler func(commandName string, args ...interface{}) (interface{}, error)
type sendHandler func(commandName string, args ...interface{}) error

type testConn struct {
	doHandler   doHandler
	sendHandler sendHandler
}

func (c *testConn) Err() error {
	return nil
}

func (c *testConn) Close() error {
	return nil
}

func (c *testConn) Do(commandName string, args ...interface{}) (interface{}, error) {
	return c.doHandler(commandName, args...)
}

func (c *testConn) Send(commandName string, args ...interface{}) error {
	return c.sendHandler(commandName, args...)
}

func (c *testConn) Flush() error {
	return fmt.Errorf("Not implemented")
}

func (c *testConn) Receive() (interface{}, error) {
	return nil, fmt.Errorf("Not implemented")
}

func newTestPool(do doHandler, send sendHandler) *redis.Pool {
	return redis.NewPool(func() (redis.Conn, error) {
		return &testConn{do, send}, nil
	}, 0)
}

func TestGetCreateTime(t *testing.T) {
	now := time.Now()

	m := &Message{Created: now.Format(time.UnixDate)}

	tm, err := m.Time()
	if err != nil {
		t.Fatal(err)
	}
	assertEquals(tm.Unix(), now.Unix(), t)
}

func TestNewMailbox(t *testing.T) {
	mbox := NewMailbox("test", "", "", "").(*mailbox)
	if mbox == nil {
		t.FailNow()
	}

	assertEquals(mbox.name, "test", t)
	assertEquals(mbox.defaultWaitTimeout, 0, t)

	assertIsNotNil(mbox.pool, t)
}

func TestUnmarshalBody(t *testing.T) {
	body := &TestBody{
		Name: "koye",
		Age:  3,
	}

	m := NewMessage()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	m.Body = data

	var actual TestBody

	if err := m.Unmarshal(&actual); err != nil {
		t.Fatal(err)
	}
	assertEquals(actual.Name, body.Name, t)
	assertEquals(actual.Age, body.Age, t)
}

func TestNewMessage(t *testing.T) {
	m := NewMessage()
	assertIsNotNil(m, t)

	if m.ID == "" {
		t.Fail()
	}

	if m.Created == "" {
		t.Fail()
	}
}

func TestNewPool(t *testing.T) {
	pool := newPool("", "", "")
	assertIsNotNil(pool, t)
}

func TestMailboxDestroy(t *testing.T) {
	mbox := NewMailbox("test", "", "", "").(*mailbox)
	m := NewMessage()

	mbox.pool = newTestPool(func(c string, args ...interface{}) (interface{}, error) {
		if len(args) > 0 {
			assertEquals(c, "DEL", t)
			key := args[0]
			assertEquals(key.(string), fmt.Sprintf("messages:%s", m.ID), t)
		}
		return nil, nil
	}, nil)

	if err := mbox.Destroy(m); err != nil {
		t.Fatal(err)
	}
}

func TestMailboxDestroyAfter(t *testing.T) {
	mbox := NewMailbox("test", "", "", "").(*mailbox)
	m := NewMessage()

	mbox.pool = newTestPool(func(c string, args ...interface{}) (interface{}, error) {
		if len(args) > 0 {
			assertEquals(c, "EXPIRE", t)
			key := args[0]
			assertEquals(key.(string), fmt.Sprintf("messages:%s", m.ID), t)
			sec := args[1]
			assertEquals(sec.(int), 10, t)
		}
		return nil, nil
	}, nil)

	if err := mbox.DestroyAfter(m, 10); err != nil {
		t.Fatal(err)
	}
}

func TestMailboxWait(t *testing.T) {
	mbox := NewMailbox("test", "", "", "").(*mailbox)

	mbox.pool = newTestPool(func(c string, args ...interface{}) (interface{}, error) {
		if len(args) > 0 {
			assertEquals(c, "BLPOP", t)

			key := args[0]
			assertEquals(key.(string), "mailbox:test", t)
			timeout := args[1]
			assertEquals(timeout.(int), 0, t)

			return nil, fmt.Errorf("Complete")
		}
		return nil, nil
	}, nil)

	if _, err := mbox.Wait(); err == nil {
		t.Log("Error should not be nil")
		t.FailNow()
	}
}

func TestArgsToMap(t *testing.T) {
	args := [][]byte{
		[]byte("name"),
		[]byte("koye"),
		[]byte("age"),
		[]byte("3"),
	}

	m := argsToMap(args)

	assertEquals(string(m["name"]), "koye", t)
	assertEquals(string(m["age"]), "3", t)
}

func TestMarshalAndUnmarshal(t *testing.T) {
	m := &Message{}

	err := m.Marshal(struct {
		Name string
		Age  int
	}{
		Name: "koye",
		Age:  3,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func BenchmarkMessageWrites(b *testing.B) {
	mbox := NewMailbox("test", "", "", "").(*mailbox)
	for i := 0; i < b.N; i++ {
		m := NewMessage()
		if err := mbox.Send(m); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMessageReads(b *testing.B) {
	mbox := NewMailbox("test", "", "", "").(*mailbox)
	for i := 0; i < b.N; i++ {
		m := NewMessage()
		if err := mbox.Send(m); err != nil {
			b.Fatal(err)
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := mbox.Wait(); err != nil {
			b.Fatal(err)
		}
	}
}

func assertIsNotNil(v interface{}, t *testing.T) {
	if v == nil {
		t.Log("Value should not be nil")
		t.FailNow()
	}
}

func assertEquals(actual, expected interface{}, t *testing.T) {
	if actual != expected {
		t.Logf("Actual value %v does not equal expected value %v", actual, expected)
		t.Fail()
	}
}
