# 《Python 网络编程》第三版

示例代码：https://github.com/brandon-rhodes/fopnp

# 1 客户端/服务端网络编程简介

从 python requests 模块发送请求，然后演示使用底层的 socket 编程来发送请求。

# 2. UDP

使用 udp 发送广播的例子：

```py
# UDP client and server for broadcast messages on a local LAN

import argparse
import socket

BUFSIZE = 65535


def server(interface, port):
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.bind((interface, port))
    print('Listening for datagrams at {}'.format(sock.getsockname()))
    while True:
        data, address = sock.recvfrom(BUFSIZE)
        text = data.decode('ascii')
        print('The client at {} says: {!r}'.format(address, text))


def client(network, port):
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    # 广播可以一次向子网内所有的主机发送数据包
    sock.setsockopt(socket.SOL_SOCKET, socket.SO_BROADCAST, 1)
    text = 'Broadcast datagram!'
    sock.sendto(text.encode('ascii'), (network, port))


if __name__ == '__main__':
    choices = {'client': client, 'server': server}
    parser = argparse.ArgumentParser(description='Send, receive UDP broadcast')
    parser.add_argument('role', choices=choices, help='which role to take')
    parser.add_argument('host', help='interface the server listens at;'
                        ' network the client sends to')
    parser.add_argument('-p', metavar='port', type=int, default=1060,
                        help='UDP port (default 1060)')
    args = parser.parse_args()
    function = choices[args.role]
    function(args.host, args.p)
```

# 3 TCP

TCP 提供可靠连的基本原理：

- 每个tcp 数据包都有一个序列号，接收方可以用来重组和请求重传
- 通过使用计数器记录发送的字节数
- 初始序号随机选择
- tcp无需等待响应就能一口气发送多个数据包，某一时刻发送方希望同时传输的数据量叫做TCP 窗口（window）大小
- 接收方的TCP 实现可以通过控制发送方的窗口大小来减缓或者暂停连接，这叫做流量控制
- TCP 认为数据包被丢弃了，它会假定网络正在变得拥挤，减少每秒发送的数据量

简单的 TCP 服务端和客户端:

```py
import argparse, socket

def recvall(sock, length):
    data = b''
    while len(data) < length:
        more = sock.recv(length - len(data))  # recv会阻塞直到有数据传输
        if not more:
            raise EOFError('was expecting %d bytes but only received'
                           ' %d bytes before the socket closed'
                           % (length, len(data)))
        data += more
    return data

def server(interface, port):
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    # 地址复用，因为OS网络栈部分谨慎处理连接的关闭
    sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    sock.bind((interface, port))
    # 如果服务器还没有通过accept()调用为某连接创建套接字，该连接就会被压栈等待
    # 但如果栈中等待的连接超过了该参数设置的最大等待数，，操作系统会忽略新的连接请求
    sock.listen(1)
    print('Listening at', sock.getsockname())
    while True:
        print('Waiting to accept a new connection')
        sc, sockname = sock.accept()  # accept 调用后返回一个新的用于客户端处理的 socket
        print('We have accepted a connection from', sockname)
        print('  Socket name:', sc.getsockname())
        print('  Socket peer:', sc.getpeername())
        message = recvall(sc, 16)
        print('  Incoming sixteen-octet message:', repr(message))
        sc.sendall(b'Farewell, client')
        sc.close()
        print('  Reply sent, socket closed')

def client(host, port):
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    # NOTE：这点和 UDP不同，UDP 的connect操作只是对绑定套接字进行了配置，设置了后续send/recv 调用所要使用的默认远程地址
    # TCP 的 connect会进行网络操作，三次握手
    sock.connect((host, port))
    print('Client has been assigned socket name', sock.getsockname())
    sock.sendall(b'Hi there, server')
    reply = recvall(sock, 16)
    print('The server said', repr(reply))
    sock.close()  # 结束 TCP 会话

if __name__ == '__main__':
    choices = {'client': client, 'server': server}
    parser = argparse.ArgumentParser(description='Send and receive over TCP')
    parser.add_argument('role', choices=choices, help='which role to play')
    parser.add_argument('host', help='interface the server listens at;'
                        ' host the client sends to')
    parser.add_argument('-p', metavar='PORT', type=int, default=1060,
                        help='TCP port (default 1060)')
    args = parser.parse_args()
    function = choices[args.role]
    function(args.host, args.p)
```

