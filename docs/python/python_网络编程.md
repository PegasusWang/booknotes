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
