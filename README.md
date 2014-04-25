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

###License
MIT

Copyright (c) 2013 Michael Crosby. michael@crosbymichael.com

Permission is hereby granted, free of charge, to any person
obtaining a copy of this software and associated documentation 
files (the "Software"), to deal in the Software without 
restriction, including without limitation the rights to use, copy, 
modify, merge, publish, distribute, sublicense, and/or sell copies 
of the Software, and to permit persons to whom the Software is 
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be 
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED,
INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, 
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. 
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT 
HOLDERS BE LIABLE FOR ANY CLAIM, 
DAMAGES OR OTHER LIABILITY, 
WHETHER IN AN ACTION OF CONTRACT, 
TORT OR OTHERWISE, 
ARISING FROM, OUT OF OR IN CONNECTION WITH 
THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
