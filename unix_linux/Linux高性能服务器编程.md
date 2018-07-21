# 1章 TCP/IP 协议族详解

linux 使用 /etc/resolv.conf 存放 DNS 服务器的 IP 地址
数据链路层、网络层、传输层在内核实现的，操作系统提供了一组系统调用 api: socket

# 2章 IP 协议详解

- IP 头部：出现在每个 IP 数据报中，用于指定 IP 通信的源IP地址、目的IP地址、指导IP 分片和重组
- IP 数据报的路由和转发

IP协议为上层提供无状态、无连接、不可靠的服务，使用 IP 协议的上层协议（比如TCP）需要自己实现数据确认、超时重传等机制。

# 3章 TCP 协议详解

### 3.1 TCP 服务的特点

面向连接、字节流、可靠传输
字节流：应用程序对数据的发送和接受是没有边界限制的。

TCP: send, recv
UDP: sendto() recvfrom()

### 3.2 TCP 头部信息

RTT: Round Trip Time
可以用 tcpdump 工具来检查

TCP状态转移，当前状态可以用 netstat 命令查看

如果客户端向处于半打开状态的连接写入数据，对方将回应一个复位报文段

# 4 TCP/IP 通信案例:访问 Internet 上的 web 服务

使用了 squid http 代理服务器.
- 正向代理：客户端自己设置代理服务器地址。每次请求发给代理服务器，由代理服务器请求资源
- 反向代理：设置在服务端，用代理服务器接受连接请求，然后请求转发给内部网络上的服务器。所以不同地方 ping 一个域名可能获取到不同的 ip 地址
- 透明代理：只能设置在网关上。可以看做正向代理的一种特殊情况

HTTP 应用层协议，一般使用 TCP 协议作为传输层。 wget 用 squid 代理服务器 tcpdump -X 抓包


# 5 Linux 网络编程基础 API

### 5.1 socket 地址 API
主机字节序：现代 cpu 一次能装载至少4个字节，这 4 个字节在内存中的排列顺序影响它被累加器装载成的整数的值
分为
- 大端字节序(big endian) : 一个整数的高位字节存储在内存的低地址处
- 小端字节序(little endian): 高位字节存储在高地址处

现在pc大多采用小端字节序（也被称为主机字节序）。
不同字节序在网络之间传输会有问题，解决方法是发送端总是转成大端字节序，接收端自行决定是否要转换。大端字节序也因此成为
网络字节序。linux 提供了 4 个函数完成主机字节序（小端）和网络字节序（大端）的转换：
(python socket.htonl)

```c
# include <netinet.h>
unsigned long int htonl(unsigned long int  hostlong)  // host to network long
unsigned short int htons(unsigned short int  hostshort)
unsigned long int ntoh(unsigned long int netlong)
unsigned short int ntohs(unsigned short int netshort)
```

socket 编程接口中表示 socket 地址的结构 sockaddr:

```c
#include <bits/socket.h>
struct sockaddr{
sa_family_t sa_family;   // 地址族 AF_UNIX, AF_INET, AF_INET6
char sa_data[14];  // 存放 socket 地址
}
```

ip 地址转换

```c
#include＜arpa/inet.h＞
in_addr_t inet_addr(const char*strptr);
int inet_aton(const char*cp,struct in_addr*inp);
char*inet_ntoa(struct in_addr in);
```

### 5.2 创建 socket
unix 一切皆文件， socket 就是可读、可写、可控制、可关闭的文件描述符

```c
#include＜sys/types.h＞
#include＜sys/socket.h＞
int socket(int domain,int type,int protocol);
```

### 5.3 命名 socket
将一个 socket 和socket 地址绑定称为给 socket 命名。

```c
#include＜sys/types.h＞
#include＜sys/socket.h＞
int bind(int sockfd,const struct sockaddr*my_addr,socklen_t addrlen);
```

### 5.4 监听 socket
bind 后还需要创建一个监听队列存放待处理的客户端连接

```c
#include＜sys/types.h＞
#include＜sys/socket.h＞
int listen(int sockfd, int backlog);
```

监听队列的长度如果超过backlog，服务器将不受理新的客户连接，客户端也将收到ECONNREFUSED错误信息。在内核版本2.2之前的Linux中，backlog参数是指所有处于半连接状态（SYN_RCVD）和完全连接状态（ESTABLISHED）的socket的上限。但自内核版本2.2之后，它只表示处于完全连接状态的socket的上限，处于半连接状态的socket的上限则由/proc/sys/net/ipv4/tcp_max_syn_backlog内核参数定义。backlog参数的典型值是5。

### 5.5 接受连接