**像使用文件一样使用 TCP 流**:

python 为每个套接字提供了一个 makefile() 方法，返回一个 python 文件对象，该对象实际会在底层调用 recv/send

# 4 套接字名与 DNS

通过 getaddrinfo 函数连接

```py
import argparse, socket, sys

def connect_to(hostname_or_ip):
    try:
        infolist = socket.getaddrinfo(
            hostname_or_ip, 'www', 0, socket.SOCK_STREAM, 0,
            socket.AI_ADDRCONFIG | socket.AI_V4MAPPED | socket.AI_CANONNAME,
            )
    except socket.gaierror as e:
        print('Name service failure:', e.args[1])
        sys.exit(1)

    info = infolist[0]  # per standard recommendation, try the first one
    socket_args = info[0:3]
    address = info[4]
    s = socket.socket(*socket_args)
    try:
        s.connect(address)
    except socket.error as e:
        print('Network failure:', e.args[1])
    else:
        print('Success: host', info[3], 'is listening on port 80')

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Try connecting to port 80')
    parser.add_argument('hostname', help='hostname that you want to contact')
    connect_to(parser.parse_args().hostname)
```

python 没有内置DNS 工具到标准库。pip install dhspython3


# 5 网络数据与网络错误

struct 把数据和二进制格式转换

```py
>>> import struct
>>> struct.pack('<i', 4253) # 小端法
b'\x9d\x10\x00\x00'
>>> struct.pack('>i', 4253) # 大端法
b'\x00\x00\x10\x9d'
>>> struct.unpack('>i', b'\x00\x00\x10\x9d')
(4253,)
```

封帧(framing)问题: 如何分隔消息，使得消息接收方能够识别消息的开始和结束。
需要考虑的问题：

- 接收方何时停止调用recv()是安全的？
- 整个消息或数据何时才能完整无缺地传达？
- 何时才能讲接收到的消息作为一个整体来解析或者处理

解决方式：

- 模式1：对于一些极简单的网络协议，只涉及到数据的发送而不关注响应。
发送方循环发送数据，直到所有数据被传递给sendall()为止，然后使用close()关闭套接字。
接收方只需要不断调用 recv()直到最后一个recv() 返回空字符串

- 模式2：两个方向都通过流发送消息，首先流在一个方向发送消息然后关闭该方向，接着在另一方向上通过流发送数据。最后关闭套接字。
一定要先完成一个方向上的数据传输，然后反过来在另一个方向上通过流发送数据。否则可能导致客户端和服务端发生死锁。

- 模式3：使用定长消息

- 模式4：通过特殊字符划分消息的边界

- 模式5：每个消息前加上长度作为前缀

- 模式6：发送多个数据块，每个数据块前加上长度作为其前缀。抵达信息结尾时，发送方可以发送一个事先约定好的信号(比如0)

HTTP混合使用了模式4和模式5。用 '\r\n' 和 Content-Length

#### 处理网络异常

针对套接字操作常发生的异常：

- OSError: socket 模块可能抛出的主要错误
- socket.gaierror: getaddrinfo() 无法找到提供的名称或者服务时抛出
- socket.timeout: 为套接字设定超时参数

如何处理异常:

- 抛出更具体的异常
- 捕捉与报告网络异常
    - granular 异常处理程序：针对每个网络调用 try except 然后在 except 从句中打印出简洁的错误信息
    - blanket 异常处理程序: 封装代码片段并且从外部捕获


# 6 TLS/SSL

Pycon2014 The Sorry State of SSL

