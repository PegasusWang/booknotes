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


# 7 Linux 服务器程序规范

### 7.1 日志
linux 提供一个守护进程 来处理日志系统 rsyslogd
应用程序使用syslog函数与rsyslogd守护进程通信。syslog函数的定义如下：

```c
#include＜syslog.h＞
void syslog(int priority,const char*message,...);
```

### 7.2 用户信息

用户信息对于服务器程序的安全性来说是很重要的，比如大部分服务器就必须以root身份启动，但不能以root身份运行。下面这一组函数可以获取和设置当前进程的真实用户ID（UID）、有效用户ID（EUID）、真实组ID（GID）和有效组ID（EGID）：

```c
#include＜sys/types.h＞
#include＜unistd.h＞
uid_t getuid();/*获取真实用户ID*/,
uid_t geteuid();/*获取有效用户ID*/
gid_t getgid();/*获取真实组ID*/
gid_t getegid();/*获取有效组ID*/
int setuid(uid_t uid);/*设置真实用户ID*/
int seteuid(uid_t uid);/*设置有效用户ID*/
int setgid(gid_t gid);/*设置真实组ID*/
int setegid(gid_t gid);/*设置有效组ID*/
```

### 7.3 进程间关系
Linux下每个进程都隶属于一个进程组，因此它们除了PID信息外，还有进程组ID（PGID）。我们可以用如下函数来获取指定进程的PGID：

```c
#include＜unistd.h＞
pid_t getpgid(pid_t pid);
```

Linux下每个进程都隶属于一个进程组，因此它们除了PID信息外，还有进程组ID（PGID）。我们可以用如下函数来获取指定进程的PGID：
```c
#include＜unistd.h＞
pid_t getpgid(pid_t pid);
```

### 7.4 系统资源限制
Linux上运行的程序都会受到资源限制的影响，比如物理设备限制（CPU数量、内存数量等）、系统策略限制（CPU时间等），以及具体实现的限制（比如文件名的最大长度）。Linux系统资源限制可以通过如下一对函数来读取和设置：

```c
#include＜sys/resource.h＞
int getrlimit(int resource,struct rlimit*rlim);
int setrlimit(int resource,const struct rlimit*rlim);
```

### 7.5 改变工作目录和根目录

获取进程当前工作目录和改变进程工作目录的函数分别是：

```c
# include＜unistd.h＞
char*getcwd(char*buf,size_t size);
int chdir(const char*path);
```

### 7.6 服务器程序后台化

```c
#include＜unistd.h＞
int daemon(int nochdir,int noclose);
```


# 8 高性能服务器程序框架

- I/O 处理单元: 四种 IO 模型，两种高效事件处理模式
- 逻辑单元: 两种高效的并发模式，以及高效的逻辑处理方式-有限状态机
- 存储单元

### 8.1 服务器模型
- C/S 模型: server 端压力大
- P2P 模型: 网络负载高 (可以看成C/S 模型的扩展)


### 8.2 服务器编程框架

- IO 处理单元：服务器管理客户端连接的框架，通常等待并接受新的客户端连接，接受客户端数据，将服务器响应返回给客户端.
但是数据收发不一定在IO 处理单元，也可能在逻辑单元中执行，取决于事件处理模式。
- 逻辑单元：通常是一个线程或者进程，分析并处理客户端数据，然后把结果传递给 IO 处理单元或者直接发送客户端
- 网络存储单元：数据库、缓存、文件等，非必须的
- 请求队列：各单元之间的通信方式的抽象，通常被实现为池的一部分。对于服务器机群，请求队列是服务器之间预先建立、静态的、永久的
TCP 连接

### 8.3 I/O 模型
socket 创建的时候默认是阻塞的，可以传递参数设置成非阻塞。

针对阻塞I/O执行的系统调用可能因为无法立即完成而被操作系统挂起，直到等待的事件发生为止。比如，客户端通过connect向服务器发起连接时，connect将首先发送同步报文段给服务器，然后等待服务器返回确认报文段。如果服务器的确认报文段没有立即到达客户端，则connect调用将被挂起，直到客户端收到确认报文段并唤醒connect调用。socket的基础API中，可能被阻塞的系统调用包括accept、send、recv和connect。

针对非阻塞I/O执行的系统调用则总是立即返回，而不管事件是否已经发生。如果事件没有立即发生，这些系统调用就返回-1，和出错的情况一样。此时我们必须根据errno来区分这两种情况。对accept、send和recv而言，事件未发生时errno通常被设置成EAGAIN（意为“再来一次”）或者EWOULDBLOCK（意为“期望阻塞”）；对connect而言，errno则被设置成EINPROGRESS

很显然，我们只有在事件已经发生的情况下操作非阻塞I/O（读、写等），才能提高程序的效率。因此，非阻塞I/O通常要和其他I/O通知机制一起使用，比如I/O复用和SIGIO信号。

I/O复用是最常使用的I/O通知机制。它指的是，应用程序通过I/O复用函数向内核注册一组事件，内核通过I/O复用函数把其中就绪的事件通知给应用程序。Linux上常用的I/O复用函数是select、poll和epoll_wait，我们将在第9章详细讨论它们。需要指出的是，I/O复用函数本身是阻塞的，它们能提高程序效率的原因在于它们具有同时监听多个I/O事件的能力。

### 8.4 两种高效的事件处理模式

服务器程序通常需要处理三类事件 I/O 事件、信号及定时事件。
两种高效事件处理模式：

 - Reactor: 同步 I/O 模型通常用于实现 Reactor
 Reactor是这样一种模式，它要求主线程（I/O处理单元，下同）只负责监听文件描述上是否有事件发生，有的话就立即将该事件通知工作线程（逻辑单元，下同）。除此之外，主线程不做任何其他实质性的工作。读写数据，接受新的连接，以及处理客户请求均在工作线程中完成。

使用同步I/O模型（以epoll_wait为例）实现的Reactor模式的工作流程是：

1. 主线程往epoll内核事件表中注册socket上的读就绪事件。
2. 主线程调用epoll_wait等待socket上有数据可读。
3. 当socket上有数据可读时，epoll_wait通知主线程。主线程则将socket可读事件放入请求队列。
4. 睡眠在请求队列上的某个工作线程被唤醒，它从socket读取数据，并处理客户请求，然后往epoll内核事件表中注册该socket上的写就绪事件。
5. 主线程调用epoll_wait等待socket可写。
6. 当socket可写时，epoll_wait通知主线程。主线程将socket可写事件放入请求队列。
7. 睡眠在请求队列上的某个工作线程被唤醒，它往socket上写入服务器处理客户请求的结果。

![](./reactor.png)

 - Proactor: 异步 I/O 模型则用于实现 Proactor 模式。与Reactor模式不同，Proactor模式将所有I/O操作都交给主线程和内核来处理，工作线程仅仅负责业务逻辑。因此，Proactor模式更符合图8-4所描述的服务器编程框架。

使用异步I/O模型（以aio_read和aio_write为例）实现的Proactor模式的工作流程是：

1. 主线程调用aio_read函数向内核注册socket上的读完成事件，并告诉内核用户读缓冲区的位置，以及读操作完成时如何通知应用程序（这里以信号为例，详情请参考sigevent的man手册）。
2. 主线程继续处理其他逻辑。
3. 当socket上的数据被读入用户缓冲区后，内核将向应用程序发送一个信号，以通知应用程序数据已经可用。
4. 应用程序预先定义好的信号处理函数选择一个工作线程来处理客户请求。工作线程处理完客户请求之后，调用aio_write函数向内核注册socket上的写完成事件，并告诉内核用户写缓冲区的位置，以及写操作完成时如何通知应用程序（仍然以信号为例）。
5. 主线程继续处理其他逻辑。
6. 当用户缓冲区的数据被写入socket之后，内核将向应用程序发送一个信号，以通知应用程序数据已经发送完毕。
7. 应用程序预先定义好的信号处理函数选择一个工作线程来做善后处理，比如决定是否关闭socket。

