# -*- coding: utf-8 -*-

"""
阅读内置的 wsgi server 的实现，了解 http web server 的实现原理，并自己仿照实现简单的 wsgi server


知识点：
- use python3
- tcp socket 编程
- python3 的 selectors 模块，封装了多路复用。This module allows high-level and efficient I/O multiplexing, built upon the select module primitives.
- http 协议
- WSGI 规范


selectors 模块:

BaseSelector   # 抽象基类
+-- SelectSelector
+-- PollSelector
+-- EpollSelector
+-- DevpollSelector
+-- KqueueSelector

BaseSelector: 有两个主要的抽象方法

abstractmethod register(fileobj, events, data=None):
    Register a file object for selection, monitoring it for I/O events.
    fileobj is the file object to monitor. It may either be an integer file descriptor or an object with a fileno() method. events is a bitwise mask of events to monitor. data is an opaque object.
    This returns a new SelectorKey instance, or raises a ValueError in case of invalid event mask or file descriptor, or KeyError if the file object is already registered.

abstractmethod unregister(fileobj):
    Unregister a file object from selection, removing it from monitoring. A file object shall be unregistered prior to being closed.
    fileobj must be a file object previously registered.
    This returns the associated SelectorKey instance, or raises a KeyError if fileobj is not registered. It will raise ValueError if fileobj is invalid (e.g. it has no fileno() method or its fileno() method has an invalid return value).

官方文档给了一个例子：


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
        """


import selectors
import socket
import threading
import time
from io import BufferedIOBase


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
        self.server_address = server_address
        self.RequestHandlerClass = RequestHandlerClass
        # Class implementing event objects. An event manages a flag that can be
        # set to true with the set() method and reset to false with the clear()
        # method. The wait() method blocks until the flag is true. The flag is
        # initially false.
        self.__is_shut_down = threading.Event()
        self.__shutdown_request = False

    def server_forever(self, poll_interval=0.5):
        """ 循环并接受 socket 请求数据处理请求"""
        self.__is_shut_down.clear()   # flag 设置为 False，其他线程调用wait 被 block，直到被某个线程 set
        try:
            with _ServerSelector as selector:   # 先去看看 python3 selectors 模块是如何对 IO多路复用封装的，如何使用
                selector.register(self, selectors.EVENT_READ)  # register 接受整数文件描述符或者 有 fileno() 方法的对象
                while not self.__shutdown_request:
                    # Wait until some registered file objects become ready, or the timeout
                    # expires., This returns a list of (key, events) tuples, one for each
                    # ready file object.
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
                ready = selector.select(timeout)  # This returns a list of (key, events) tuples, one for each ready file object. # noqa
                if ready:
                    return self._handle_request_noblock()
                else:
                    if timeout is not None:  # 处理超时的情况
                        timeout = deadline - time()
                        if timeout < 0:
                            return self.handle_timeout()

    def _handle_request_noblock(self):
        """Handle one request, without blocking. 这个时候 socket 应该是可读的了
        I assume that selector.select() has returned that the socket is
        readable before this function was called, so there should be no risk of
        blocking in get_request().
        """
        try:
            request, client_address = self.get_request()
        # This exception is raised when a system function returns a system-related
        # error, including I/O failures such as “file not found” or “disk full”
        # (not for illegal argument types or other incidental errors).
        except OSError:
            return
        if self.verify_request(request, client_address):
            try:
                self.process_request(request, client_address)
            except BaseException:
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
        print('-' * 40)
        print('Exception happened during processing of request from', end=' ')
        print(client_address)
        import traceback
        traceback.print_exc()  # XXX But this goes to stderr!
        print('-' * 40)


"""
实现了 BaseServer 的基础上，继承它来实现 SocketServer，为后续的 HttpServer 打基础
这一章需要 socket 编程的知识，可以看看 python 的 socket 模块。
如果之前对网络编程不熟悉，可以先看看 《unix网络编程》《unix 环境高级编程》涉及到的一些函数，这两本都是砖头书，
可以针对相关章节看看大致有个 socket 编程的概念。
也可以直接看 python 网络编程相关的内容然后, 先从实现个简单的 tcp socket server 和 client 来了解下 socket 编程
"""


