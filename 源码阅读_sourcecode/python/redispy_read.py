#!/usr/bin/env python
# -*- coding:utf-8 -*-

"""
先看代码了解关键部分的实现：socket 连接池，send/recv 数据如何实现(socket编程)，协议如何解析(resp)
看代码自顶向下先有个大致概念，借助如下工具：

- 日志
- 开发工具
- 断点调试工具
- 流程图(一个调用流程)
- 思维导图等

仿写一个 redispy 客户端了解它是如何工作的
如何仿写：

包含哪些部分：

连接：Connection: 负责 socket 建立和关闭
解析：Parser:  解析 redis 协议
Client: 执行命令

知识点：
- socket 编程
- redis 协议，发送和解析


疑问：？
线程安全的吗？

redis protocol: 看下 redis 官方文档关于返回协议格式的部分

In RESP, the type of some data depends on the first byte:
For Simple Strings the first byte of the reply is "+"
For Errors the first byte of the reply is "-"
For Integers the first byte of the reply is ":"
For Bulk Strings the first byte of the reply is "$"
For Arrays the first byte of the reply is "*"
Additionally RESP is able to represent a Null value using a special variation of Bulk Strings or Array as specified later.
In RESP different parts of the protocol are always terminated with "\r\n" (CRLF).

Sending commands to a Redis Server:
A client sends to the Redis server a RESP Array consisting of just Bulk Strings.
A Redis server replies to clients sending any valid RESP data type as reply.


Pipeline:
A Request/Response server can be implemented so that it is able to process new requests even if the client didn't already read the old responses. This way it is possible to send multiple commands to the server without waiting for the replies at all, and finally read the replies in a single step.
This is called pipelining.

```
$ (printf "PING\r\nPING\r\nPING\r\n"; sleep 1) | nc localhost 6379
+PONG
+PONG
+PONG
```
server 端需要占用内存让命令入队，所以不要一次性发太多命令。
"""

import os
import threading
import socket
from itertools import chain
import sys
from StringIO import StringIO as BytesIO
from itertools import imap
from select import select

SERVER_CLOSED_CONNECTION_ERROR = "Connection closed by server."


if sys.version_info[0] < 3:
    from urlparse import parse_qs, urlparse
    from itertools import imap, izip
    from string import letters as ascii_letters
    from Queue import Queue
    try:
        from cStringIO import StringIO as BytesIO
    except ImportError:
        from StringIO import StringIO as BytesIO

    iteritems = lambda x: x.iteritems()
    iterkeys = lambda x: x.iterkeys()
    itervalues = lambda x: x.itervalues()
    nativestr = lambda x: \
        x if isinstance(x, str) else x.encode('utf-8', 'replace')
    u = lambda x: x.decode()
    b = lambda x: x
    next = lambda x: x.next()
    byte_to_chr = lambda x: x
    unichr = unichr
    xrange = xrange
    basestring = basestring
    unicode = unicode
    bytes = str
    long = long
else:
    from urllib.parse import parse_qs, urlparse
    from io import BytesIO
    from string import ascii_letters
    from queue import Queue

    iteritems = lambda x: iter(x.items())
    iterkeys = lambda x: iter(x.keys())
    itervalues = lambda x: iter(x.values())
    byte_to_chr = lambda x: chr(x)
    nativestr = lambda x: \
        x if isinstance(x, str) else x.decode('utf-8', 'replace')
    u = lambda x: x
    b = lambda x: x.encode('latin-1') if not isinstance(x, bytes) else x
    next = next
    unichr = chr
    imap = map
    izip = zip
    xrange = range
    basestring = str
    unicode = str
    bytes = bytes
    long = int

try:  # Python 3
    from queue import LifoQueue, Empty, Full
