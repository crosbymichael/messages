package main

import (
	"fmt"
	"github.com/crosbymichael/messages"
	"time"
)

type Feed struct {
	Url    string
	Author string
}

func main() {
	mbox := messages.NewMailbox("feeds", "tcp", "127.0.0.1:6379", "")
	feed := &Feed{
		Url:    "http://crosbymichael.com/feeds/all.atom.xml",
		Author: "Michael",
	}

	m := mbox.NewMessage()
	if err := m.Marshal(feed); err != nil {
		panic(err)
	}

	if err := mbox.Send(m); err != nil {
		panic(err)
	}

	m, err := mbox.Wait()
	if err != nil {
		panic(err)
	}

	var newFeed Feed
	if err := m.Unmarshal(&newFeed); err != nil {
		panic(err)
	}

	t, err := m.Time()
	if err != nil {
		panic(err)
	}
	fmt.Printf("ID: %s\nMailbox: %s\nCreated: %s\n", m.ID, m.Mailbox, t.Format(time.RubyDate))
	fmt.Printf("Body: %v\n", newFeed)

	if err := mbox.DestoryAfter(m, 500); err != nil {
		panic(err)
	}
}