![](./proactor.png)

- 使用同步方式模拟 Proactor 模式。 其原理是：主线程执行数据读写操作，读写完成之后，主线程向工作线程通知这一“完成事件”。那么从工作线程的角度来看，它们就直接获得了数据读写的结果，接下来要做的只是对读写的结果进行逻辑处理。

使用同步I/O模型（仍然以epoll_wait为例）模拟出的Proactor模式的工作流程如下：

1. 主线程往epoll内核事件表中注册socket上的读就绪事件。
2. 主线程调用epoll_wait等待socket上有数据可读。
3. 当socket上有数据可读时，epoll_wait通知主线程。主线程从socket循环读取数据，直到没有更多数据可读，然后将读取到的数据封装成一个请求对象并插入请求队列。
4. 睡眠在请求队列上的某个工作线程被唤醒，它获得请求对象并处理客户请求，然后往epoll内核事件表中注册socket上的写就绪事件。
5. 主线程调用epoll_wait等待socket可写。
6. 当socket可写时，epoll_wait通知主线程。主线程往socket上写入服务器处理客户请求的结果。

![](./epoll_proactor.png)

### 8.5 两种高效的并发模式

- 半同步/半异步(half-sync/half-async): 同步线程用来处理客户逻辑，异步线程处理I/O事件
![](./half-sync_half-reactive.png)

- 领导者/追随者(Leader/Followers):领导者/追随者模式是多个工作线程轮流获得事件源集合，轮流监听、分发并处理事件的一种模式。在任意时间点，程序都仅有一个领导者线程，它负责监听I/O事件。而其他线程则都是追随者，它们休眠在线程池中等待成为新的领导者。当前的领导者如果检测到I/O事件，首先要从线程池中推选出新的领导者线程，然后处理I/O事件。此时，新的领导者等待新的I/O事件，而原来的领导者则处理I/O事件，二者实现了并发。
![](./leader_followers.png)

### 8.6 有限状态机
逻辑单元内部的一种高效编程方法

### 8.7 提高服务器性能的其他建议
- 池(pool):空间换时间，预先创建并且初始化资源。内存池、进程池、线程池、连接池
- 用户复制：减少不必要的数据复制，尤其是在用户代码和内核之间，零拷贝函数。工作进程之间应该考虑共享内存而不是管道或者消息队列.
- 上下文切换和锁：进程或者线程切换开销；共享资源加锁保护导致服务器效率低下（减小锁的粒度）


# 9章 I/O 复用
I/O 复用使程序能同时监听多个文件描述符，但是它本身是阻塞的，并且当多个文件描述符同时就绪的时候，如果不采用额外措施，
程序只能依次处理其中的每一个文件描述符，如果要实现并发就要用多进程或多线程等编程手段

### 9.1 select 系统调用

select系统调用的用途是：在一段指定时间内，监听用户感兴趣的文件描述符上的可读、可写和异常等事件。

```c
#include＜sys/select.h＞
int select(int nfds,fd_set*readfds,fd_set*writefds,fd_set*exceptfds,struct timeval*timeout);
```

哪些情况下文件描述符可以被认为是可读、可写或者出现异常，对于select的使用非常关键。在网络编程中，下列情况下socket可读：

❑socket内核接收缓存区中的字节数大于或等于其低水位标记SO_RCVLOWAT。此时我们可以无阻塞地读该socket，并且读操作返回的字节数大于0。

❑socket通信的对方关闭连接。此时对该socket的读操作将返回0。

❑监听socket上有新的连接请求。

❑socket上有未处理的错误。此时我们可以使用getsockopt来读取和清除该错误。

下列情况下socket可写：

❑socket内核发送缓存区中的可用字节数大于或等于其低水位标记SO_SNDLOWAT。此时我们可以无阻塞地写该socket，并且写操作返回的字节数大于0。

❑socket的写操作被关闭。对写操作被关闭的socket执行写操作将触发一个SIGPIPE信号。

### 9.2 poll 系统调用
poll系统调用和select类似，也是在指定时间内轮询一定数量的文件描述符，以测试其中是否有就绪者。poll的原型如下：

```c
#include＜poll.h＞
int poll(struct pollfd*fds,nfds_t nfds,int timeout);
struct pollfd
{
  int fd;/*文件描述符*/
  short events;/*注册的事件*/
  short revents;/*实际发生的事件，由内核填充*/
};
```

### 9.3 epoll 系列系统调用

epoll是Linux特有的I/O复用函数。它在实现和使用上与select、poll有很大差异。首先，epoll使用一组函数来完成任务，而不是单个函数。其次，epoll把用户关心的文件描述符上的事件放在内核里的一个事件表中，从而无须像select和poll那样每次调用都要重复传入文件描述符集或事件集。但epoll需要使用一个额外的文件描述符，来唯一标识内核中的这个事件表。这个文件描述符使用如下epoll_create函数来创建：

```c
#include＜sys/epoll.h＞
int epoll_create(int size)
```

size参数现在并不起作用，只是给内核一个提示，告诉它事件表需要多大。该函数返回的文件描述符将用作其他所有epoll系统调用的第一个参数，以指定要访问的内核事件表。下面的函数用来操作epoll的内核事件表：

```c
#include＜sys/epoll.h＞
int epoll_ctl(int epfd,int op,int fd,struct epoll_event*event)
struct epoll_event
{
__uint32_t events;/*epoll事件*/
epoll_data_t data;/*用户数据*/
};
typedef union epoll_data
{
void*ptr;
int fd;
uint32_t u32;
uint64_t u64;
}epoll_data_t;
```

epoll系列系统调用的主要接口是epoll_wait函数。它在一段超时时间内等待一组文件描述符上的事件，其原型如下：

```c
#include＜sys/epoll.h＞
// 该函数成功时返回就绪的文件描述符的个数，失败时返回-1并设置errno。
int epoll_wait(int epfd,struct epoll_event*events,int maxevents,int timeout)
```

epoll对文件描述符的操作有两种模式：LT（Level Trigger，电平触发）模式和ET（Edge Trigger，边沿触发）模式。LT模式是默认的工作模式，这种模式下epoll相当于一个效率较高的poll。当往epoll内核事件表中注册一个文件描述符上的EPOLLET事件时，epoll将以ET模式来操作该文件描述符。ET模式是epoll的高效工作模式。

对于采用LT工作模式的文件描述符，当epoll_wait检测到其上有事件发生并将此事件通知应用程序后，应用程序可以不立即处理该事件。这样，当应用程序下一次调用epoll_wait时，epoll_wait还会再次向应用程序通告此事件，直到该事件被处理。而对于采用ET工作模式的文件描述符，当epoll_wait检测到其上有事件发生并将此事件通知应用程序后，应用程序必须立即处理该事件，因为后续的epoll_wait调用将不再向应用程序通知这一事件。可见，ET模式在很大程度上降低了同一个epoll事件被重复触发的次数，因此效率要比LT模式高。

我们期望的是一个socket连接在任一时刻都只被一个线程处理。这一点可以使用epoll的EPOLLONESHOT事件实现。


### 9.4 三组 I/O 复用函数的比较
![](./io_diff.png)

### 9.5 I/O 复用高级应用一：非阻塞 connect