except ImportError:
    from Queue import Empty, Full
    try:  # Python 2.6 - 2.7
        from Queue import LifoQueue
    except ImportError:  # Python 2.5
        from Queue import Queue
        # From the Python 2.7 lib. Python 2.5 already extracted the core
        # methods to aid implementating different queue organisations.

        class LifoQueue(Queue):
            "Override queue methods to implement a last-in first-out queue."

            def _init(self, maxsize):
                self.maxsize = maxsize
                self.queue = []

            def _qsize(self, len=len):
                return len(self.queue)

            def _put(self, item):
                self.queue.append(item)

            def _get(self):
                return self.queue.pop()


class RedisError(Exception):
    pass


class ResponseError(RedisError):
    pass


class AuthenticationError(RedisError):
    pass


class ConnectionError(RedisError):
    pass


class TimeoutError(RedisError):
    pass


class BaseParser(RedisError):
    pass


class InvalidResponse(RedisError):
    pass


class SocketBuffer(object):
    def __init__(self, socket, socket_read_size):
        self._sock = socket
        self.socket_read_size = socket_read_size
        self._buffer = BytesIO()
        # number of bytes written to the buffer from the socket
        self.bytes_written = 0
        # number of bytes read from the buffer
        self.bytes_read = 0

    @property
    def length(self):
        return self.bytes_written - self.bytes_read

    def _read_from_socket(self, length=None):
        socket_read_size = self.socket_read_size
        buf = self._buffer
        buf.seek(self.bytes_written)
        marker = 0

        try:
            while True:
                data = self._sock.recv(socket_read_size)
                if isinstance(data, bytes) and len(data) == 0:
                    raise socket.error(SERVER_CLOSED_CONNECTION_ERROR)
                buf.write(data)
                data_length = len(data)
                self.bytes_written += data_length
                marker += data_length

                if length is not None and length > marker:
                    continue
                break
        except socket.timeout:
            raise TimeoutError("Timeout reading from socket")
        except socket.error:
            e = sys.exc_info()[1]
            raise ConnectionError("Error while reading from socket: %s" %
                                  (e.args,))

    def read(self, length):
        length = length + 2  # make sure to read the \r\n terminator
        # make sure we've read enough data from the socket
        if length > self.length:
            self._read_from_socket(length - self.length)

        self._buffer.seek(self.bytes_read)
        data = self._buffer.read(length)
        self.bytes_read += len(data)

        # purge the buffer when we've consumed it all so it doesn't
        # grow forever
        if self.bytes_read == self.bytes_written:
            self.purge()

        return data[:-2]

    def readline(self):
        buf = self._buffer
        buf.seek(self.bytes_read)
        data = buf.readline()
        while not data.endswith(SYM_CRLF):
            # there's more data in the socket that we need
            self._read_from_socket()
            buf.seek(self.bytes_read)
            data = buf.readline()
        self.bytes_read += len(data)
        # purge the buffer when we've consumed it all so it doesn't
        # grow forever
        if self.bytes_read == self.bytes_written:
            self.purge()

        return data[:-2]

    def purge(self):
        self._buffer.seek(0)
        self._buffer.truncate()
        self.bytes_written = 0
        self.bytes_read = 0

    def close(self):
        self.purge()
        self._buffer.close()
        self._buffer = None
        self._sock = None