class TCPServer(BaseServer):
    address_family = socket.AF_INET  # ipv4

    socket_type = socket.SOCK_STREAM  # tcp

    request_queue_size = 5

    allow_reuse_address = False

    def __init__(self, server_address, RequestHandlerClass, bind_and_active=True):
        """ 继承自 BaseServer，其实整个 TCPServer 都可以看成是对
        socket 操作的封装
        """
        super().__init__(server_address, RequestHandlerClass)  # use py3 super()
        self.socket = socket.socket(self.address_family, self.socket_type)  # 构造 socket 对象
        if bind_and_active:
            try:
                self.server_bind()
                self.server_activate()
            except BaseException:
                self.server_close()
                raise

    def server_bind(self):
        """ 设置 socket 选项，并绑定 地址和端口"""
        if self.allow_reuse_address:
            self.socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        self.socket.bind(self.server_address)
        self.server_address = self.socket.getsockname()  # Return the socket’s own address.

    def server_activate(self):
        self.socket.listen(self.request_queue_size)

    def server_close(self):
        self.socket.close()

    def fileno(self):  # file of descriptor
        """返回socket 文件描述符"""
        return self.socket.fileno()  # 注意 selector 的 register 的方法需要传入 file descriptor 或者有 fileno() 方法

    def get_request(self):
        # 建议先了解 下 python socket 编程的知识
        # The return value is a pair (conn, address) where conn is a new socket
        # object usable to send and receive data on the connection, and address is
        # the address bound to the socket on the other end of the connection.
        return self.socket.accept()  # return (new_request_socket, addr)

    def shutdown_request(self, request):
        """ 关闭单个 request """
        try:
            request.shutdown(socket.SHUT_WR)  # 显示关闭
        except IOError:
            pass
        self.close_request(request)

    def close_request(self, request):
        request.close()


"""
以上我们就编写完成了 TCPServer，其实就是在继承了 BaseServer 的基础上加上了对 socket 操作的一些封装。
接下来让我们利用 threading 模块来编写一个 ThreadingMixIn ，通过让 TCPServer 继承它来支持多线程
"""


class ThreadingMixIn:
    """mixin class 对每个请求开一个新的 线程来处理"""

    daemon_threads = False

    def process_request_thread(self, request, client_address):
        try:
            self.finish_request(request, client_address)
        except Exception:
            self.handle_error(request, client_address)
        finally:
            self.shutdown_request(request)

    def process_request(self, request, client_address):
        t = threading.Thread(target=self.process_request_thread, args=(request, client_address))
        t.daemon = self.daemon_threads
        t.start()


class ThreadingTCPServer(ThreadingMixIn, TCPServer):
    """多继承就能简单实现一个 多线程的 tcp socket 啦"""
    pass


"""
接下开始编写 RequestHandler，之前编写的 TCPServer 负责处理 socket 的连接和关闭
但是对于 socket 建立连接后如何处理我们还需要一个 RequestHandler 来完成
接下来就开始写 RequestHandler 基类
"""


class BaseRequestHandler:

    def __init__(self, request, client_address, server):
        """

        :param request: 对于 TCPServer request 其实是个 accept() 第一个返回值，新的 socket 对象
        :param client_address:
        :param server:
        """
        self.request = request
        self.client_address = client_address
        self.server = server
        try:
            self.setup()
        except BaseException:
            self.finish()

    def setup(self):
        pass

    def handle(self):
        pass

    def finish(self):
        pass


"""
接下来我们编写 StreamRequestHandler 用来对 socket 进行流式处理，优化 IO
"""


class StreamRequestHandler:
    rbufsize = -1
    wbufsize = 0

    timeout = None
    disable_nagle_algorithm = False  # 是否禁用 nagle 算法

    def setup(self):
        self.connection = self.request
        if self.timeout is not None:
            self.connection.settimeout(self.timeout)
        if self.disable_nagle_algorithm:
            self.connection.setsockopt(socket.IPPROTO_TCP, socket.TCP_NODELAY, True)
        # Return a file object associated with the socket. The exact returned type
        # depends on the arguments given to makefile(). These arguments are
        # interpreted the same way as by the built-in open() function, except the
        # only supported mode values are 'r' (default), 'w' and 'b'.
        self.rfile = self.connection.makefile('rb', self.rbufsize)
        if self.wbufsize == 0:
            self.wfile = _SocketWriter(self.connection)  # 自己实现一个，避免 flush 操作
        else:
            self.wfile = self.connection.makefile('wb', self.wbufsize)

    def finish(self):
        if not self.wfile.closed:
            try:
                self.wfile.flush()
            except socket.error:
                pass
        self.wfile.close()
        self.rfile.close()


class _SocketWriter(BufferedIOBase):

    def __init__(self, sock):
        self._sock = sock

    def writable(self):
        return True

    def write(self, b):
        self._sock.sendall(b)
        # 参考 python3 的built-in types: memoryview objects allow Python code to
        # access the internal data of an object that supports the buffer protocol
        # without copying.
        with memoryview(b) as view:
            return view.nbytes

    def fileno(self):
        return self._sock.fileno()