### 9.8 超级服务 xinetd

Linux因特网服务inetd是超级服务。它同时管理着多个子服务，即监听多个端口。现在Linux系统上使用的inetd服务程序通常是其升级版本xinetd

![](./xinetd.png)


# 10章 信号
### 10.1 Linux 信号概述

Linux下，一个进程给其他进程发送信号的API是kill函数。其定义如下：

```c
#include＜sys/types.h＞
#include＜signal.h＞
int kill(pid_t pid,int sig);
```

目标进程在收到信号时，需要定义一个接收函数来处理之。信号处理函数的原型如下：

```c
#include＜signal.h＞
typedef void(*__sighandler_t)(int);
```
信号处理函数只带有一个整型参数，该参数用来指示信号类型。信号处理函数应该是可重入的，否则很容易引发一些竞态条件。所以在信号处理函数中严禁调用一些不安全的函数。

### 10.2 signal 系统调用

要为一个信号设置处理函数，可以使用下面的signal系统调用：

```c
#include＜signal.h＞
_sighandler_t signal(int　sig,_sighandler_t_handler)
```

设置信号处理函数的更健壮的接口是如下的系统调用：

```c
#include＜signal.h＞
int sigaction(int sig,const struct sigaction*act,struct sigaction*oact);
```

### 10.3 信号集函数

Linux使用数据结构sigset_t来表示一组信号。其定义如下：

```c
#include＜bits/sigset.h＞
#define_SIGSET_NWORDS(1024/(8*sizeof(unsigned long int)))
typedef struct
{
unsigned long int__val[_SIGSET_NWORDS];
}__sigset_t;
```

### 10.4 统一事件源
信号是一种异步事件：信号处理函数和程序的主循环是两条不同的执行路线。很显然，信号处理函数需要尽可能快地执行完毕，以确保该信号不被屏蔽（前面提到过，为了避免一些竞态条件，信号在处理期间，系统不会再次触发它）太久。一种典型的解决方案是：把信号的主要处理逻辑放到程序的主循环中，当信号处理函数被触发时，它只是简单地通知主循环程序接收到信号，并把信号值传递给主循环，主循环再根据接收到的信号值执行目标信号对应的逻辑代码。信号处理函数通常使用管道来将信号“传递”给主循环：信号处理函数往管道的写端写入信号值，主循环则从管道的读端读出该信号值。那么主循环怎么知道管道上何时有数据可读呢?这很简单，我们只需要使用I/O复用系统调用来监听管道的读端文件描述符上的可读事件。如此一来，信号事件就能和其他I/O事件一样被处理，即统一事件源。

很多优秀的I/O框架库和后台服务器程序都统一处理信号和I/O事件，比如Libevent I/O框架库和xinetd超级服务。

### 10.5 网络编程相关信号
- SIGHUP: 当挂起进程的控制终端时， SIGHUP 信号将被触发，对于没有控制终端的网络后台程序而言，它们通常利用该信号来强制服务器重新读取配置文件
- SIGPIPE: 默认往一个读端关闭的管道或者socket连接中写数据将引发 SIGPIPE 信号
- SIGURG: 在Linux环境下，内核通知应用程序带外数据到达主要有两种方法：一种是第9章介绍的I/O复用技术，select等系统调用在接收到带外数据时将返回，并向应用程序报告socket上的异常事件，代码清单9-1给出了一个这方面的例子；另外一种方法就是使用SIGURG信号


# 11章 定时器

两种高效的管理定时器的容器：时间轮和时间堆

### 11.1 socket 选项 SO_RCVTIMEO 和 SO_SNDTIMEO

![](./timeo.png)

在程序中，我们可以根据系统调用（send、sendmsg、recv、recvmsg、accept和connect）的返回值以及errno来判断超时时间是否已到，进而决定是否开始处理定时任务。

### 11.2 SIGALRM 信号
第10章提到，由alarm和setitimer函数设置的实时闹钟一旦超时，将触发SIGALRM信号。因此，我们可以利用该信号的信号处理函数来处理定时任务。但是，如果要处理多个定时任务，我们就需要不断地触发SIGALRM信号，并在其信号处理函数中执行到期的任务。一般而言，SIGALRM信号按照固定的频率生成，即由alarm或setitimer函数设置的定时周期T保持不变。如果某个定时任务的超时时间不是T的整数倍，那么它实际被执行的时间和预期的时间将略有偏差。因此定时周期T反映了定时的精度。

### 11.3 I/O 复用系统调用的超时参数

Linux下的3组I/O复用系统调用都带有超时参数，因此它们不仅能统一处理信号和I/O事件，也能统一处理定时事件。但是由于I/O复用系统调用可能在超时时间到期之前就返回（有I/O事件发生），所以如果我们要利用它们来定时，就需要不断更新定时参数以反映剩余的时间，如代码清单11-4所示。
```c
#define TIMEOUT 5000
int timeout=TIMEOUT;
time_t start=time(NULL);
time_t end=time(NULL);
while(1)
{
    printf("the timeout is now%d mil-seconds\n",timeout);
    start=time(NULL);
    int number=epoll_wait(epollfd,events,MAX_EVENT_NUMBER,timeout);
    if((number＜0)＆＆(errno!=EINTR))
    {
        printf("epoll failure\n");
        break;
    }
    /*如果epoll_wait成功返回0，则说明超时时间到，此时便可处理定时任务，并重置定时时间*/
    if(number==0)
    {
        timeout=TIMEOUT;
        continue;
    }
    end=time(NULL);
    /*如果epoll_wait的返回值大于0，则本次epoll_wait调用持续的时间是(end-start)*1000 ms，我们需要将定时时间timeout减去这段时间，以获得下次epoll_wait调用的超时参数*/
    timeout-=(end-start)*1000;
    /*重新计算之后的timeout值有可能等于0，说明本次epoll_wait调用返回时，不仅有文件描述符就绪，而且其超时时间也刚好到达，此时我们也要处理定时任务，并重置定时时间*/
    if(timeout＜=0)
    {
        timeout=TIMEOUT;
    }
    //handle connections
}
```

### 11.4 高性能定时器
- 时间轮: 指针指向轮子上的一个槽（slot）。它以恒定的速度顺时针转动，每转动一步就指向下一个槽（虚线指针指向的槽），每次转动称为一个滴答（tick）。一个滴答的时间称为时间轮的槽间隔si（slot interval），它实际上就是心搏时间。该时间轮共有N个槽，因此它运转一周的时间是N*si。每个槽指向一条定时
器链表，每条链表上的定时器具有相同的特征：它们的定时时间相差N*si的整数倍。时间轮正是利用这个关系将定时器散列到不同的链表中。假如现在指针指向槽cs，我们要添加一个定时时间为ti的定时器，则该定时器将被插入槽ts（timer
slot）对应的链表中：ts=(cs+(ti/si))%N

![时间轮](./time_wheel.png)

- 时间堆: 设计定时器的另外一种思路是：将所有定时器中超时时间最小的一个定时器的超时值作为心搏间隔。这样，一旦心搏函数tick被调用，超时时间最小的定时器必然到期，我们就可以在tick函数中处理该定时器。然后，再次从剩余的定时器中找出超时时间最小的一个，并将这段最小时间设置为下一次心搏间隔。如此反复，就实现了较为精确的定时。


# 12章 高性能 I/O 框架库 Libevent
linux服务器必须处理三种事件： I/O事件、信号和定时事件，处理这三类事件时通常需要考虑三个问题：
- 统一事件源。利用 I/O 复用系统调用来管理所有事件
- 可移植性， FreeBSD kqueue, Linux epoll
- 并发编程支持。多进程和多线程下，需要考虑各执行实体如何协同处理客户连接、信号和定时器，避免竞态条件