class PythonParser(BaseParser):
    encoding=None

    def __init__(self, socket_read_size):
        self.socket_read_size = socket_read_size
        self._sock = None
        self._buffer = None

    def __del__(self):
        try:
            self.on_disconnect()
        except Exception:
            pass

    def on_connect(self, connection):
        """on_connect

        :param connection: Connection object
        """
        self._sock = connection._sock
        self._buffer = SocketBuffer(self._sock, self.socket_read_size)

    def on_disconnect(self):
        if self._sock is not None:
            self._sock.close()
            self._sock = None
        if self._buffer is not None:
            self._buffer.close()
            self._buffer = None

    def can_read(self):
        return self._buffer and bool(self._buffer.length)

    def read_response(self):

        response = self._buffer.readline()
        if not response:
            raise ConnectionError(SERVER_CLOSED_CONNECTION_ERROR)

        byte, response = byte_to_chr(response[0]), response[1:]

        if byte not in ('-', '+', ':', '$', '*'):
            raise InvalidResponse("Protocol Error: %s, %s" %
                                  (str(byte), str(response)))

        # server returned an error, "-Error message\r\n"
        if byte == '-':
            response = nativestr(response)
            error = self.parse_error(response)
            # if the error is a ConnectionError, raise immediately so the user
            # is notified
            if isinstance(error, ConnectionError):
                raise error
            # otherwise, we're dealing with a ResponseError that might belong
            # inside a pipeline response. the connection's read_response()
            # and/or the pipeline's execute() will raise this error if
            # necessary, so just return the exception instance here.
            return error
        # single value, Simple Strings are used to transmit non binary safe strings with minimal overhead.
        elif byte == '+': # "+OK\r\n"
            pass
        # int value, ":1000\r\n", is guaranteed to be in the range of a signed 64 bit integer.
        elif byte == ':':
            response = long(response)
        # bulk response, represent a single binary safe string up to 512 MB in length.
        elif byte == '$':  # "$6\r\nfoobar\r\n"
            length = int(response)
            if length == -1:  #  Null Bulk String.represent a Null value. "$-1\r\n"
                return None
            response = self._buffer.read(length)
        # multi-bulk response
        elif byte == '*':
            # A * character as the first byte, followed by the number of elements in the array as a decimal number, followed by CRLF.
            # An additional RESP type for every element of the Array.
            length = int(response)
            if length == -1:
                return None
            response = [self.read_response() for i in xrange(length)]   # 递归调用
        if isinstance(response, bytes) and self.encoding:
            response = response.decode(self.encoding)
        return response

class Token(object):
    """
    Literal strings in Redis commands, such as the command names and any
    hard-coded arguments are wrapped in this class so we know not to apply
    and encoding rules on them.
    """

    def __init__(self, value):
        if isinstance(value, Token):
            value = value.value
        self.value = value

    def __repr__(self):
        return self.value

    def __str__(self):
        return self.value


def b(x): return x


SYM_STAR = b('*')
SYM_DOLLAR = b('$')
SYM_CRLF = b('\r\n')
SYM_EMPTY = b('')