"""
然后开始实现 http server，在之前的 TCPServer 的基础上
"""


class HTTPServer(TCPServer):
    allow_reuse_address = True

    def server_bind(self):
        super().bind()
        host, port = self.server_address[:2]
        self.server_name = socket.getfqdn(host)
        self.server_port = port


"""
然后是编写 BaseHttpRequestHandler，这里你要了解 http 协议的知识
"""
import sys


class BaseHttpRequestHandler(StreamRequestHandler):
    """
    http(超文本传输协议，HyperText Transfer Protocol) 协议构建在 tcp/ip 基础上，


    1. One line identifying the request type and path
    2. An optional set of RFC-822-style headers
    3. An optional data part

    我建议你使用 curl 配合 -v 参数观察一下 http 请求的过程，比如：
    curl -v http://www.baidu.com > output.txt 2>&1
    你就能看到请求头了。注意这里(2>&1) 我重定向了错误输出，具体原因在以下连接：
    (https://unix.stackexchange.com/questions/259367/copy-output-of-curl-to-file
    curl prints its status to stderr rather than stdout. To capture stderr in the same file, you need to redirect stderr to stdout by adding 2>&1 AFTER your stdout redirection:
    )
    得到的结果如下：

        * Rebuilt URL to: http://baidu.com/
    % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                    Dload  Upload   Total   Spent    Left  Speed

    0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0*   Trying 123.125.115.110...
    * TCP_NODELAY set
    * Connected to baidu.com (123.125.115.110) port 80 (#0)
    > GET / HTTP/1.1
    > Host: baidu.com
    > User-Agent: curl/7.54.0
    > Accept: */*
    >
    < HTTP/1.1 200 OK
    < Date: Sun, 08 Jul 2018 09:56:18 GMT
    < Server: Apache
    < Last-Modified: Tue, 12 Jan 2010 13:48:00 GMT
    < ETag: "51-47cf7e6ee8400"
    < Accept-Ranges: bytes
    < Content-Length: 81
    < Cache-Control: max-age=86400
    < Expires: Mon, 09 Jul 2018 09:56:18 GMT
    < Connection: Keep-Alive
    < Content-Type: text/html
    <
    { [81 bytes data]

    100    81  100    81    0     0   1677      0 --:--:-- --:--:-- --:--:--  1687
    * Connection #0 to host baidu.com left intact
    <html>
    <meta http-equiv="refresh" content="0;url=http://www.baidu.com/">
    </html>

    编写这个类比较麻烦的就是解析 http 协议的函数，这里我们直接拷贝下原来的代码
    """
    sys_version = "Python/" + sys.version.split()[0]

    server_version = "BaseHTTP/" + "0.2"

    error_message_format = """\
    <!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN"
            "http://www.w3.org/TR/html4/strict.dtd">
    <html>
        <head>
            <meta http-equiv="Content-Type" content="text/html;charset=utf-8">
            <title>Error response</title>
        </head>
        <body>
            <h1>Error response</h1>
            <p>Error code: %(code)d</p>
            <p>Message: %(message)s.</p>
            <p>Error code explanation: %(code)s - %(explain)s.</p>
        </body>
    </html>
    """

    error_content_type = "text/html;charset=utf-8"
    default_request_version = "HTTP/0.9"

    def parse_request(self):
        """ 解析请求，这里就是按照 http 规定的格式解析 http 协议 """
        self.command = None
        self.request_version = version = self.default_request_version
        self.close_connection = True
        requestline = str(self.raw_requestline, 'iso-8859-1')
        requestline = requestline.rstrip('\r\n')  # http 是 '\r\n' 分割的
        self.request = requestline
        words = self.requestline.split()
        if len(words) == 3:
            command, path, version = words  # 看上边 curl -v 得到的请求头格式就能理解了
            try:  # 以下都是处理 http 不同吧版本细节的 ，可以先忽略
                if version[:5] != 'HTTP/':
                    raise ValueError
                base_version_number = version.split('/', 1)[1]
                version_number = base_version_number.split(".")
                # RFC 2145 section 3.1 says there can be only one "." and
                #   - major and minor numbers MUST be treated as
                #      separate integers;
                #   - HTTP/2.4 is a lower version than HTTP/2.13, which in
                #      turn is lower than HTTP/12.3;
                #   - Leading zeros MUST be ignored by recipients.
                if len(version_number) != 2:
                    raise ValueError
                version_number = int(version_number[0]), int(version_number[1])
            except (ValueError, IndexError):
                self.send_error( HTTPStatus.BAD_REQUEST, "Bad request version (%r)" % version)
                return False
            if version_number >= (1, 1) and self.protocol_version >= "HTTP/1.1":
                self.close_connection = False
            if version_number >= (2, 0):
                self.send_error(
                    HTTPStatus.HTTP_VERSION_NOT_SUPPORTED,
                    "Invalid HTTP version (%s)" % base_version_number)
                return False
        elif len(words) == 2:
            command, path = words
            self.close_connection = True
            if command != 'GET':
                self.send_error( HTTPStatus.BAD_REQUEST, "Bad HTTP/0.9 request type (%r)" % command)
                return False
        elif not words:
            return False
        else:
            self.send_error(HTTPStatus.BAD_REQUEST, "Bad request syntax (%r)" % requestline)
            return False
        self.command, self.path, self.request_version = command, path, version

        # Examine the headers and look for a Connection directive.
        try:
            self.headers = http.client.parse_headers(self.rfile,
                                                     _class=self.MessageClass)
        except http.client.LineTooLong as err:
            self.send_error(
                HTTPStatus.REQUEST_HEADER_FIELDS_TOO_LARGE,
                "Line too long",
                str(err))
            return False
        except http.client.HTTPException as err:
            self.send_error(
                HTTPStatus.REQUEST_HEADER_FIELDS_TOO_LARGE,
                "Too many headers",
                str(err)
            )
            return False

        conntype = self.headers.get('Connection', "")
        if conntype.lower() == 'close':
            self.close_connection = True
        elif (conntype.lower() == 'keep-alive' and
              self.protocol_version >= "HTTP/1.1"):
            self.close_connection = False
        # Examine the headers and look for an Expect directive
        expect = self.headers.get('Expect', "")
        if (expect.lower() == "100-continue" and
                self.protocol_version >= "HTTP/1.1" and
                self.request_version >= "HTTP/1.1"):
            if not self.handle_expect_100():
                return False
        return True

    def handle_expect_100(self):
        """Decide what to do with an "Expect: 100-continue" header.

        If the client is expecting a 100 Continue response, we must
        respond with either a 100 Continue or a final response before
        waiting for the request body. The default is to always respond
        with a 100 Continue. You can behave differently (for example,
        reject unauthorized requests) by overriding this method.

        This method should either return True (possibly after sending
        a 100 Continue response) or send an error response and return
        False.

        """
        self.send_response_only(HTTPStatus.CONTINUE)
        self.end_headers()
        return True

    def handle_one_request(self):
        """ 刚才处理完并且验证了请求头后，处理单个 http 请求"""
        try:
            self.raw_requestline = self.rfile.readline(65537)
            if len(self.raw_requestline) > 65536:
                self.requestline = ''
                self.request_version = ''
                self.command = ''
                self.send_error(HTTPStatus.REQUEST_URI_TOO_LONG)
                return
            if not self.raw_requestline:
                self.close_connection = True
                return
            if not self.parse_request():
                return
            mname = 'do_' + self.command   # 后续实现 do_GET 等方法
            if not hasattr(self, mname):
                self.send_error( HTTPStatus.NOT_IMPLEMENTED, "Unsupported method (%r)" % self.command)
                return
            method = getattr(self, mname)
            method()
            self.wfile.flush()    # 发送响应体
        except socket.timeout as e:
            self.low_error("Request timeout :%r", e)
            self.close_connection = True
            return

    def handle(self):
        self.close_connection = True
        self.handle_one_request()
        while not self.close_connection:
            self.handle_one_request()

    def send_error(self, code, message=None, explain=None):
        """Send and log an error reply.

        Arguments are
        * code:    an HTTP error code
                   3 digits
        * message: a simple optional 1 line reason phrase.
                   *( HTAB / SP / VCHAR / %x80-FF )
                   defaults to short entry matching the response code
        * explain: a detailed message defaults to the long entry
                   matching the response code.

        This sends an error response (so it must be called before any
        output has been generated), logs the error, and finally sends
        a piece of HTML explaining the error to the user.

        """

        try:
            shortmsg, longmsg = self.responses[code]
        except KeyError:
            shortmsg, longmsg = '???', '???'
        if message is None:
            message = shortmsg
        if explain is None:
            explain = longmsg
        self.log_error("code %d, message %s", code, message)
        self.send_response(code, message)
        self.send_header('Connection', 'close')

        # Message body is omitted for cases described in:
        #  - RFC7230: 3.3. 1xx, 204(No Content), 304(Not Modified)
        #  - RFC7231: 6.3.6. 205(Reset Content)
        body = None
        if (code >= 200 and
            code not in (HTTPStatus.NO_CONTENT,
                         HTTPStatus.RESET_CONTENT,
                         HTTPStatus.NOT_MODIFIED)):
            # HTML encode to prevent Cross Site Scripting attacks
            # (see bug #1100201)
            content = (self.error_message_format % {
                'code': code,
                'message': html.escape(message, quote=False),
                'explain': html.escape(explain, quote=False)
            })
            body = content.encode('UTF-8', 'replace')
            self.send_header("Content-Type", self.error_content_type)
            self.send_header('Content-Length', int(len(body)))
        self.end_headers()

        if self.command != 'HEAD' and body:
            self.wfile.write(body)

    def send_response(self, code, message=None):
        self.log_request(code)
        self.send_response_only(code, message)
        self.send_header('Server', self.version_string())
        self.send_header('Date', self.date_time_string())

    def send_response_only(self, code, message=None):
        """Send the response header only."""
        if self.request_version != 'HTTP/0.9':
            if message is None:
                if code in self.responses:
                    message = self.responses[code][0]
                else:
                    message = ''
            if not hasattr(self, '_headers_buffer'):
                self._headers_buffer = []
            self._headers_buffer.append(("%s %d %s\r\n" %
                    (self.protocol_version, code, message)).encode(
                        'latin-1', 'strict'))

    def send_header(self, keyword, value):
        """Send a MIME header to the headers buffer."""
        if self.request_version != 'HTTP/0.9':
            if not hasattr(self, '_headers_buffer'):
                self._headers_buffer = []  # 这里把所有的请求头放到一个 list buffer 里，统一发送
            self._headers_buffer.append(
                ("%s: %s\r\n" % (keyword, value)).encode('latin-1', 'strict'))

        if keyword.lower() == 'connection':
            if value.lower() == 'close':
                self.close_connection = True
            elif value.lower() == 'keep-alive':
                self.close_connection = False

    def end_headers(self):
        """Send the blank line ending the MIME headers."""
        if self.request_version != 'HTTP/0.9':
            self._headers_buffer.append(b"\r\n")
            self.flush_headers()

    def flush_headers(self):
        if hasattr(self, '_headers_buffer'):
            self.wfile.write(b"".join(self._headers_buffer))  # 发送所有 header buffer 里的数据
            self._headers_buffer = []

    def log_request(self, code='-', size='-'):
        """Log an accepted request.

        This is called by send_response().

        """
        if isinstance(code, HTTPStatus):
            code = code.value
        self.log_message('"%s" %s %s',
                         self.requestline, str(code), str(size))

    def log_error(self, format, *args):
        """Log an error.

        This is called when a request cannot be fulfilled.  By
        default it passes the message on to log_message().

        Arguments are the same as for log_message().

        XXX This should go to the separate error log.

        """

        self.log_message(format, *args)

    def log_message(self, format, *args):
        """Log an arbitrary message.

        This is used by all other logging functions.  Override
        it if you have specific logging wishes.

        The first argument, FORMAT, is a format string for the
        message to be logged.  If the format string contains
        any % escapes requiring parameters, they should be
        specified as subsequent arguments (it's just like
        printf!).

        The client ip and current date/time are prefixed to
        every message.

        """

        sys.stderr.write("%s - - [%s] %s\n" %
                         (self.address_string(),
                          self.log_date_time_string(),
                          format%args))

    def version_string(self):
        """Return the server software version string."""
        return self.server_version + ' ' + self.sys_version

    def date_time_string(self, timestamp=None):
        """Return the current date and time formatted for a message header."""
        if timestamp is None:
            timestamp = time.time()
        return email.utils.formatdate(timestamp, usegmt=True)

    def log_date_time_string(self):
        """Return the current time formatted for logging."""
        now = time.time()
        year, month, day, hh, mm, ss, x, y, z = time.localtime(now)
        s = "%02d/%3s/%04d %02d:%02d:%02d" % (
                day, self.monthname[month], year, hh, mm, ss)
        return s

    weekdayname = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']

    monthname = [None,
                 'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
                 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']

    def address_string(self):
        """Return the client address."""

        return self.client_address[0]


    # Essentially static class variables

    # The version of the HTTP protocol we support.
    # Set this to HTTP/1.1 to enable automatic keepalive
    protocol_version = "HTTP/1.0"

    # MessageClass used to parse headers
    MessageClass = http.client.HTTPMessage

    # hack to maintain backwards compatibility
    responses = {
        v: (v.phrase, v.description)
        for v in HTTPStatus.__members__.values()
    }
