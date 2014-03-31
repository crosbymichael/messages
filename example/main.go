package main

import (
	"fmt"
	"github.com/crosbymichael/messages"
	"sync"
	"time"
)

type Feed struct {
	Url    string
	Author string
}

func worker(c chan *messages.Message, group *sync.WaitGroup, mbox messages.Mailbox) {
	defer group.Done()

	for m := range c {
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
}

func main() {
	mbox := messages.NewMailbox("feeds", "tcp", "127.0.0.1:6379", "")
	defer mbox.Close()

	defer func(now time.Time) {
		fmt.Printf("Finished in %s\n", time.Now().Sub(now))
	}(time.Now())

	var (
		c     = make(chan *messages.Message)
		group = &sync.WaitGroup{}
		feed  = &Feed{
			Url:    "http://crosbymichael.com/feeds/all.atom.xml",
			Author: "Michael",
		}
	)

	for i := 0; i < 10; i++ {
		group.Add(1)
		go worker(c, group, mbox)
	}

	go func() {
		for i := 0; i < 10000; i++ {
			m, err := mbox.Wait()
			if err != nil {
				panic(err)
			}
			c <- m
		}
		close(c)
	}()

	for i := 0; i < 10000; i++ {
		m := mbox.NewMessage()
		if err := m.Marshal(feed); err != nil {
			panic(err)
		}
		if err := mbox.Send(m); err != nil {
			panic(err)
		}
	}

	group.Wait()
}