### 12.1 I/O 框架库概述
以 Reactor， Proactor 或者同时以两种模式实现

![I/O框架库组件](./io_library.png)

1.句柄

I/O框架库要处理的对象，即I/O事件、信号和定时事件，统一称为事件源。一个事件源通常和一个句柄绑定在一起。句柄的作用是，当内核检测到就绪事件时，它将通过句柄来通知应用程序这一事件。在Linux环境下，I/O事件对应的句柄是文件描述符，信号事件对应的句柄就是信号值。

2.事件多路分发器

事件的到来是随机的、异步的。我们无法预知程序何时收到一个客户连接请求，又亦或收到一个暂停信号。所以程序需要循环地等待并处理事件，这就是**事件循环**。在事件循环中，等待事件一般使用I/O复用技术来实现。I/O框架库一般将系统支持的各种I/O复用系统调用封装成统一的接口，称为事

件多路分发器。事件多路分发器的demultiplex方法是等待事件的核心函数，其内部调用的是select、poll、epoll_wait等函数。

此外，事件多路分发器还需要实现register_event和remove_event方法，以供调用者往事件多路分发器中添加事件和从事件多路分发器中删除事件。

3.事件处理器和具体事件处理器

事件处理器执行事件对应的业务逻辑。它通常包含一个或多个handle_event回调函数，这些回调函数在事件循环中被执行。I/O框架库提供的事件处理器通常是一个接口，用户需要继承它来实现自己的事件处理器，即具体事件处理器。因此，事件处理器中的回调函数一般被声明为虚函数，以支持用户的扩展。

此外，事件处理器一般还提供一个get_handle方法，它返回与该事件处理器关联的句柄。那么，事件处理器和句柄有什么关系？当事件多路分发器检测到有事件发生时，它是通过句柄来通知应用程序的。因此，我们必须将事件处理器和句柄绑定，才能在事件发生时获取到正确的事件处理器。

4.Reactor

Reactor是I/O框架库的核心。它提供的几个主要方法是：

❑handle_events。该方法执行事件循环。它重复如下过程：等待事件，然后依次处理所有就绪事件对应的事件处理器。

❑register_handler。该方法调用事件多路分发器的register_event方法来往事件多路分发器中注册一个事件。

❑remove_handler。该方法调用事件多路分发器的remove_event方法来删除事件多路分发器中的一个事件。

![](./time_order.png)

### 12.2 Libevent 源码分析
从使用方式入手分析代码：

```c
#include＜sys/signal.h＞
#include＜event.h＞
void signal_cb(int fd,short event,void*argc)  // 回调函数相当于 handle_event，arg 参数是 Reactor 传递给回调函数的参数
{
    struct event_base*base=(event_base*)argc;
    struct timeval delay={2,0};
    printf("Caught an interrupt signal;exiting cleanly in two seconds...\n");
    event_base_loopexit(base,＆delay);
}
void timeout_cb(int fd,short event,void*argc)
{
    printf("timeout\n");
}
int main()
{
    struct event_base*base=event_init();    // 调用event_init 创建 event_base 对象，一个 event_base 相当于一个 Reactor 实例
    struct event*signal_event=evsignal_new(base,SIGINT,signal_cb,base);  // 创建信号事件处理器
    event_add(signal_event,NULL);  // 相当于 register_handler
    timeval tv={1,0};
    struct event*timeout_event=evtimer_new(base,timeout_cb,NULL);  // 创建定时器事件处理器
    event_add(timeout_event,＆tv);
    event_base_dispatch(base);  // 调用 event_base_dispatch 执行事件循环

    event_free(timeout_event);
    event_free(signal_event);
    event_base_free(base);
}
```


# 13章 多进程编程

进程是Linux操作系统环境的基础，它控制着系统上几乎所有的活动。本章从系统程序员的角度来讨论Linux多进程编程，包括如下内容：

❑复制进程映像的fork系统调用和替换进程映像的exec系列系统调用。

❑僵尸进程以及如何避免僵尸进程。

❑进程间通信（Inter-Process Communication，IPC）最简单的方式：管道。

❑3种System V进程间通信方式：信号量、消息队列和共享内存。它们都是由AT＆T System V2版本的UNIX引入的，所以统称为System V IPC。

❑在进程间传递文件描述符的通用方法：通过UNIX本地域socket传递特殊的辅助数据

### 13.1 fork 系统调用

Linux下创建新进程的系统调用是fork

```c
#include＜sys/types.h＞
#include＜unistd.h＞
pid_t fork(void);<Paste>
```

该函数的每次调用都返回两次，在父进程中返回的是子进程的PID，在子进程中则返回0。该返回值是后续代码判断当前进程是父进程还是子进程的依据。

### 13.2 exec 系列系统调用

有时我们需要在子进程中执行其他程序，即替换当前进程映像，这就需要使用如下exec系列函数之一：

```c
#include＜unistd.h＞
extern char**environ;
int execl(const char*path,const char*arg,...);
int execlp(const char*file,const char*arg,...);
int execle(const char*path,const char*arg,...,char*const envp[]);
int execv(const char*path,char*const argv[]);
int execvp(const char*file,char*const argv[]);
int execve(const char*path,char*const argv[],char*const envp[]);
```

### 13.3 处理僵尸进程

如果父进程没有正确地处理子进程的返回信息，子进程都将停留在僵尸态，并占据着内核资源。这是绝对不能容许的，毕竟内核资源有限。下面这对函数在父进程中调用，以等待子进程的结束，并获取子进程的返回信息，从而避免了僵尸进程的产生，或者使子进程的僵尸态立即结束：

```c
#include＜sys/types.h＞
#include＜sys/wait.h＞
pid_t wait(int*stat_loc);
pid_t waitpid(pid_t pid,int*stat_loc,int options); //可以是非阻塞的
```

wait函数将阻塞进程，直到该进程的某个子进程结束运行为止。它返回结束运行的子进程的PID，并将该子进程的退出状态信息存储于stat_loc参数指向的内存中。

当一个进程结束时，它将给其父进程发送一个SIGCHLD信号。因此，我们可以在父进程中捕获SIGCHLD信号，并在信号处理函数中调用waitpid函数以“彻底结束”一个子进程，

```c
static void handle_child(int sig)
{
pid_t pid;
int stat;
while((pid=waitpid(-1,＆stat,WNOHANG))＞0)
  {
  /*对结束的子进程进行善后处理*/
  }
}
```

### 13.4 管道
pipe 也是父子进程之间通信的常用手段

![](./pipe.png)

### 13.5 信号量

当多个进程同时访问系统上的某个资源的时候，比如同时写一个数据库的某条记录，或者同时修改某个文件，就需要考虑进程的同步问题，以确保任一时刻只有一个进程可以拥有对资源的独占式访问。通常，程序对共享资源的访问的代码只是很短的一段，但就是这一段代码引发了进程之间的竞态条件。我们称这段代码为关键代码段，或者**临界区**。对进程同步，也就是确保任一时刻只有一个进程能进入关键代码段。

Dijkstra提出的信号量（Semaphore）概念是并发编程领域迈出的重要一步。信号量是一种特殊的变量，它只能取自然数值并且只支持两种操作：等待（wait）和信号（signal）。不过在Linux/UNIX中，“等待”和“信号”都已经具有特殊的含义，所以对信号量的这两种操作更常用的称呼是P、V操作。这两个字母来自于荷兰语单词passeren（传递，就好像进入临界区）和vrijgeven（释放，就好像退出临界区）。假设有信号量SV，则对它的P、V操作含义如下：