```c
#include＜sys/types.h＞
#include＜sys/socket.h＞
int accept(int sockfd, struct sockaddr *addr, socklen_t *addrlen);
```
accept 成功时返回一个新的连接 socket，其唯一标识了被接受的这个连接，服务器可以通过读写该 socket 来与被接受连接
的对应客户端通信。
注意 accept 只是从监听队列中取出连接，而不论连接处于何种状态，更不关心网络的变化。
我们把执行过listen调用、处于 LISTEN 状态的 socket 称为监听 socket， 而所有处于 ESTABLISHED 状态的 socket 称为
连接socket.

### 5.6 发起连接
服务器通过listen 被动接受连接，客户端需要通过 connect 主动与服务器建立连接

```c
#include＜sys/types.h＞
#include＜sys/socket.h＞
int connect(int sockfd, const struct sockaddr *serv_addr, socklen_t addrlen);
```
一旦成功建立连接，sockfd就唯一地标识了这个连接，客户端就可以通过读写sockfd来与服务器通信。connect失败则返回-1并设置errno。其中两种常见的errno是ECONNREFUSED和ETIMEDOUT，

### 5.7 关闭连接

```c
#include<unistd.h>
int close(int fd);  // fd 引用计数-1
```

有个专门为网络编程设计的函数，无论如何都会立即终止连接

```c
#include＜sys/socket.h＞
int shutdown(int sockfd,int howto)  // SHUT_RD, SHUT_WR, SHUT_RDWR
```

### 5.8 socket 数据读写
对于文件读写操作 read/write 同样适用于socket，不过socket 提供了专门的系统调用

```c
#include＜sys/types.h＞
#include＜sys/socket.h＞
ssize_t recv(int sockfd,void*buf,size_t len,int flags);
ssize_t send(int sockfd,const void*buf,size_t len,int flags);
```

udp 数据包读写系统调用（亦可以用于 stream）:

```c
#include＜sys/types.h＞
#include＜sys/socket.h＞
ssize_t recvfrom(int sockfd,void*buf,size_t len,int flags,struct sockaddr*src_addr,socklen_t*addrlen);
ssize_t sendto(int sockfd,const void*buf,size_t len,int flags,const struct sockaddr*dest_addr,socklen_t addrlen)
```

通用数据读写函数：

```c
#include＜sys/socket.h＞
ssize_t recvmsg(int sockfd,struct msghdr*msg,int flags);
ssize_t sendmsg(int sockfd,struct msghdr*msg,int flags);
```

# 5.9 带外标记
linux 检测到 TCP 紧急标志时，将通知应用程序有带外数据需要接收。内核通知应用程序带外数据到达有两种常见方式：
I/O复用产生的异常事件和 SIGURG 信号

```c
#include＜sys/socket.h＞
int sockatmark(int sockfd);
```
sockatmark判断sockfd是否处于带外标记，即下一个被读取到的数据是否是带外数据。如果是，sockatmark返回1，此时我们就可以利用带MSG_OOB标志的recv调用来接收带外数据。如果不是，则sockatmark返回0。

# 5.10 地址信息函数
同样 python 在 socket 模块可以查到

```c
#include＜sys/socket.h＞
// 获取 sockfd 对应的本端 socket 地址，存储在aaddress 参数指定的内存中
int getsockname(int sockfd,struct sockaddr*address,socklen_t*address_len);
// 获取 sockfd 对应的远端 socket 地址
int getpeername(int sockfd,struct sockaddr*address,socklen_t*address_len);
```

# 5.11 socket 选项
下面两个系统调用则是专门用来读取和设置socket文件描述符属性的方法：

```c
#include＜sys/socket.h＞
int getsockopt(int sockfd,int level,int option_name,void*option_value,socklen_t*restrict option_len);
int setsockopt(int sockfd,int level,int option_name,const void*option_value,socklen_t option_len);
```
常用:
- SO_REUSEADDR: 强制使用被处于 TIME_WAIT 状态的连接占用的 socket 地址
- SO_RCVBUF/SO_SNDBUF: 表示TCP 接收缓冲区和发送缓冲区的大小
- SO_RECLOWAT/SO_SNDLOWAT: 表示TCP 接收缓冲区和发送缓冲区的低水位标记，一般被I/O
  复用系统调用用来判断socket是否可读或者可写（默认1字节）
- SO_LINGER: 控制close系统调用在关闭TCP连接时候的行为

 # 5.12 网络信息 API

