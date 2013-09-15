import os
import redis
from datetime import datetime as dt
import json

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
        proto = os.environ.get('REDIS_PROTO', 'tcp')
        addr = os.environ.get('REDIS_ADDR', 'localhost:6379')
        p = addr.split(':')
        if proto is 'unix':
            self._pool = redis.ConnectionPool(unix_socket_path=addr)
        else:
            self._pool = redis.ConnectionPool(host=p[0], port=int(p[1]))

    def send(self, m):
        """ Send the message """
        pass
    
    def delete(self, m):
        """ Delete the message from the mailbox now """
        return self.delete_after(m, 0)

    def delete_after(self, m, seconds):
        """ Delete the message from the mailbox after n seconds """
        pass
    
    def new_message(self):
        """ Return a new message """
        pass

    def wait(self):
        """ Wait for new messages in the mailbox """
        pass