class Connection(object):
    """管理到 redis server 的 TCP 连接"""

    def __init__(self, host='localhost', port=6379, db=0, password=None,
                 encoding='utf8', parser_class=PythonParser,
                 decode_responses=False,
                 socket_keepalive=False,
                 socket_timeout=None,
                 socket_connect_timeout=None,
                 socket_read_size=65535):
        self.pid = os.getpid()
        self.host = host
        self.port = port
        self.db = db
        self.password = password
        self.encoding = encoding
        self.decode_responses = decode_responses
        self.socket_timeout = socket_timeout
        self.socket_keepalive = socket_keepalive
        self.socket_connect_timeout = socket_connect_timeout
        self.socket_read_size = socket_read_size
        self._parser = parser_class(socket_read_size)
        self._sock = None

        self._connect_callbacks = []

    def __del__(self):
        try:
            self.disconnect()
        except Exception:
            pass

    def connect(self):
        """如果没有连接到 redis server 就连接上"""
        if self._sock:
            return
        try:
            sock = self._connect()
        except socket.error:
            e = sys.exc_info()[1]
            raise RedisError(e)
        self._sock = sock

        try:
            self.on_connect()
        except RedisError:
            self.disconnect()
            raise

        for callback in self._connect_callbacks:
            callback(self)

    def _connect(self):
        """ 创建 tcp 连接 """
        for res in socket.getaddrinfo(self.host, self.port, 0, socket.SOCK_STREAM):
            family, socktype, proto, canonname, socket_address = res
            sock = None
            try:
                sock = socket.socket(family, socktype, proto)
                # TCP_NODELAY
                sock.setsockopt(socket.IPPROTO_TCP, socket.TCP_NODELAY, 1)

                # TCP_KEEPALIVE
                if self.socket_keepalive:
                    sock.setsockopt(socket.SOL_SOCKET, socket.SO_KEEPALIVE, 1)

                sock.connect(socket_address)
                sock.settimeout(self.socket_timeout)
                return sock
            except socket.error as _:
                err = _
                if sock is not None:
                    sock.close()
        if err is not None:
            raise err
        raise socket.error("socket.getaddrinfo returned an empty list")

    def on_connect(self):
        """初始化连接，认证和选择数据库"""
        self._parser.on_connect(self)

        if self.password:
            self.send_command('AUTH', self.password)
            if self.read_response() != 'OK':
                raise AuthenticationError('Invalid Password')

        if self.db:
            self.send_command('SELECT', self.db)
            if self.read_response() != 'OK':
                raise ConnectionError('Invalid database')

    def disconnect(self):
        self._parser.on_disconnect()
        if self._sock is None:
            return

        try:
            self._sock.shutdown(socket.SHUT_RDWR)
            self._sock.close()
        except socket.error:
            pass
        self._sock = None

    def send_packed_command(self, command):
        if not self._sock:
            self.connect()

        try:
            if isinstance(command, str):
                command = [command]
            for item in command:
                self._sock.sendall(item)
        except socket.timeout:
            self.disconnect()
            raise TimeoutError("Timeout writing to socket")
        except socket.errir:
            self.disconnect()
        except:
            self.disconnect()
            raise

    def send_command(self, *args):
        self.send_packed_command(self.pack_command(*args))

    def can_read(self):
        "Poll the socket to see if there's data that can be read."
        sock = self._sock
        if not sock:
            self.connect()
            sock = self._sock
        return bool(select([sock], [], [], 0)[0]) or self._parser.can_read()

    def read_response(self):
        "Read the response from a previously sent command"
        try:
            response = self._parser.read_response()
        except:
            self.disconnect()
            raise
        if isinstance(response, ResponseError):
            raise response

        return response

    def encode(self, value):
        "Return a bytestring representation of the value"
        if isinstance(value, Token):
            return b(value.value)
        elif isinstance(value, bytes):
            return value
        elif isinstance(value, (int, long)):
            value = b(str(value))
        elif isinstance(value, float):
            value = b(repr(value))
        elif not isinstance(value, basestring):
            value = str(value)
        if isinstance(value, unicode):
            value = value.encode(self.encoding, self.encoding_errors)
        return value

    def pack_command(self, *args):
        output = []
        command = args[0]
        if ' ' in command:
            args = tuple([Token(s) for s in command.split(' ')]) + args[1:]
        else:
            args = (Token(command),) + args[1:]

        buff = SYM_EMPTY.join(
            (SYM_STAR, b(str(len(args))), SYM_CRLF))

        for arg in imap(self.encode, args):
            # to avoid large string mallocs, chunk the command into the
            # output list if we're sending large values
            buff = SYM_EMPTY.join((buff, SYM_DOLLAR, b(str(len(arg))),
                                   SYM_CRLF, arg, SYM_CRLF))
        output.append(buff)
        return output

    def pack_commands(self, commands):
        "Pack multiple commands into the Redis protocol"
        output = []
        pieces = []
        buffer_length = 0

        for cmd in commands:
            for chunk in self.pack_command(*cmd):
                pieces.append(chunk)
                buffer_length += len(chunk)

            if buffer_length > 6000:
                output.append(SYM_EMPTY.join(pieces))
                buffer_length = 0
                pieces = []

        if pieces:
            output.append(SYM_EMPTY.join(pieces))
        return output