```py
#!/usr/bin/env python3
# Foundations of Python Network Programming, Third Edition
# https://github.com/brandon-rhodes/fopnp/blob/m/py3/chapter06/test_tls.py
# Attempt a TLS connection and, if successful, report its properties

import argparse, socket, ssl, sys, textwrap
import ctypes
from pprint import pprint

def open_tls(context, address, server=False):
    raw_sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    if server:
        raw_sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        raw_sock.bind(address)
        raw_sock.listen(1)
        say('Interface where we are listening', address)
        raw_client_sock, address = raw_sock.accept()
        say('Client has connected from address', address)
        return context.wrap_socket(raw_client_sock, server_side=True)
    else:
        say('Address we want to talk to', address)
        raw_sock.connect(address)
        return context.wrap_socket(raw_sock)

def describe(ssl_sock, hostname, server=False, debug=False):
    cert = ssl_sock.getpeercert()
    if cert is None:
        say('Peer certificate', 'none')
    else:
        say('Peer certificate', 'provided')
        subject = cert.get('subject', [])
        names = [name for names in subject for (key, name) in names
                 if key == 'commonName']
        if 'subjectAltName' in cert:
            names.extend(name for (key, name) in cert['subjectAltName']
                         if key == 'DNS')

        say('Name(s) on peer certificate', *names or ['none'])
        if (not server) and names:
            try:
                ssl.match_hostname(cert, hostname)
            except ssl.CertificateError as e:
                message = str(e)
            else:
                message = 'Yes'
            say('Whether name(s) match the hostname', message)
        for category, count in sorted(context.cert_store_stats().items()):
            say('Certificates loaded of type {}'.format(category), count)

    try:
        protocol_version = SSL_get_version(ssl_sock)
    except Exception:
        if debug:
            raise
    else:
        say('Protocol version negotiated', protocol_version)

    cipher, version, bits = ssl_sock.cipher()
    compression = ssl_sock.compression()

    say('Cipher chosen for this connection', cipher)
    say('Cipher defined in TLS version', version)
    say('Cipher key has this many bits', bits)
    say('Compression algorithm in use', compression or 'none')

    return cert

class PySSLSocket(ctypes.Structure):
    """The first few fields of a PySSLSocket (see Python's Modules/_ssl.c)."""

    _fields_ = [('ob_refcnt', ctypes.c_ulong), ('ob_type', ctypes.c_void_p),
                ('Socket', ctypes.c_void_p), ('ssl', ctypes.c_void_p)]

def SSL_get_version(ssl_sock):
    """Reach behind the scenes for a socket's TLS protocol version."""

    lib = ctypes.CDLL(ssl._ssl.__file__)
    lib.SSL_get_version.restype = ctypes.c_char_p
    address = id(ssl_sock._sslobj)
    struct = ctypes.cast(address, ctypes.POINTER(PySSLSocket)).contents
    version_bytestring = lib.SSL_get_version(struct.ssl)
    return version_bytestring.decode('ascii')

def lookup(prefix, name):
    if not name.startswith(prefix):
        name = prefix + name
    try:
        return getattr(ssl, name)
    except AttributeError:
        matching_names = (s for s in dir(ssl) if s.startswith(prefix))
        message = 'Error: {!r} is not one of the available names:\n {}'.format(
            name, ' '.join(sorted(matching_names)))
        print(fill(message), file=sys.stderr)
        sys.exit(2)

def say(title, *words):
    print(fill(title.ljust(36, '.') + ' ' + ' '.join(str(w) for w in words)))

def fill(text):
    return textwrap.fill(text, subsequent_indent='    ',
                         break_long_words=False, break_on_hyphens=False)

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Protect a socket with TLS')
    parser.add_argument('host', help='hostname or IP address')
    parser.add_argument('port', type=int, help='TCP port number')
    parser.add_argument('-a', metavar='cafile', default=None,
                        help='authority: path to CA certificate PEM file')
    parser.add_argument('-c', metavar='certfile', default=None,
                        help='path to PEM file with client certificate')
    parser.add_argument('-C', metavar='ciphers', default='ALL',
                        help='list of ciphers, formatted per OpenSSL')
    parser.add_argument('-p', metavar='PROTOCOL', default='SSLv23',
                        help='protocol version (default: "SSLv23")')
    parser.add_argument('-s', metavar='certfile', default=None,
                        help='run as server: path to certificate PEM file')
    parser.add_argument('-d', action='store_true', default=False,
                        help='debug mode: do not hide "ctypes" exceptions')
    parser.add_argument('-v', action='store_true', default=False,
                        help='verbose: print out remote certificate')
    args = parser.parse_args()

    address = (args.host, args.port)
    protocol = lookup('PROTOCOL_', args.p)

    context = ssl.SSLContext(protocol)
    context.set_ciphers(args.C)
    context.check_hostname = False
    if (args.s is not None) and (args.c is not None):
        parser.error('you cannot specify both -c and -s')
    elif args.s is not None:
        context.verify_mode = ssl.CERT_OPTIONAL
        purpose = ssl.Purpose.CLIENT_AUTH
        context.load_cert_chain(args.s)
    else:
        context.verify_mode = ssl.CERT_REQUIRED
        purpose = ssl.Purpose.SERVER_AUTH
        if args.c is not None:
            context.load_cert_chain(args.c)
    if args.a is None:
        context.load_default_certs(purpose)
    else:
        context.load_verify_locations(args.a)

    print()
    ssl_sock = open_tls(context, address, args.s)
    cert = describe(ssl_sock, args.host, args.s, args.d)
    print()
    if args.v:
        pprint(cert)
```


