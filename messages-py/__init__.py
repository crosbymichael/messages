import os
import redis
from datetime import datetime as dt
import json
from hashlib import md5
import random, string

def _new_id(mailbox):
    n = dt.now()
    created = n.timestamp().__str__() 
    h = md5() 
    buf = ''.join(random.choice(string.ascii_uppercase + string.digits) for x in xrange(32))
    h.write(buff)
    h.write(created)
    return ('message:%s:%s' % (h.hexdigext(), mailbox), created,)

class Message(object):
    def __init__(self, id, created, mailbox):
        self.id = id
        self.mailbox = mailbox
        self.created = dt.now()
        self.body = None

    def load(self):
        """ Return the body as a python dict """
        return json.loads(self.body)

    def dump(self, obj):
        """ Dump the python object into the body """
        self.body = json.dumps(obj)

class Mailbox(object):
    def __init__(self, name):
        """ Create a new mailbox with the name """
        self.name = name
        self.default_wait_timeout = 0
        proto = os.environ.get('REDIS_PROTO', 'tcp')
        addr = os.environ.get('REDIS_ADDR', 'localhost:6379')
        p = addr.split(':')
        if proto is 'unix':
            self._pool = redis.ConnectionPool(unix_socket_path=addr)
        else:
            self._pool = redis.ConnectionPool(host=p[0], port=int(p[1]))

    def send(self, m):
        """ Send the message """
    
    def delete(self, m):
        """ Delete the message from the mailbox now """
        return self.delete_after(m, 0)

    def delete_after(self, m, seconds):
        """ Delete the message from the mailbox after n seconds """
        conn = redis.Redis(connection_pool=self._pool)
        if seconds == 0:
            conn.delete(m.id)
        else:
            conn.expire(m.id, seconds)
    
    def new_message(self):
        """ Return a new message """
        id, created = _new_id(self.name)
        return Message(id, created, self.name)


    def wait(self):
        """ Wait for new messages in the mailbox """
        pass