❑P(SV)，如果SV的值大于0，就将它减1；如果SV的值为0，则挂起进程的执行。

❑V(SV)，如果有其他进程因为等待SV而挂起，则唤醒之；如果没有，则将SV加1。

注意　使用一个普通变量来模拟二进制信号量是行不通的，因为所有高级语言都没有一个原子操作可以同时完成如下两步操作：检测变量是否为true/false，如果是则再将它设置为false/true。

semget系统调用创建一个新的信号量集，或者获取一个已经存在的信号量集。其定义如下：

```c
#include＜sys/sem.h＞
int semget(key_t key,int num_sems,int sem_flags);
```

semop系统调用改变信号量的值，即执行P、V操作。

```c
unsigned short semval;
/*信号量的值*/
unsigned short semzcnt;

/*等待信号量值变为0的进程数量*/
unsigned short semncnt;
/*等待信号量值增加的进程数量*/
pid_t sempid;
/*最后一次执行semop操作的进程ID*/

//semop对信号量的操作实际上就是对这些内核变量的操作。
int semop(int sem_id,struct sembuf*sem_ops,size_t num_sem_ops);
```

semctl系统调用允许调用者对信号量进行直接控制。其定义如下：

```c
#include＜sys/sem.h＞
int semctl(int sem_id,int sem_num,int command,...);
```

semget的调用者可以给其key参数传递一个特殊的键值IPC_PRIVATE（其值为0），这样无论该信号量是否已经存在，semget都将创建一个新的信号量。使用该键值创建的信号量并非像它的名字声称的那样是进程私有的。其他进程，尤其是子进程，也有方法来访问这个信号量。所以semget的man手册的BUGS部分上说，使用名字IPC_PRIVATE有些误导（历史原因），应该称为IPC_NEW。

### 13.6 共享内存
共享内存是最高效的 IPC 机制，不涉及进程之间任何数据传递，但需要同步进程对共享内存的访问，否则会产生静态条件

shmget系统调用创建一段新的共享内存，或者获取一段已经存在的共享内存。其定义如下：

```c
#include＜sys/shm.h＞
int shmget(key_t key,size_t size,int shmflg);
```
如果shmget用于创建共享内存，则这段共享内存的所有字节都被初始化为0，与之关联的内核数据结构shmid_ds将被创建并初始化。shmid_ds结构体

```c
struct shmid_ds
{
struct ipc_perm shm_perm;/*共享内存的操作权限*/
size_t shm_segsz;/*共享内存大小，单位是字节*/
__time_t shm_atime;/*对这段内存最后一次调用shmat的时间*/
__time_t shm_dtime;/*对这段内存最后一次调用shmdt的时间*/
__time_t shm_ctime;/*对这段内存最后一次调用shmctl的时间*/
__pid_t shm_cpid;/*创建者的PID*/
__pid_t shm_lpid;/*最后一次执行shmat或shmdt操作的进程的PID*/
shmatt_t shm_nattach;/*目前关联到此共享内存的进程数量*/
/*省略一些填充字段*/
};
```

共享内存被创建/获取之后，我们不能立即访问它，而是需要先将它关联到进程的地址空间中。使用完共享内存之后，我们也需要将它从进程地址空间中分离。这两项任务分别由如下两个系统调用实现：

```c
#include＜sys/shm.h＞
void*shmat(int shm_id,const void*shm_addr,int shmflg);
int shmdt(const void*shm_addr);
```

shmctl系统调用控制共享内存的某些属性。其定义如下：

```c
#include＜sys/shm.h＞
int shmctl(int shm_id,int command,struct shmid_ds*buf);
```

### 13.7 消息队列

消息队列是在两个进程之间传递二进制块数据的一种简单有效的方式。每个数据块都有一个特定的类型，接收方可以根据类型来有选择地接收数据，而不一定像管道和命名管道那样必须以先进先出的方式接收数据。
msgget系统调用创建一个消息队列，或者获取一个已有的消息队列。其定义如下：

```c
#include＜sys/msg.h＞
int msgget(key_t key,int msgflg);
```
如果msgget用于创建消息队列，则与之关联的内核数据结构msqid_ds将被创建并初始化。msqid_ds结构体的定义如下：

```c
struct msqid_ds
{
struct ipc_perm msg_perm;/*消息队列的操作权限*/
time_t msg_stime;/*最后一次调用msgsnd的时间*/
time_t msg_rtime;/*最后一次调用msgrcv的时间*/
time_t msg_ctime;/*最后一次被修改的时间*/
unsigned long__msg_cbytes;/*消息队列中已有的字节数*/
msgqnum_t msg_qnum;/*消息队列中已有的消息数*/
msglen_t msg_qbytes;/*消息队列允许的最大字节数*/
pid_t msg_lspid;/*最后执行msgsnd的进程的PID*/
pid_t msg_lrpid;/*最后执行msgrcv的进程的PID*/
};
```

msgsnd系统调用把一条消息添加到消息队列中。其定义如下：

```c
#include＜sys/msg.h＞
int msgsnd(int msqid,const void*msg_ptr,size_t msg_sz,int msgflg);
struct msgbuf
{
long mtype;/*消息类型*/
char mtext[512];/*消息数据*/
};
```

msgrcv系统调用从消息队列中获取消息。其定义如下：

```c
#include＜sys/msg·h＞
int msgrcv(int msqid,void*msg_ptr,size_t msg_sz,long int msgtype,int msgflg);
```

msgctl系统调用控制消息队列的某些属性。其定义如下：

```c
#include＜sys/msg.h＞
int msgctl(int msqid,int command,struct msqid_ds*buf);
```

### 13.8 IPC 命令

上述3种System V IPC进程间通信方式都使用一个全局唯一的键值（key）来描述一个共享资源。当程序调用semget、shmget或者msgget时，就创建了这些共享资源的一个实例。Linux提供了ipcs命令，以观察当前系统上拥有哪些共享资源实例。

```c
$sudo ipcs
------Shared Memory Segments--------
key shmid owner perms bytes nattch status
------Semaphore Arrays--------
key semid owner perms nsems
0x00000000 196608 apache 600 1
0x00000000 229377 apache 600 1
0x00000000 262146 apache 600 1
0x00000000 294915 apache 600 1
0x00000000 327684 apache 600 1
0x00000000 360453 apache 600 1
0x00000000 393222 apache 600 1
------Message Queues--------
key msqid owner perms used-bytes messages
```
输出结果分段显示了系统拥有的共享内存、信号量和消息队列资源。可见，该系统目前尚未使用任何共享内存和消息队列，却分配了一组键值为0（IPC_PRIVATE）的信号量。这些信号量的所有者是apache，因此它们是由httpd服务器程序创建的。其中标识符为393222的信号量正是我们在13.5.5小节讨论的那个用于在httpd各个子进程之间同步epoll_wait使用权的信号量。
，我们可以使用ipcrm命令来删除遗留在系统中的共享资源。

### 13.9 在进程间传递文件描述符

由于fork调用之后，父进程中打开的文件描述符在子进程中仍然保持打开，所以文件描述符可以很方便地从父进程传递到子进程。需要注意的是，传递一个文件描述符并不是传递一个文件描述符的值，而是要在接收进程中创建一个新的文件描述符，并且该文件描述符和发送进程中被传递的文件描述符指向内核中相同的文件表项。

那么如何把子进程中打开的文件描述符传递给父进程呢？或者更通俗地说，如何在两个不相干的进程之间传递文件描述符呢？在Linux下，我们可以利用UNIX域socket在进程间传递特殊的辅助数据，以实现文件描述符的传递[2]。