# 7 服务器架构

现代操作系统网络栈的两个特点使得异步模式编写服务器成为了现实：

- 提供系统级调用，支持监听多个客户端套接字（epoll/poll/select）
- 可以将一个套接字设置为非阻塞(nonblocking)，非阻塞套接字在send/recv 系统调用会立刻返回，如果发生延迟，调用方会负责在
稍后客户端准备好继续进行交互时重试。

```py
import asyncio, zen_utils

@asyncio.coroutine
def handle_conversation(reader, writer):
    address = writer.get_extra_info('peername')
    print('Accepted connection from {}'.format(address))
    while True:
        data = b''
        while not data.endswith(b'?'):
            more_data = yield from reader.read(4096)
            if not more_data:
                if data:
                    print('Client {} sent {!r} but then closed'
                          .format(address, data))
                else:
                    print('Client {} closed socket normally'.format(address))
                return
            data += more_data
        answer = zen_utils.get_answer(data)
        writer.write(answer)

if __name__ == '__main__':
    address = zen_utils.parse_command_line('asyncio server using coroutine')
    loop = asyncio.get_event_loop()
    coro = asyncio.start_server(handle_conversation, *address)
    server = loop.run_until_complete(coro)
    print('Listening at {}'.format(address))
    try:
        loop.run_forever()
    finally:
        server.close()
        loop.close()
```

# 8 缓存与消息队列

缓存：redis/memcached, 分片
消息队列: AMQP是常见的跨语言消息队列协议实现之一，许多支持AMQP的开源服务器 RabbitMQ, Apache Qpid等
很多时候用一些第三方库将消息队列的功能封装起来，比如Celery 分布式任务队列
消息队列支持多种拓扑结构：

- 管道，生产者创建消息，然后将消息提交至队列中，消费者从队列中接收消息。
- 发布订阅。
- 请求响应模式，消息需要进行往返


# 9 HTTP 客户端

可以使用 `curl -v http://httpbin.org/headers` 来观察 http 协议通信过程

```sh
> GET /headers HTTP/1.1
> Host: httpbin.org
> User-Agent: curl/7.54.0
> Accept: */*
> 
< HTTP/1.1 200 OK
< Connection: keep-alive
< Server: gunicorn/19.9.0
< Date: Thu, 06 Sep 2018 16:01:32 GMT
< Content-Type: application/json
< Content-Length: 133
< Access-Control-Allow-Origin: *
< Access-Control-Allow-Credentials: true

{
  "headers": {
    "Accept": "*/*",
    "Connection": "close",
    "Host": "httpbin.org",
    "User-Agent": "curl/7.54.0"
  }
}
```

