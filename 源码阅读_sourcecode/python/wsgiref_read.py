# -*- coding: utf-8 -*-

"""
阅读内置的 wsgi server 的实现，了解 http web server 的实现原理，并自己仿照实现简单的 wsgi server


知识点：
- use python3
- tcp socket 编程
- python3 的 selectors 模块，封装了多路复用。This module allows high-level and efficient I/O multiplexing, built upon the select module primitives.
- http 协议
- WSGI 规范
"""


"""
selectors 模块:

BaseSelector   # 抽象基类
+-- SelectSelector
+-- PollSelector
+-- EpollSelector
+-- DevpollSelector
+-- KqueueSelector

BaseSelector: 有两个主要的抽象方法

abstractmethod register(fileobj, events, data=None)
    Register a file object for selection, monitoring it for I/O events.
    fileobj is the file object to monitor. It may either be an integer file descriptor or an object with a fileno() method. events is a bitwise mask of events to monitor. data is an opaque object.
    This returns a new SelectorKey instance, or raises a ValueError in case of invalid event mask or file descriptor, or KeyError if the file object is already registered.

abstractmethod unregister(fileobj):
    Unregister a file object from selection, removing it from monitoring. A file object shall be unregistered prior to being closed.
    fileobj must be a file object previously registered.
    This returns the associated SelectorKey instance, or raises a KeyError if fileobj is not registered. It will raise ValueError if fileobj is invalid (e.g. it has no fileno() method or its fileno() method has an invalid return value).

官方文档给了一个例子：

```py
import selectors
import socket

sel = selectors.DefaultSelector()

def accept(sock, mask):
    conn, addr = sock.accept()  # Should be ready
    print('accepted', conn, 'from', addr)
    conn.setblocking(False)
    sel.register(conn, selectors.EVENT_READ, read)

def read(conn, mask):
    data = conn.recv(1000)  # Should be ready
    if data:
        print('echoing', repr(data), 'to', conn)
        conn.send(data)  # Hope it won't block
    else:
        print('closing', conn)
        sel.unregister(conn)
        conn.close()

sock = socket.socket()
sock.bind(('localhost', 1234))
sock.listen(100)
sock.setblocking(False)
sel.register(sock, selectors.EVENT_READ, accept)

while True:
    events = sel.select()
    for key, mask in events:
        callback = key.data
        callback(key.fileobj, mask)
```
"""


import selectors
import threading


def application(environ, start_response):
    """一个简单的 wsgi 规范的 app 函数，编写 wsgi server 跑起来它"""
    status = '200 OK'
    headers = [('Content-Type', 'text/html; charset=utf8')]

    start_response(status, headers)
    return [b"<h1>Hello</h1>"]


"""
先编写 BaseServer，之后我们会编写 TCPServer 继承它
"""

_ServerSelector = selectors.PollSelector


class BaseServer:
    timeout = None

    def __init__(self, server_address, RequestHandlerClass):
        self.server_addres = server_address
        self.RequestHandlerClass = RequestHandlerClass
        # Class implementing event objects. An event manages a flag that can be set to true with the set() method and reset to false with the clear() method. The wait() method blocks until the flag is true. The flag is initially false.
        self.__is_shut_down = threading.Event()
        self.__shutdown_request = False

    def server_forever(self, poll_interval=0.5):
        """ 循环并接受 socket 请求数据处理请求"""
        self.__is_shut_down.clear()   # flag 设置为 False，其他线程调用wait 被 block，直到被某个线程 set
        try:
            with _ServerSelector as selector:   # 先去看看 python3 selectors 模块是如何对 IO多路复用封装的，如何使用
                selector.register(self, selectors.EVENT_READ)  # register 接受整数文件描述符或者 有 fileno() 方法的对象
                while not self.__shutdown_request:
                    # Wait until some registered file objects become ready, or the timeout expires., This returns a list of (key, events) tuples, one for each ready file object.
                    ready = selector.select(poll_interval)
                    if ready:
                        self._handle_request_noblock()

        finally:
            self.__shutdown_request = False
            self.__is_shut_down.set()   # set 之后其他 wait 的线程才能执行

    def shutdown(self):
        self.__shutdown_request = True
        self.__is_shut_down.wait()  # Block until the internal flag is true，这里必须等待 loop 结束 ( __is_shut_down.set)

    # The distinction between handling, getting, processing and finishing a
    # request is fairly arbitrary.  Remember:
    #
    # - handle_request() is the top-level call.  It calls selector.select(),
    #   get_request(), verify_request() and process_request()
    # - get_request() is different for stream or datagram sockets
    # - process_request() is the place that may fork a new process or create a
    #   new thread to finish the request
    # - finish_request() instantiates the request handler class; this
    #   constructor will handle the request all by itself
    def handle_request(self):
        """处理一个 request"""
        timeout = self.socket.gettimeout()
        if timeout is None:
            timeout = self.timeout
        elif self.timeout is not None:
            timeout = min(timeout, self.timeout)
        if timeout is not None:
            deadline = time() + timeout

        # Wait until a request arrives or the timeout expires - the loop is
        # necessary to accommodate early wakeups due to EINTR.
        with _ServerSelector as selector:
            selector.register(self, selectors.EVENT_READ)
            while True:
                ready = selector.select(timeout)
                if ready:
                    return self._handle_request_noblock()
                else:
                    if timeout is not None:  # 处理超时的情况
                        timeout = deadline - time()
                        if timeou < 0:
                            return self.handle_timeout()

    def _handle_request_noblock(self):
        """Handle one request, without blocking. 这个时候 socket 应该是可读的了
        I assume that selector.select() has returned that the socket is
        readable before this function was called, so there should be no risk of
        blocking in get_request().
        """
        try:
            request, client_address = self.get_request()
        except OSError: # This exception is raised when a system function returns a system-related error, including I/O failures such as “file not found” or “disk full” (not for illegal argument types or other incidental errors).
            return
        if self.verify_request(request, client_address):
            try:
                self.process_request(request, client_address)
            except:
                self.handle_error(request, client_address)
                self.shutdown_request(request)

    def handle_timeout(self):
        pass

    def verify_request(self, request, client_address):
        return True

    def process_request(self, request, client_address):
        self.finish_request(request, client_address)
        self.shutdown_request(request)

    def finish_request(self, request, client_address):
        """注意这里实例化了一个  RequestHandlerClass 实例来处理请求，后续我们会实现一个简单的
        RequestHandlerClass
        """
        self.RequestHandlerClass(request, client_address, self)

    def shutdown_request(self, request):
        self.close_request(request)

    def close_request(self, request):
        pass

    def handle_error(self, request, client_address):
        """Handle an error gracefully.  May be overridden.

        The default is to print a traceback and continue.

        """
        print('-'*40)
        print('Exception happened during processing of request from', end=' ')
        print(client_address)
        import traceback
        traceback.print_exc() # XXX But this goes to stderr!
        print('-'*40)