# 14章 多线程编程

### 14.1 Linux 线程概述

线程是程序中完成一个独立任务的完整执行序列，即一个可调度的实体。根据运行环境和调度者的身份，线程可分为内核线程和用户线程。内核线程，在有的系统上也称为LWP（Light Weight Process，轻量级进程），运行在内核空间，由内核来调度；用户线程运行在用户空间，由线程库来调度。当进程的一个内核线程获得CPU的使用权时，它就加载并运行一个用户线程。可见，内核线程相当于用户线程运行的“容器”。一个进程可以拥有M个内核线程和N个用户线程，其中M≤N。并且在一个系统的所有进程中，M和N的比值都是固定的。按照M:N的取值，线程的实现方式可分为三种模式：完全在用户空间实现、完全由内核调度和双层调度（two level scheduler）。

现代Linux上默认使用的线程库是NPTL。用户可以使用如下命令来查看当前系统上所使用的线程库：


### 14.2 创建和结束线程

创建一个线程的函数是pthread_create。其定义如下：

```c
#include＜pthread.h＞
int pthread_create(pthread_t*thread,const pthread_attr_t*attr,void*(*start_routine)(void*),void*arg);<Paste>
// thread参数是新线程的标识符，后续pthread_*函数通过它来引用新线程。其类型pthread_t的定义如下：
#include＜bits/pthreadtypes.h＞
typedef unsigned long int pthread_t;
```

线程一旦被创建好，内核就可以调度内核线程来执行start_routine函数指针所指向的函数了。线程函数在结束时最好调用如下函数，以确保安全、干净地退出：

```c
#include＜pthread.h＞
// pthread_exit函数通过retval参数向线程的回收者传递其退出信息。它执行完之后不会返回到调用者，而且永远不会失败。
void pthread_exit(void*retval);
```

一个进程中的所有线程都可以调用pthread_join函数来回收其他线程（前提是目标线程是可回收的，见后文），即等待其他线程结束，这类似于回收进程的wait和waitpid系统调用。pthread_join的定义如下：

```c
#include＜pthread.h＞
int pthread_join(pthread_t thread,void**retval);
```

thread参数是目标线程的标识符，retval参数则是目标线程返回的退出信息。该函数会一直阻塞，直到被回收的线程结束为止。该函数成功时返回0，失败则返回错误码。

![](./pthread_jon_retval.png)

有时候我们希望异常终止一个线程，即取消线程，它是通过如下函数实现的：

```c
#include＜pthread.h＞
int pthread_cancel(pthread_t thread);
```

thread参数是目标线程的标识符。该函数成功时返回0，失败则返回错误码。不过，接收到取消请求的目标线程可以决定是否允许被取消以及如何取消，这分别由如下两个函数完成：

```c
#include＜pthread.h＞
int pthread_setcancelstate(int state,int*oldstate);
int pthread_setcanceltype(int type,int*oldtype);
```

### 14.3 线程属性

pthread_attr_t结构体定义了一套完整的线程属性，如下所示：
可见，各种线程属性全部包含在一个字符数组中。线程库定义了一系列函数来操作pthread_attr_t类型的变量，以方便我们获取和设置线程属性。这些函数包括：


```c
#include＜bits/pthreadtypes.h＞
#define__SIZEOF_PTHREAD_ATTR_T 36
typedef union
{
char__size[__SIZEOF_PTHREAD_ATTR_T];
long int__align;
}pthread_attr_t;

#include＜pthread.h＞
/*初始化线程属性对象*/
int pthread_attr_init(pthread_attr_t*attr);
/*销毁线程属性对象。被销毁的线程属性对象只有再次初始化之后才能继续使用*/
int pthread_attr_destroy(pthread_attr_t*attr);
/*下面这些函数用于获取和设置线程属性对象的某个属性*/
int pthread_attr_getdetachstate(const pthread_attr_t*attr,int*detachstate);
int pthread_attr_setdetachstate(pthread_attr_t*attr,int detachstate);
int pthread_attr_getstackaddr(const pthread_attr_t*attr,void**stackaddr);
int pthread_attr_setstackaddr(pthread_attr_t*attr,void*stackaddr);
int pthread_attr_getstacksize(const pthread_attr_t*attr,size_t*stacksize);
int pthread_attr_setstacksize(pthread_attr_t*attr,size_t stacksize);
int pthread_attr_getstack(const pthread_attr_t*attr,void**stackaddr,size_t*stacksize);
int pthread_attr_setstack(pthread_attr_t*attr,void*stackaddr,size_t stacksize);
int pthread_attr_getguardsize(const pthread_attr_t*__attr,size_t*guardsize);
int pthread_attr_setguardsize(pthread_attr_t*attr,size_t guardsize);
int pthread_attr_getschedparam(const pthread_attr_t*attr,struct sched_param*param);
int pthread_attr_setschedparam(pthread_attr_t*attr,const struct sched_param*param);
int pthread_attr_getschedpolicy(const pthread_attr_t*attr,int*policy);
int pthread_attr_setschedpolicy(pthread_attr_t*attr,int policy);
int pthread_attr_getinheritsched(const pthread_attr_t*attr,int*inherit);
int pthread_attr_setinheritsched(pthread_attr_t*attr,int inherit);
int pthread_attr_getscope(const pthread_attr_t*attr,int*scope);
int pthread_attr_setscope(pthread_attr_t*attr,int scope);
```

### 14.4 POSIX 信号量
3种专门用于线程同步的机制：POSIX信号量、互斥量和条件变量。
在Linux上，信号量API有两组。一组是第13章讨论过的System V IPC信号量，另外一组是我们现在要讨论的POSIX信号量。这两组接口很相似，但不保证能互换。由于这两种信号量的语义完全相同，因此我们不再赘述信号量的原理。

POSIX信号量函数的名字都以sem_开头，并不像大多数线程函数那样以pthread_开头。常用的POSIX信号量函数是下面5个：

```c
#include＜semaphore.h＞
int sem_init(sem_t*sem,int pshared,unsigned int value);
int sem_destroy(sem_t*sem);
int sem_wait(sem_t*sem);
int sem_trywait(sem_t*sem);
int sem_post(sem_t*sem);
```

### 14.5 互斥锁
互斥锁（也称互斥量）可以用于保护关键代码段，以确保其独占式的访问，这有点像一个二进制信号量（见13.5.1小节）。当进入关键代码段时，我们需要获得互斥锁并将其加锁，这等价于二进制信号量的P操作；当离开关键代码段时，我们需要对互斥锁解锁，以唤醒其他等待该互斥锁的线程，这等价于二进制信号量的V操作。

```c
#include＜pthread.h＞
int pthread_mutex_init(pthread_mutex_t*mutex,const pthread_mutexattr_t*mutexattr);
int pthread_mutex_destroy(pthread_mutex_t*mutex);
int pthread_mutex_lock(pthread_mutex_t*mutex);
int pthread_mutex_trylock(pthread_mutex_t*mutex);
int pthread_mutex_unlock(pthread_mutex_t*mutex);
```

pthread_mutexattr_t结构体定义了一套完整的互斥锁属性。线程库提供了一系列函数来操作pthread_mutexattr_t类型的变量，以方便我们获取和设置互斥锁属性。这里我们列出其中一些主要的函数：