HTTP 协议第一行和头信息都通过表示结束的 CR-LF 进行封帧，这两部分作为一个整体又是通过空行来封帧的。
对消息体进行封帧有三种选择：

- Content-Length 头，值是十进制整数，表示消息体包含的字节数。但是动态生成的没法用
- 头消息中指定 Transfer-Encoding 头，并将其设置为 chunked，分块传输
- 服务器指定 Connection: close，然后发送任意大小的消息体之后关闭 TCP 套接字

requests 库的 Session 对象使用了第三方 urllib3，它会维护一个连接池，保存与最近通信的http服务器的处于打开状态的连接。
在向同一个网站请求其他资源时，就可以自动重用连接池中保存的连接了。


# 10 HTTP 服务器

WSGI: 统一web服务器和应用框架的规范

```py
from pprint import pformat
from wsgiref.simple_server import make_server

def app(environ, start_response):
    headers = {'Content-Type': 'text/plain; charset=utf-8'}
    start_response('200 OK', list(headers.items()))
    yield 'Here is the WSGI environment:\r\n\r\n'.encode('utf-8')
    yield pformat(environ).encode('utf-8')

if __name__ == '__main__':
    httpd = make_server('', 8000, app)
    host, port = httpd.socket.getsockname()
    print('Serving on', host, 'port', port)
    httpd.serve_forever()
```


web服务器和框架区别：
web服务器负责建立并管理监听套接字，运行accept()接收新连接，并解析所有收到的HTTP请求。
web服务器会将符合规范的完整请求传递给web框架或应用程序代码，这一个过程功过调用在web服务器上注册的
WSGI可调用对象实现。


# 11 万维网

使用 urllib.parse 模块解析和构造 url

```py
    from urllib.parse import urlsplit
    u = urlsplit('https://www.google.com/search?q=qpod&btnI=yes')
    print(u)
    #SplitResult(scheme='https', netloc='www.google.com', path='/search', query='q=qpod&btnI=yes', fragment='')
```


# 16 Telnet 和 SSH

SSH(Secure Sheel)
python 社区关于远程命令行自动化的工具：

- Fabric 提供了在脚本中通过SSH 连接服务器的功能
- Ansible 对大量远程机器配置方式的管理
- SaltStack 可以在每台客户端机器上安装自己的代理，而不只是借助  ssh 运行
- pexpect: 对远程命令行提示符的交互进行自动化

Telnet 协议：安全性很差，基本被废弃了。 python telnetlib


SSH/SFTP: python paramiko

```py
import argparse, paramiko, sys

class AllowAnythingPolicy(paramiko.MissingHostKeyPolicy):
    def missing_host_key(self, client, hostname, key):
        return

def main(hostname, username):
    client = paramiko.SSHClient()
    client.set_missing_host_key_policy(AllowAnythingPolicy())
    client.connect(hostname, username=username)  # password='')

    channel = client.invoke_shell()
    stdin = channel.makefile('wb')
    stdout = channel.makefile('rb')

    stdin.write(b'echo Hello, world\rexit\r')
    output = stdout.read()
    client.close()

    sys.stdout.buffer.write(output)

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Connect over SSH')
    parser.add_argument('hostname', help='Remote machine name')
    parser.add_argument('username', help='Username on the remote machine')
    args = parser.parse_args()
    main(args.hostname, args.username)
```


# 18 RPC

远程过程调用(RPC, Remote Procedure Cal)允许使用调用本地API或者本地库的语法来调用另一个进程或远程服务器上的函数
特性：

- 只支持有限的数据传输类型，数字字符串、序列、结构体、关联数组等。
- 只要服务器在运行远程函数时发生异常，就能发出异常通知。
- 许多RPC机制自省功能，允许客户端列出特定的rpc服务器支持的所有调用，还可能会列出每个调用接受的参数
- 任何RPC都提供一种寻址方法，支持用户连接特定的远程API
- 有些RPC支持认证和访问控制

python 有自带的  xml-rpc 和三方的 json-rpc 库。
原生使用 python开发的 rpc 系统： Pyro 和 RPyC
