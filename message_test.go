package messages

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"os"
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
	mbox := NewMailbox("test")
	if mbox == nil {
		t.FailNow()
	}

	assertEquals(mbox.Name, "test", t)
	assertEquals(mbox.DefaultWaitTimeout, 0, t)

	assertIsNotNil(mbox.pool, t)
}

func TestUnmarshalBody(t *testing.T) {
	body := &TestBody{
		Name: "koye",
		Age:  3,
	}

	mbox := NewMailbox("test")

	m := mbox.NewMessage()
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
	mbox := NewMailbox("test")

	m := mbox.NewMessage()
	assertIsNotNil(m, t)

	assertEquals(m.Mailbox, "test", t)

	if m.ID == "" {
		t.Fail()
	}

	if m.Created == "" {
		t.Fail()
	}
}

func TestDefaultEnv(t *testing.T) {
	os.Setenv("MAILBOX_TEST", "test")
	v := defaultEnv("MAILBOX_TEST", "nottest")
	assertEquals(v, "test", t)

	v = defaultEnv("MAILBOX_NOTEST", "notest")
	assertEquals(v, "notest", t)
}

func TestNewPool(t *testing.T) {
	pool := newPool()
	assertIsNotNil(pool, t)
}

func TestMailboxLen(t *testing.T) {
	mbox := NewMailbox("test")
	mbox.pool = newTestPool(func(c string, args ...interface{}) (interface{}, error) {
		if len(args) > 0 {
			assertEquals(c, "LLEN", t)
			key := args[0]
			assertEquals(key.(string), "mailbox:test:messages", t)
			return int64(5), nil
		}
		return nil, nil
	}, nil)
	assertEquals(mbox.Len(), int64(5), t)
}

func TestMailboxDestroy(t *testing.T) {
	mbox := NewMailbox("test")
	m := mbox.NewMessage()

	mbox.pool = newTestPool(func(c string, args ...interface{}) (interface{}, error) {
		if len(args) > 0 {
			assertEquals(c, "DEL", t)
			key := args[0]
			assertEquals(key.(string), m.ID, t)
		}
		return nil, nil
	}, nil)

	if err := mbox.Destroy(m); err != nil {
		t.Fatal(err)
	}
}

func TestMailboxDestroyAfter(t *testing.T) {
	mbox := NewMailbox("test")
	m := mbox.NewMessage()

	mbox.pool = newTestPool(func(c string, args ...interface{}) (interface{}, error) {
		if len(args) > 0 {
			assertEquals(c, "EXPIRE", t)
			key := args[0]
			assertEquals(key.(string), m.ID, t)
			sec := args[1]
			assertEquals(sec.(int), 10, t)
		}
		return nil, nil
	}, nil)

	if err := mbox.DestoryAfter(m, 10); err != nil {
		t.Fatal(err)
	}
}

func TestMailboxWait(t *testing.T) {
	mbox := NewMailbox("test")

	mbox.pool = newTestPool(func(c string, args ...interface{}) (interface{}, error) {
		if len(args) > 0 {
			assertEquals(c, "BLPOP", t)

			key := args[0]
			assertEquals(key.(string), "mailbox:test:messages", t)
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
	mbox := NewMailbox("test")
	for i := 0; i < b.N; i++ {
		m := mbox.NewMessage()
		if err := mbox.Send(m); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMessageReads(b *testing.B) {
	mbox := NewMailbox("test")
	for i := 0; i < b.N; i++ {
		m := mbox.NewMessage()
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