```c
#include＜pthread.h＞
/*初始化互斥锁属性对象*/
int pthread_mutexattr_init(pthread_mutexattr_t*attr);
/*销毁互斥锁属性对象*/
int pthread_mutexattr_destroy(pthread_mutexattr_t*attr);
/*获取和设置互斥锁的pshared属性*/
int pthread_mutexattr_getpshared(const pthread_mutexattr_t*attr,int*pshared);
int pthread_mutexattr_setpshared(pthread_mutexattr_t*attr,int pshared);
/*获取和设置互斥锁的type属性*/
int pthread_mutexattr_gettype(const pthread_mutexattr_t*attr,int*type);
int pthread_mutexattr_settype(pthread_mutexattr_t*attr,int type);
```

### 14.6 条件变量

如果说互斥锁是用于同步线程对共享数据的访问的话，那么条件变量则是用于在线程之间同步共享数据的值。条件变量提供了一种线程间的通知机制：当某个共享数据达到某个值的时候，唤醒等待这个共享数据的线程。
条件变量的相关函数主要有如下5个：

```c
#include＜pthread.h＞
int pthread_cond_init(pthread_cond_t*cond,const pthread_condattr_t*cond_attr);
int pthread_cond_destroy(pthread_cond_t*cond);
int pthread_cond_broadcast(pthread_cond_t*cond);
int pthread_cond_signal(pthread_cond_t*cond);
int pthread_cond_wait(pthread_cond_t*cond,pthread_mutex_t*mutex);
```

### 14.8 多线程环境

##### 可重入函数
如果一个函数能被多个线程同时调用且不发生竞态条件，则我们称它是线程安全的（thread safe），或者说它是可重入函数。Linux库函数只有一小部分是不可重入的，比如5.1.4小节讨论的inet_ntoa函数，以及5.12.2小节讨论的getservbyname和getservbyport函数。关于Linux上不可重入的库函数的完整列表，请读者参考相关书籍，这里不再赘述。这些库函数之所以不可重入，主要是因为其内部使用了静态变量。不过Linux对很多不可重入的库函数提供了对应的可重入版本，这些可重入版本的函数名是在原函数名尾部加上_r。比如，函数localtime对应的可重入函数是localtime_r。在多线程程序中调用库函数，一定要使用其可重入版本，否则可能导致预想不到的结果。

##### 线程和进程
思考这样一个问题：如果一个多线程程序的某个线程调用了fork函数，那么新创建的子进程是否将自动创建和父进程相同数量的线程呢？答案是“否”，正如我们期望的那样。子进程只拥有一个执行线程，它是调用fork的那个线程的完整复制。并且子进程将自动继承父进程中互斥锁（条件变量与之类似）的状态。也就是说，父进程中已经被加锁的互斥锁在子进程中也是被锁住的。

##### 线程和信号
每个线程都可以独立地设置信号掩码。我们在10.3.2小节讨论过设置进程信号掩码的函数sigprocmask，但在多线程环境下我们应该使用如下所示的pthread版本的sigprocmask函数来设置线程信号掩码：

```c
#include＜pthread.h＞
#include＜signal.h＞
int pthread_sigmask(int how,const sigset_t*newmask,sigset_t*oldmask);
```

最后，pthread还提供了下面的方法，使得我们可以明确地将一个信号发送给指定的线程：

```c
#include＜signal.h＞
int pthread_kill(pthread_t thread,int sig);
```


# 15章 进程池和线程池

### 15.1 进程池和线程池概述

进程池中的所有子进程都运行着相同的代码，并具有相同的属性，比如优先级、PGID等。因为进程池在服务器启动之初就创建好了，所以每个子进程都相对“干净”，即它们没有打开不必要的文件描述符（从父进程继承而来），也不会错误地使用大块的堆内存（从父进程复制得到）。

![](./process_pool.png)

### 15.2 处理多客户

如果客户任务是存在上下文关系的，则最好一直用同一个子进程来为之服务，否则实现起来将比较麻烦，因为我们不得不在各子进程之间传递上下文数据。在9.3.4小节中，我们讨论了epoll的EPOLLONESHOT事件，这一事件能够确保一个客户连接在整个生命周期中仅被一个线程处理。

### 15.3　半同步/半异步进程池实现

为了避免在父、子进程之间传递文件描述符，我们将接受新连接的操作放到子进程中。很显然，对于这种模式而言，一个客户连接上的所有任务始终是由一个子进程来处理的。

后面都是代码实现，具体参考原书


# 第16章　服务器调制、调试和测试

### 16.1　最大文件描述符数
文件描述符是服务器程序的宝贵资源，几乎所有的系统调用都是和文件描述符打交道。系统分配给应用程序的文件描述符数量是有限制的，所以我们必须总是关闭那些已经不再使用的文件描述符，以释放它们占用的资源。比如作为守护进程运行的服务器程序就应该总是关闭标准输入、标准输出和标准错误这3个文件描述符。
Linux对应用程序能打开的最大文件描述符数量有两个层次的限制：用户级限制和系统级限制。用户级限制是指目标用户运行的所有进程总共能打开的文件描述符数；系统级的限制是指所有用户总共能打开的文件描述符数。

```shell
# 下面这个命令是最常用的查看用户级文件描述符数限制的方法：
ulimit -n
```

为永久修改用户级文件描述符数限制，可以在/etc/security/limits.conf文件中加入如下两项：

```shell
*hard nofile max-file-number
*hard nofile max-file-number
*soft nofile max-file-number
```

要永久更改系统级文件描述符数限制，则需要在/etc/sysctl.conf文件中添加如下一项：

```shell
fs.file-max=max-file-number # 然后通过执行sysctl-p命令使更改生效。
```

### 16.2　调整内核参数

##### 16.2.1　/proc/sys/fs目录下的部分文件

❑/proc/sys/fs/file-max，系统级文件描述符数限制。
❑/proc/sys/fs/epoll/max_user_watches，一个用户能够往epoll内核事件表中注册的事件的总量。

##### 16.2.2　/proc/sys/net目录下的部分文件

❑/proc/sys/net/core/somaxconn，指定listen监听队列里，能够建立完整连接从而进入ESTABLISHED状态的socket的最大数目。

❑/proc/sys/net/ipv4/tcp_max_syn_backlog，指定listen监听队列里，能够转移至ESTAB-LISHED或者SYN_RCVD状态的socket的最大数目。o

❑/proc/sys/net/ipv4/tcp_wmem，它包含3个值，分别指定一个socket的TCP写缓冲区的最小值、默认值和最大值。

❑/proc/sys/net/ipv4/tcp_rmem，它包含3个值，分别指定一个socket的TCP读缓冲区的最小值、默认值和最大值。在代码清单3-6中，我们正是通过修改这个参数来改变接收通告窗口大小的。

❑/proc/sys/net/ipv4/tcp_syncookies，指定是否打开TCP同步标签（syncookie）。同步标签通过启动cookie来防止一个监听socket因不停地重复接收来自同一个地址的连接请求（同步报文段），而导致listen监听队列溢出（所谓的SYN风暴）。<Paste>

### gdb 调试

##### 16.3.1　用gdb调试多进程程序

如果一个进程通过fork系统调用创建了子进程，gdb会继续调试原来的进程，子进程则正常运行。那么该如何调试子进程呢？常用的方法有如下两种。

1.单独调试子进程

子进程从本质上说也是一个进程，因此我们可以用通用的gdb调试方法来调试它。举例来说，如果要调试代码清单15-2描述的CGI进程池服务器的某一个子进程，则我们可以先运行服务器，然后找到目标子进程的PID，再将其附加（attach）到gdb调试器上，具体操作如代码清单16-1所示。

2.使用调试器选项follow-fork-mode

gdb调试器的选项follow-fork-mode允许我们选择程序在执行fork系统调用后是继续调试父进程还是调试子进程。其用法如下：

`(gdb)set follow-fork-mode mode`