class ConnectionPool(object):
    """实现一个连接池，每次需要的时候从 pool 里获取一个连接发送"""

    def __init__(self, connection_class=Connection, max_connections=None,
                 **connection_kwargs):    # 创建 connection 需要的一些参数

        self.connection_class = connection_class
        self.connection_kwargs = connection_kwargs
        self.max_connections = max_connections or 2 ** 31
        self.reset()

    def reset(self):
        self.pid = os.getpid()
        self._created_connections = 0
        self._available_connections = []  # 为啥不用deque 呢？
        self._in_use_connections = set()
        self._check_lock = threading.Lock()

    def _checkpid(self):
        if self.pid != os.getpid():
            with self._check_lock:
                if self.pid == os.getpid():
                    # another thread already did the work while we waited
                    # on the lock.
                    return
                self.disconnect()
                self.reset()

    def get_connection(self, command_name, *keys, **options):
        """有则取一个连接，木有则创建一个"""
        self._checkpid()
        try:
            c = self._available_connections.pop()
        except IndexError:
            c = self.make_connection()
        self._in_use_connections.add(c)
        return c

    def make_connection(self):
        """ create a new connection """
        if self._created_connections >= self.max_connections:
            raise ConnectionError("too many connections")
        self._created_connections += 1
        return self.connection_class(**self.connection_kwargs)

    def release(self, connection):
        self._checkpid()
        if connection.pid != self.pid:
            return
        self._in_use_connections.remove(connection)
        self._available_connections.append(connection)

    def disconnect(self):
        all_conns = chain(self._available_connections, self._in_use_connections)
        for c in all_conns:
            c.disconnect()


class RedisClient(object):
    """发送命令的 client 客户端"""

    RESPONSE_CALLBACKS = {
    }

    def __init__(self, host='localhost', port=6379,
                 db=0, password=None, socket_timeout=None,
                 connection_pool=None, encoding='utf8', encoding_errors='strict',
                 socket_keepalive=None, socket_keepalive_options=None,
                 socket_connect_timeout=None,
                 retry_on_timeout=False,
                 errors=None,
                 decode_responses=False):

        if not connection_pool:
            kwargs = {
                'db': db,
                'password': password,
                'socket_timeout': socket_timeout,
                'encoding': encoding,
                # 'encoding_errors': encoding_errors,
                'decode_responses': decode_responses,
                # 'retry_on_timeout': retry_on_timeout
            }
            # based on input, setup appropriate connection args
            kwargs.update({
                'host': host,
                'port': port,
                'socket_connect_timeout': socket_connect_timeout,
                'socket_keepalive': socket_keepalive,
                # 'socket_keepalive_options': socket_keepalive_options,
            })
            connection_pool = ConnectionPool(**kwargs)
        self.connection_pool = connection_pool
        self.response_callbacks = self.__class__.RESPONSE_CALLBACKS.copy()

    def set_response_callback(self, command, callback):
        self.response_callbacks[command] = callback

    def execute_command(self, *args, **options):
        pool = self.connection_pool
        command_name = args[0]
        connection = pool.get_connection(command_name, **options)
        try:
            connection.send_command(*args)
            return self.parse_response(connection, command_name, **options)
        except (ConnectionError, TimeoutError) as e:
            connection.disconnect()
            if not connection.retry_on_timeout and isinstance(e, TimeoutError):
                raise
            connection.send_command(*args)
            return self.parse_response(connection, command_name, **options)
        finally:
            pool.release(connection)   # 用完一个就释放出去

    def parse_response(self, connection, command_name, **options):
        response = connection.read_response()
        if command_name in self.response_callbacks:
            return self.response_callbacks[command_name](response, **options)
        return response

    def get(self, name):
        return self.execute_command('GET', name)

    def set(self, name, value, timeout_seconds=None):
        pieces = [name, value]
        if timeout_seconds is not None:
            pieces.append('EX{}'.format(timeout_seconds))
        return self.execute_command('SET', *pieces)

    # 后边就可以开始实现一堆命令了，借助 execute_command


def test_redis():
    c = RedisClient()
    c.set('k', '|||||||||||||||||||||||')
    print(c.get('k'))


if __name__ == '__main__':
    test_redis()