```c
#include＜netdb.h＞
// 根据主机名获取主机的完整信息
struct hostent*gethostbyname(const char*name);
// 根据 IP 地址获取主机的完整信息
struct hostent*gethostbyaddr(const void*addr,size_t len,int type);

//getservbyname函数根据名称获取某个服务的完整信息，
struct servent*getservbyname(const char*name,const char*proto);
// getservbyport函数根据端口号获取某个服务的完整信息。它们实际上都是通过读取/etc/services文件来获取服务的信息的
struct servent*getservbyport(int port,const char*proto);

// getaddrinfo函数既能通过主机名获取ip地址，也能通过服务名获取端口号
int getaddrinfo(const char*hostname,const char*service,const struct addrinfo*hints,struct addrinfo**result);

// getnameinfo函数能通过socket地址同时获得以字符串表示的主机名（内部使用的是gethostbyaddr函数）和服务名（内部使用的是getservbyport函数）
int getnameinfo(const struct sockaddr*sockaddr,socklen_t addrlen,char*host,socklen_t hostlen,char*serv,socklen_t servlen,int flags);
```

上述4个函数都是不可重入的，即非线程安全的。 <netdb.h> 给出了可重入版本


# 6 高级 IO 函数

# 6.1 pipe 函数
创建一个管道以实现进程间通信

```c
#include <unistd.h>
int pipe(inf fd[2]);  // f[0] 只能用于从管道读取数据， f[1] 写数据

#include＜sys/types.h＞
#include＜sys/socket.h＞
int socketpair(int domain,int type,int protocol,int fd[2]);
```

# 6.2 dup/dup2 函数
有时我们希望把标准输入重定向到一个文件，或者把标准输出重定向到一个网络连接（比如CGI编程）。这可以通过下面的用于复制文件描述符的dup或dup2函数来实现：
```c
#include＜unistd.h＞
int dup(int file_descriptor);
int dup2(int file_descriptor_one,int file_descriptor_two);
```
dup函数创建一个新的文件描述符，该新文件描述符和原有文件描述符file_descriptor指向相同的文件、管道或者网络连接。并且dup返回的文件描述符总是取系统当前可用的最小整数值。dup2和dup类似，不过它将返回第一个不小于file_descriptor_two的整数值。dup和dup2系统调用失败时返回-1并设置errno。

# 6.3 readv/writev 函数

readv函数将数据从文件描述符读到分散的内存块中，即分散读；writev函数则将多块分散的内存数据一并写入文件描述符中，即集中写。它们的定义如下：

```c
#include＜sys/uio.h＞
ssize_t readv(int fd,const struct iovec*vector,int count)；
ssize_t writev(int fd,const struct iovec*vector,int count);
```

# 6.4 sendfile 函数

sendfile函数在两个文件描述符之间直接传递数据（完全在内核中操作），从而避免了内核缓冲区和用户缓冲区之间的数据拷贝，效率很高，这被称为零拷贝。sendfile函数的定义如下：
```c
#include＜sys/sendfile.h＞
ssize_t sendfile(int out_fd,int in_fd,off_t*offset,size_t count);
```

in_fd参数是待读出内容的文件描述符，out_fd参数是待写入内容的文件描述符。offset参数指定从读入文件流的哪个位置开始读，如果为空，则使用读入文件流默认的起始位置。count参数指定在文件描述符in_fd和out_fd之间传输的字节数。sendfile成功时返回传输的字节数，失败则返回-1并设置errno。该函数的man手册明确指出，in_fd必须是一个支持类似mmap函数的文件描述符，即它必须指向真实的文件，不能是socket和管道；而out_fd则必须是一个socket。由此可见，sendfile几乎是专门为在网络上传输文件而设计的。

# 6.5 mmap/munmap 函数
mmap函数用于申请一段内存空间。我们可以将这段内存作为进程间通信的共享内存，也可以将文件直接映射到其中。munmap函数则释放由mmap创建的这段内存空间。它们的定义如下：

```c
#include＜sys/mman.h＞
void*mmap(void*start,size_t length,int prot,int flags,int fd,off_t offset);
int munmap(void*start,size_t length);
```

# 6.6 splice 函数
splice函数用于在两个文件描述符之间移动数据，也是零拷贝操作。splice函数的定义如下：

```c
#include＜fcntl.h＞
ssize_t splice(int fd_in,loff_t*off_in,int fd_out,loff_t*off_out,size_t len,unsigned int flags);
```

# 6.7 tee
tee函数在两个管道文件描述符之间复制数据，也是零拷贝操作。它不消耗数据，因此源文件描述符上的数据仍然可以用于后续的读操作。tee函数的原型如下：

```c
#include＜fcntl.h＞
ssize_t tee(int fd_in,int fd_out,size_t len,unsigned int flags);
```

# 6.8 fcntl 函数 (file control)

```c
#include＜fcntl.h＞
int fcntl(int fd,int cmd,…);
```