##### 16.3.2　用gdb调试多线程程序

gdb有一组命令可辅助多线程程序的调试。下面我们仅列举其中常用的一些：

❑info threads，显示当前可调试的所有线程。gdb会为每个线程分配一个ID，我们可以使用这个ID来操作对应的线程。ID前面有“*”号的线程是当前被调试的线程。

❑thread ID，调试目标ID指定的线程。

❑set scheduler-locking[off|on|step]。调试多线程程序时，默认除了被调试的线程在执行外，其他线程也在继续执行，但有的时候我们希望只让被调试的线程运行。这可以通过这个命令来实现。该命令设置scheduler-locking的值：off表示不锁定任何线程，即所有线程都可以继续执行，这是默认值；on表示只有当前被调试的线程会继续执行；step表示在单步执行的时候，只有当前线程会执行。

### 16.4 压力测试
压力测试程序有很多种实现方式，比如I/O复用方式，多线程、多进程并发编程方式，以及这些方式的结合使用。不过，单纯的I/O复用方式的施压程度是最高的，因为线程和进程的调度本身也是要占用一定CPU时间的。我们将使用epoll来实现一个通用的服务器压力测试程序，

```c
#include＜stdlib.h＞
#include＜stdio.h＞
#include＜assert.h＞
#include＜unistd.h＞
#include＜sys/types.h＞
#include＜sys/epoll.h＞
#include＜fcntl.h＞
#include＜sys/socket.h＞
#include＜netinet/in.h＞
#include＜arpa/inet.h＞
#include＜string.h＞
/*每个客户连接不停地向服务器发送这个请求*/
static const char*request="GET http://localhost/index.html HTTP/1.1\r\nConnection:keep-alive\r\n\r\nxxxxxxxxxxxx";
int setnonblocking(int fd)
{
    int old_option=fcntl(fd,F_GETFL);
    int new_option=old_option|O_NONBLOCK;
    fcntl(fd,F_SETFL,new_option);
    return old_option;
}
void addfd(int epoll_fd,int fd)
{
    epoll_event event;
    event.data.fd=fd;
    event.events=EPOLLOUT|EPOLLET|EPOLLERR;
    epoll_ctl(epoll_fd,EPOLL_CTL_ADD,fd,＆event);
    setnonblocking(fd);
}
/*向服务器写入len字节的数据*/
bool write_nbytes(int sockfd,const char*buffer,int len)
{
    int bytes_write=0;
    printf("write out%d bytes to socket%d\n",len,sockfd);
    while(1)
    {
        bytes_write=send(sockfd,buffer,len,0);
        if(bytes_write==-1)
        {
            return false;
        }
        else if(bytes_write==0)
        {
            return false;
        }
        len-=bytes_write;
        buffer=buffer+bytes_write;
        if(len＜=0)
        {
            return true;
        }
    }
}
/*从服务器读取数据*/
bool read_once(int sockfd,char*buffer,int len)
{
    int bytes_read=0;
    memset(buffer,'\0',len);
    bytes_read=recv(sockfd,buffer,len,0);
    if(bytes_read==-1)
    {
        return false;
    }
    else if(bytes_read==0)
    {
        return false;
    }
    printf("read in%d bytes from socket%d with content:%s\n",bytes_read,sockfd,buffer);
    return true;
}
/*向服务器发起num个TCP连接，我们可以通过改变num来调整测试压力*/
void start_conn(int epoll_fd,int num,const char*ip,int port)
{
    int ret=0;
    struct sockaddr_in address;
    bzero(＆address,sizeof(address));
    address.sin_family=AF_INET;
    inet_pton(AF_INET,ip,＆address.sin_addr);
    address.sin_port=htons(port);
    for(int i=0;i＜num;++i)
    {
        sleep(1);
        int sockfd=socket(PF_INET,SOCK_STREAM,0);
        printf("create 1 sock\n");
        if(sockfd＜0)
        {
            continue;
        }
        if(connect(sockfd,(struct sockaddr*)＆address,sizeof(address))==0)
        {
            printf("build connection%d\n",i);
            addfd(epoll_fd,sockfd);
        }
    }
}
void close_conn(int epoll_fd,int sockfd)
{
    epoll_ctl(epoll_fd,EPOLL_CTL_DEL,sockfd,0);
    close(sockfd);
}
int main(int argc,char*argv[])
{
    assert(argc==4);
    int epoll_fd=epoll_create(100);
    start_conn(epoll_fd,atoi(argv[3]),argv[1],atoi(argv[2]));
    epoll_event events[10000];
    char buffer[2048];
    while(1)
    {
        int fds=epoll_wait(epoll_fd,events,10000,2000);
        for(int i=0;i＜fds;i++)
        {
            int sockfd=events[i].data.fd;
            if(events[i].events＆EPOLLIN)
            {
                if(!read_once(sockfd,buffer,2048))
                {
                    close_conn(epoll_fd,sockfd);
                }
                struct epoll_event event;
                event.events=EPOLLOUT|EPOLLET|EPOLLERR;
                event.data.fd=sockfd;
                epoll_ctl(epoll_fd,EPOLL_CTL_MOD,sockfd,＆event);
            }
            else if(events[i].events＆EPOLLOUT)
            {
                if(!write_nbytes(sockfd,request,strlen(request)))
                {
                    close_conn(epoll_fd,sockfd);
                }
                struct epoll_event event;
                event.events=EPOLLIN|EPOLLET|EPOLLERR;
                event.data.fd=sockfd;
                epoll_ctl(epoll_fd,EPOLL_CTL_MOD,sockfd,＆event);
            }
            else if(events[i].events＆EPOLLERR)
            {
                close_conn(epoll_fd,sockfd);
            }
        }
    }
}
```

# 17 系统检测工具

- tcpdump: 网络抓包工具 (Wireshark)
- lsof（list open file）:是一个列出当前系统打开的文件描述符的工具。通过它我们可以了解感兴趣的进程打开了哪些文件描述符，或者我们感兴趣的文件描述符被哪些进程打开了。
- nc: nc（netcat）命令短小精干、功能强大，有着“瑞士军刀”的美誉。它主要被用来快速构建网络连接。我们可以让它以服务器方式运行，监听某个端口并接收客户连接，因此它可用来调试客户端程序。我们也可以使之以客户端方式运行，向服务器发起连接并收发数据，因此它可以用来调试服务器程序，此时它有点像telnet程序。
- strace: 测试服务器性能的重要工具。它跟踪程序运行过程中执行的系统调用和接收到的信号，并将系统调用名、参数、返回值及信号名输出到标准输出或者指定的文件。
- netstat: 是一个功能很强大的网络信息统计工具。它可以打印本地网卡接口上的全部连接、路由表信息、网卡接口信息等。对本书而言，我们主要利用的是上述功能中的第一个，即显示TCP连接及其状态信息。毕竟，要获得路由表信息和网卡接口信息，我们可以使用输出内容更丰富的route和ifconfig命令。
- vmstat:是 virtual memory statistics的缩写，它能实时输出系统的各种资源的使用情况，比如进程信息、内存使用、CPU使用率以及I/O使用情况。不过，我们可以使用iostat命令获得磁盘使用情况的更多信息，也可以使用mpstat获得CPU使用情况的更多信息。vmstat命令主要用于查看系统内存的使用情况。
- ifstat: 是interface statistics的缩写，它是一个简单的网络流量监测工具。
- mpstat: 是multi-processor statistics的缩写，它能实时监测多处理器系统上每个CPU的使用情况。mpstat命令和iostat命令通常都集成在包sysstat中，安装sysstat即可获得这两个命令。
