##Messages

Messages is a very simple go pkg to send and receive messages using redis as the broker.  The message format is simple so other languages can implement the format.
A message consists of it's `ID`, unix timestamp when it was `Created`, the `Mailbox` that was used to send it, and the `Body` which is a json object.

You configure the redis connection via environment variables:

```bash
REDIS_PROTO=tcp
REDIS_ADDR=127.0.0.0:6379
```

###Example

**Create a new Mailbox**

```golang
mbox := messages.NewMailbox("feeds") 
```


**Create a new message with a json encoded body**
```golang
feed := &Feed{
    Url:    "http://crosbymichael.com/feeds/all.atom.xml",
    Author: "Michael",
}

m := mbox.NewMessage()
if err := m.Marshal(feed); err != nil {
    panic(err)
}
```

**Send the message to the mailbox**
```golang
if err := mbox.Send(m); err != nil {
    panic(err)
}
```

**Wait for messages to arrive in the mailbox**
```golang
m, err := mbox.Wait()
if err != nil {
    panic(err)
}
```

**Read data from the message**
```golang
var newFeed Feed
if err := m.Unmarshal(&newFeed); err != nil {
    panic(err)
}

fmt.Printf("ID: %s\nMailbox: %s\nCreated: %s\n", m.ID, m.Mailbox, m.Time().Format(time.RubyDate))
fmt.Printf("Body: %v\n", newFeed)
```

**Destroy the message data from the mailbox after 500 seconds**
```golang
if err := mbox.DestoryAfter(m, 500); err != nil {
    panic(err)
}
```

**Result**
```bash
ID: message:feeds:ec9cf5aee06277cff5dd0ea3e18c79db5d6422e731704e1f24bcfe189d93c7b5
Mailbox: feeds
Created: Sun Sep 15 18:27:25 +0000 2013
Body: {http://crosbymichael.com/feeds/all.atom.xml Michael}
```

Simple messaging that works across process or the world.

###Code Coverage
github.com/crosbymichael/messages/messages.go	 argsToMap		 100.00% (6/6)
github.com/crosbymichael/messages/messages.go	 defaultEnv		 100.00% (4/4)
github.com/crosbymichael/messages/messages.go	 Mailbox.send		 100.00% (3/3)
github.com/crosbymichael/messages/messages.go	 Mailbox.Len		 100.00% (2/2)
github.com/crosbymichael/messages/messages.go	 Mailbox.Destroy	 100.00% (1/1)
github.com/crosbymichael/messages/messages.go	 Message.Unmarshal	 100.00% (1/1)
github.com/crosbymichael/messages/messages.go	 NewMailbox		 100.00% (1/1)
github.com/crosbymichael/messages/messages.go	 Mailbox.key		 100.00% (1/1)
github.com/crosbymichael/messages/messages.go	 Message.Time		 100.00% (1/1)
github.com/crosbymichael/messages/messages.go	 Mailbox.NewMessage	 80.00% (4/5)
github.com/crosbymichael/messages/messages.go	 Mailbox.DestoryAfter	 71.43% (5/7)
github.com/crosbymichael/messages/messages.go	 Mailbox.Wait		 33.33% (3/9)
github.com/crosbymichael/messages/messages.go	 newPool		 25.00% (1/4)
github.com/crosbymichael/messages/messages.go	 Mailbox.Send		 0.00% (0/13)
github.com/crosbymichael/messages/messages.go	 newMessageFromId	 0.00% (0/9)
github.com/crosbymichael/messages		 --------------------	 49.25% (33/67)


###License
MIT
