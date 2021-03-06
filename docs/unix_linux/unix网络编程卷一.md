# unix 网络编程代码下载

```bash
# clone code
git clone https://github.com/DingHe/unpv13e.git

cd unpv13e
./configure
cd lib
make

cd ..
cp libunp.a /usr/local/lib/

cd ./tcpcliserv
make all
```

# 1-5 Tcp 套接口

Tcp 回显示例：

```c

// strcli11.c
/* Use standard echo server; baseline measurements for nonblocking version */
#include	"unp.h"

int
main(int argc, char **argv)
{
	int					sockfd;
	struct sockaddr_in	servaddr;

	if (argc != 2)
		err_quit("usage: tcpcli <IPaddress>");

	sockfd = Socket(AF_INET, SOCK_STREAM, 0);

	bzero(&servaddr, sizeof(servaddr));
	servaddr.sin_family = AF_INET;
	// servaddr.sin_port = htons(7);
	servaddr.sin_port = htons(SERV_PORT);
	Inet_pton(AF_INET, argv[1], &servaddr.sin_addr);

	Connect(sockfd, (SA *) &servaddr, sizeof(servaddr));

	str_cli(stdin, sockfd);		/* do it all */

	exit(0);
}

#include	"unp.h"

void
str_cli(FILE *fp, int sockfd)
{
	char	sendline[MAXLINE], recvline[MAXLINE];

	while (Fgets(sendline, MAXLINE, fp) != NULL) {

		Writen(sockfd, sendline, 1); // 第一次引发一个 RST
		sleep(1);
		Writen(sockfd, sendline+1, strlen(sendline)-1); //第二次产生 SIGPIPE


		if (Readline(sockfd, recvline, MAXLINE) == 0)
			err_quit("str_cli: server terminated prematurely");

		Fputs(recvline, stdout);
	}
}


// tcpcliserv/tcpserv09.c
#include	"unp.h"

int
main(int argc, char **argv)
{
	int					listenfd, connfd;
	pid_t				childpid;
	socklen_t			clilen;
	struct sockaddr_in	cliaddr, servaddr;
	void				sig_chld(int);

	listenfd = Socket(AF_INET, SOCK_STREAM, 0);

	bzero(&servaddr, sizeof(servaddr));
	servaddr.sin_family      = AF_INET;
	servaddr.sin_addr.s_addr = htonl(INADDR_ANY);
	servaddr.sin_port        = htons(SERV_PORT);

	Bind(listenfd, (SA *) &servaddr, sizeof(servaddr));

	Listen(listenfd, LISTENQ);

	Signal(SIGCHLD, sig_chld);

	for ( ; ; ) {
		clilen = sizeof(cliaddr);
		if ( (connfd = accept(listenfd, (SA *) &cliaddr, &clilen)) < 0) {
			if (errno == EINTR)
				continue;		/* back to for() */
			else
				err_sys("accept error");
		}

		if ( (childpid = Fork()) == 0) {	/* child process */
			Close(listenfd);	/* close listening socket */
			str_echo(connfd);	/* process the request */
			exit(0);
		}
		Close(connfd);			/* parent closes connected socket */
	}
}
```

# 6章 I/O 复用：select 和 poll 函数

unix 可用的五种 I/O 模型：
- 阻塞式 I/O
- 非阻塞式 I/O, nonblocking
- I/O 复用(select poll), multiplexing
- 信号驱动式 I/O (SIGIO), signal-driven，内核告诉我们何时可以启动一个 I/O 操作
- 异步 I/O (POSIX 的 aio_系列函数), asynchronous I/O，内核告诉我们 I/O 操作何时完成

一个输入操作通常包括两个不同阶段：
1. 等待数据准备好
2. 从内核向进程复制数据


POSIX 定义术语：
- 同步 I/O 操作(synchronous I/O opetation) 导致请求进程阻塞，直到 I/O 操作完成。
- 异步 I/O 操作(asynchronous I/O opetation) 不导致请求进程阻塞。前面五种只有异步 I/O 与 POSIX  定义的异步 I/O 匹配。

select 函数：允许进程指示内核等待多个事件中的任何一个发生，并只在有一个或多个事件发生或经历一段指定的时间后才唤醒它。

```c
#include <sys/select.h>
#include <sys/time.h>
int select(int maxfdp1, fd_et *readset, fd_set *writeset, fd_set *exceptset, const struct timeval *timeout);
```


```c
//select/strcliselect02.c
#include	"unp.h"

void
str_cli(FILE *fp, int sockfd)
{
	int			maxfdp1, stdineof;
	fd_set		rset;
	char		buf[MAXLINE];
	int		n;

	stdineof = 0;
	FD_ZERO(&rset);
	for ( ; ; ) {
		if (stdineof == 0)
			FD_SET(fileno(fp), &rset);
		FD_SET(sockfd, &rset);
		maxfdp1 = max(fileno(fp), sockfd) + 1;
		Select(maxfdp1, &rset, NULL, NULL, NULL);

		if (FD_ISSET(sockfd, &rset)) {	/* socket is readable */
			if ( (n = Read(sockfd, buf, MAXLINE)) == 0) {
				if (stdineof == 1)
					return;		/* normal termination */
				else
					err_quit("str_cli: server terminated prematurely");
			}

			Write(fileno(stdout), buf, n);
		}

		if (FD_ISSET(fileno(fp), &rset)) {  /* input is readable */
			if ( (n = Read(fileno(fp), buf, MAXLINE)) == 0) {
				stdineof = 1;
				Shutdown(sockfd, SHUT_WR);	/* send FIN */
				FD_CLR(fileno(fp), &rset);
				continue;
			}

			Writen(sockfd, buf, n);
		}
	}
}


/* include fig01 */
#include	"unp.h"

int
main(int argc, char **argv)
{
	int					i, maxi, maxfd, listenfd, connfd, sockfd;
	int					nready, client[FD_SETSIZE];
	ssize_t				n;
	fd_set				rset, allset;
	char				buf[MAXLINE];
	socklen_t			clilen;
	struct sockaddr_in	cliaddr, servaddr;

	listenfd = Socket(AF_INET, SOCK_STREAM, 0);

	bzero(&servaddr, sizeof(servaddr));
	servaddr.sin_family      = AF_INET;
	servaddr.sin_addr.s_addr = htonl(INADDR_ANY);
	servaddr.sin_port        = htons(SERV_PORT);

	Bind(listenfd, (SA *) &servaddr, sizeof(servaddr));

	Listen(listenfd, LISTENQ);

	maxfd = listenfd;			/* initialize */
	maxi = -1;					/* index into client[] array */
	for (i = 0; i < FD_SETSIZE; i++)
		client[i] = -1;			/* -1 indicates available entry */
	FD_ZERO(&allset);
	FD_SET(listenfd, &allset);
/* end fig01 */

/* include fig02 */
	for ( ; ; ) {
		rset = allset;		/* structure assignment */
		nready = Select(maxfd+1, &rset, NULL, NULL, NULL); // 等待事件发生。返回就绪描述字正数目，0-超时，-1-出错

		if (FD_ISSET(listenfd, &rset)) {	/* new client connection */
			clilen = sizeof(cliaddr);
			connfd = Accept(listenfd, (SA *) &cliaddr, &clilen);
#ifdef	NOTDEF
			printf("new client: %s, port %d\n",
					Inet_ntop(AF_INET, &cliaddr.sin_addr, 4, NULL),
					ntohs(cliaddr.sin_port));
#endif

			for (i = 0; i < FD_SETSIZE; i++)
				if (client[i] < 0) {
					client[i] = connfd;	/* save descriptor */
					break;
				}
			if (i == FD_SETSIZE)
				err_quit("too many clients");

			FD_SET(connfd, &allset);	/* add new descriptor to set */
			if (connfd > maxfd)
				maxfd = connfd;			/* for select */
			if (i > maxi)
				maxi = i;				/* max index in client[] array */

			if (--nready <= 0)
				continue;				/* no more readable descriptors */
		}

		for (i = 0; i <= maxi; i++) {	/* check all clients for data */
			if ( (sockfd = client[i]) < 0)
				continue;
			if (FD_ISSET(sockfd, &rset)) {
				if ( (n = Read(sockfd, buf, MAXLINE)) == 0) {
						/*4connection closed by client */
					Close(sockfd);
					FD_CLR(sockfd, &allset);
					client[i] = -1;
				} else
					Writen(sockfd, buf, n);

				if (--nready <= 0)
					break;				/* no more readable descriptors */
			}
		}
	}
}
/* end fig02 */
```

使用 poll 函数改写 tcp server

```c
#include <poll.h>
int poll(struct pollfd *fdarray, unsigned long nfds, int timeout);//返回就绪描述字个数。0 超时，-1 出错

struct pollfd {
	int fd; // descriptor to check
	short events; // events of interest on fd
	short revents; // events occurred on fd
}
```

```c
/* include fig01 */
#include	"unp.h"
#include	<limits.h>		/* for OPEN_MAX */

int
main(int argc, char **argv)
{
	int					i, maxi, listenfd, connfd, sockfd;
	int					nready;
	ssize_t				n;
	char				buf[MAXLINE];
	socklen_t			clilen;
	struct pollfd		client[OPEN_MAX];
	struct sockaddr_in	cliaddr, servaddr;

	listenfd = Socket(AF_INET, SOCK_STREAM, 0);

	bzero(&servaddr, sizeof(servaddr));
	servaddr.sin_family      = AF_INET;
	servaddr.sin_addr.s_addr = htonl(INADDR_ANY);
	servaddr.sin_port        = htons(SERV_PORT);

	Bind(listenfd, (SA *) &servaddr, sizeof(servaddr));

	Listen(listenfd, LISTENQ);

	client[0].fd = listenfd;
	client[0].events = POLLRDNORM;
	for (i = 1; i < OPEN_MAX; i++)
		client[i].fd = -1;		/* -1 indicates available entry */
	maxi = 0;					/* max index into client[] array */
/* end fig01 */

/* include fig02 */
	for ( ; ; ) {
		nready = Poll(client, maxi+1, INFTIM);

		if (client[0].revents & POLLRDNORM) {	/* new client connection */
			clilen = sizeof(cliaddr);
			connfd = Accept(listenfd, (SA *) &cliaddr, &clilen);
#ifdef	NOTDEF
			printf("new client: %s\n", Sock_ntop((SA *) &cliaddr, clilen));
#endif

			for (i = 1; i < OPEN_MAX; i++) // 从 1 开始，0 用作监听套接口
				if (client[i].fd < 0) { // 找到第一个可用项
					client[i].fd = connfd;	/* save descriptor */
					break;
				}
			if (i == OPEN_MAX)
				err_quit("too many clients");

			client[i].events = POLLRDNORM;
			if (i > maxi)
				maxi = i;				/* max index in client[] array */

			if (--nready <= 0)
				continue;				/* no more readable descriptors */
		}

		for (i = 1; i <= maxi; i++) {	/* check all clients for data */
			if ( (sockfd = client[i].fd) < 0)
				continue;
			if (client[i].revents & (POLLRDNORM | POLLERR)) {
				if ( (n = read(sockfd, buf, MAXLINE)) < 0) {
					if (errno == ECONNRESET) {
							/*4connection reset by client */
#ifdef	NOTDEF
						printf("client[%d] aborted connection\n", i);
#endif
						Close(sockfd);
						client[i].fd = -1;
					} else
						err_sys("read error");
				} else if (n == 0) {
						/*4connection closed by client */
#ifdef	NOTDEF
					printf("client[%d] closed connection\n", i);
#endif
					Close(sockfd);
					client[i].fd = -1;
				} else
					Writen(sockfd, buf, n);

				if (--nready <= 0)
					break;				/* no more readable descriptors */
			}
		}
	}
}
/* end fig02 */
```

# 7. 套接口选项

```c
#include <sys/socket.h>
int getsockopt(int sockfd, int level, int optname, vod *optval, socklen_t *optlen);
int setsockopt(int sockfd, int level, int optname, const void *optval, socklen_t *optlen);

# include<fcntl.h>
int fcntl(int fd, int cmd, ... /* int arg */); // file control
```

# 8 udp 套接口编程

回显client示例：

```c
//udpcliserv/udpcli04.c
#include	"unp.h"

int
main(int argc, char **argv)
{
	int					sockfd;
	struct sockaddr_in	servaddr;

	if (argc != 2)
		err_quit("usage: udpcli <IPaddress>");

	bzero(&servaddr, sizeof(servaddr));
	servaddr.sin_family = AF_INET;
	servaddr.sin_port = htons(SERV_PORT);
	Inet_pton(AF_INET, argv[1], &servaddr.sin_addr);

	sockfd = Socket(AF_INET, SOCK_DGRAM, 0);

	dg_cli(stdin, sockfd, (SA *) &servaddr, sizeof(servaddr));

	exit(0);
}

//udpcliserv/dgcliconnect.c
#include	"unp.h"

void
dg_cli(FILE *fp, int sockfd, const SA *pservaddr, socklen_t servlen)
{
	int		n;
	char	sendline[MAXLINE], recvline[MAXLINE + 1];

	Connect(sockfd, (SA *) pservaddr, servlen);

	while (Fgets(sendline, MAXLINE, fp) != NULL) {

		Write(sockfd, sendline, strlen(sendline));

		n = Read(sockfd, recvline, MAXLINE);

		recvline[n] = 0;	/* null terminate */
		Fputs(recvline, stdout);
	}
}

```

回显 udp server 示例：

```c
#include	"unp.h"

int
main(int argc, char **argv)
{
	int					sockfd;
	struct sockaddr_in	servaddr, cliaddr;

	sockfd = Socket(AF_INET, SOCK_DGRAM, 0);

	bzero(&servaddr, sizeof(servaddr));
	servaddr.sin_family      = AF_INET;
	servaddr.sin_addr.s_addr = htonl(INADDR_ANY);
	servaddr.sin_port        = htons(SERV_PORT);

	Bind(sockfd, (SA *) &servaddr, sizeof(servaddr));

	dg_echo(sockfd, (SA *) &cliaddr, sizeof(cliaddr));
}


#include	"unp.h"

static void	recvfrom_int(int);
static int	count;

void
dg_echo(int sockfd, SA *pcliaddr, socklen_t clilen)
{
	socklen_t	len;
	char		mesg[MAXLINE];

	Signal(SIGINT, recvfrom_int);

	for ( ; ; ) {
		len = clilen;
		Recvfrom(sockfd, mesg, MAXLINE, 0, pcliaddr, &len);

		count++;
	}
}

static void
recvfrom_int(int signo)
{
	printf("\nreceived %d datagrams\n", count);
	exit(0);
}
```


# 11. 名字与地址转换

以下这些函数可以在 python 中测试一下，使用起来要比 c 语言简便很多。

```c
#include <netdb.h>

// 根据主机名获取 ipv4 。 python : socket.gethostbyname("www.baidu.com")
struct hostent *gethostbyname(const char *hostname);

// 根据二进制 ip 地址找到主机名。 python：socket.gethostbyaddr("4.2.2.2")
struct hostent *gethostbyaddr(const char *addr, socklen_t len, int family);

// socket.getaddrinfo(host, port, family=0, type=0, proto=0, flags=0)。协议无关
#include <netdb.h>
int getaddrinfo(const char *hostname , const char *service , const struct addrinfo *hints , struct addrinfo **result );
// 返回：若成功则为0，若出错则为非0（见图11-7）
```

[可重入函数](https://baike.baidu.com/item/%E5%8F%AF%E9%87%8D%E5%85%A5%E5%87%BD%E6%95%B0)：
可重入函数主要用于多任务环境中，一个可重入的函数简单来说就是可以被中断的函数，也就是说，可以在这个函数执行的任何时刻中断它，
转入OS调度下去执行另外一段代码，而返回控制时不会出现什么错误；而不可重入的函数由于使用了一些系统资源，比如全局变量区，
中断向量表等，所以它如果被中断的话，可能会出现问题，这类函数是不能运行在多任务环境下的。


# 12. ipv4 与 ipv6 的互操作性

双栈(dual stacks):ipv4和ipv6协议栈。

# 13. 守护进程和 inetd 超级服务器

守护进程(daemon): 后台运行且不与任何控制终端关联的进程。

```c
#include	"unp.h"
#include	<time.h>

int main(int argc, char **argv)
{
	int listenfd, connfd;
	socklen_t addrlen, len;
	struct sockaddr	*cliaddr;
	char buff[MAXLINE];
	time_t ticks;

	if (argc < 2 || argc > 3)
		err_quit("usage: daytimetcpsrv2 [ <host> ] <service or port>");

	daemon_init(argv[0], 0);

	if (argc == 2)
		listenfd = Tcp_listen(NULL, argv[1], &addrlen);
	else
		listenfd = Tcp_listen(argv[1], argv[2], &addrlen);

	cliaddr = Malloc(addrlen);

	for ( ; ; ) {
		len = addrlen;
		connfd = Accept(listenfd, cliaddr, &len);
		err_msg("connection from %s", Sock_ntop(cliaddr, len));

		ticks = time(NULL);
		snprintf(buff, sizeof(buff), "%.24s\r\n", ctime(&ticks));
		Write(connfd, buff, strlen(buff));

		Close(connfd);
	}
}
```

# 14. 高级 IO 函数

## 14.2 套接口超时

在涉及套接字的I/O操作上设置超时的方法有以下3种。

1. 调用alarm，它在指定超时期满时产生SIGALRM信号。这个方法涉及信号处理，而信号处理在不同的实现上存在差异，而且可能干扰进程中现有的alarm调用。
2. 在select中阻塞等待I/O（select有内置的时间限制），以此代替直接阻塞在read或write调用上。
3. 使用较新的SO_RCVTIMEO和SO_SNDTIMEO套接字选项。这个方法的问题在于并非所有实现都支持这两个套接字选项

## 14.9 高级轮询技术

- dev/poll
- kqueue

```c
// advio/str_cli_kqueue04.c
#include	"unp.h"

void
str_cli(FILE *fp, int sockfd)
{
	int				kq, i, n, nev, stdineof = 0, isfile;
	char			buf[MAXLINE];
	struct kevent	kev[2];
	struct timespec	ts;
	struct stat		st;

	isfile = ((fstat(fileno(fp), &st) == 0) &&
			  (st.st_mode & S_IFMT) == S_IFREG);

	EV_SET(&kev[0], fileno(fp), EVFILT_READ, EV_ADD, 0, 0, NULL);
	EV_SET(&kev[1], sockfd, EVFILT_READ, EV_ADD, 0, 0, NULL);

	kq = Kqueue();
	ts.tv_sec = ts.tv_nsec = 0;
	Kevent(kq, kev, 2, NULL, 0, &ts);

	for ( ; ; ) {
		nev = Kevent(kq, NULL, 0, kev, 2, NULL);

		for (i = 0; i < nev; i++) {
			if (kev[i].ident == sockfd) {	/* socket is readable */
				if ( (n = Read(sockfd, buf, MAXLINE)) == 0) {
					if (stdineof == 1)
						return;		/* normal termination */
					else
						err_quit("str_cli: server terminated prematurely");
				}

				Write(fileno(stdout), buf, n);
			}

			if (kev[i].ident == fileno(fp)) {  /* input is readable */
				n = Read(fileno(fp), buf, MAXLINE);
				if (n > 0)
					Writen(sockfd, buf, n);

				if (n == 0 || (isfile && n == kev[i].data)) {
					stdineof = 1;
					Shutdown(sockfd, SHUT_WR);	/* send FIN */
					kev[i].flags = EV_DELETE;
					Kevent(kq, &kev[i], 1, NULL, 0, &ts);	/* remove kevent */
					continue;
				}
			}
		}
	}
}
```

# 15. unix 域协议

unix域协议不是一个实际的协议族，而是单个主机上执行客户/服务器通信的一种方法。


```c
// unixdomain/unixdgserv01.c
#include	"unp.h"

int
main(int argc, char **argv)
{
	int					sockfd;
	struct sockaddr_un	servaddr, cliaddr;

	sockfd = Socket(AF_LOCAL, SOCK_DGRAM, 0);

	unlink(UNIXDG_PATH);
	bzero(&servaddr, sizeof(servaddr));
	servaddr.sun_family = AF_LOCAL;
	strcpy(servaddr.sun_path, UNIXDG_PATH);

	Bind(sockfd, (SA *) &servaddr, sizeof(servaddr));

	dg_echo(sockfd, (SA *) &cliaddr, sizeof(cliaddr));
}

// unixdomain/unixdgcli01.c
#include	"unp.h"

int
main(int argc, char **argv)
{
	int					sockfd;
	struct sockaddr_un	cliaddr, servaddr;

	sockfd = Socket(AF_LOCAL, SOCK_DGRAM, 0);

	bzero(&cliaddr, sizeof(cliaddr));		/* bind an address for us */
	cliaddr.sun_family = AF_LOCAL;
	strcpy(cliaddr.sun_path, tmpnam(NULL));

	Bind(sockfd, (SA *) &cliaddr, sizeof(cliaddr));

	bzero(&servaddr, sizeof(servaddr));	/* fill in server's address */
	servaddr.sun_family = AF_LOCAL;
	strcpy(servaddr.sun_path, UNIXDG_PATH);

	dg_cli(stdin, sockfd, (SA *) &servaddr, sizeof(servaddr));

	exit(0);
}
```

- socketpair 函数


# 16. 非阻塞 IO

套接字的默认状态是阻塞的。这就意味着当发出一个不能立即完成的套接字调用时，其进程将被投入睡眠，等待相应操作完成。可能阻塞的套接字调用可分为以下四类。

1. 输入操作，包括read、readv、recv、recvfrom和recvmsg共5个函数。如果某个进程对一个阻塞的TCP套接字（默认设置）调用这些输入函数之一，而且该套接字的接收缓冲区中没有数据可读，该进程将被投入睡眠，直到有一些数据到达。既然TCP是字节流协议，该进程的唤醒就是只要有一些数据到达，这些数据既可能是单个字节，也可以是一个完整的TCP分节中的数据。如果想等到某个固定数目的数据可读为止，那么可以调用我们的readn函数（图3-15），或者指定MSG_WAITALL标志（图14-6）。
2. 输出操作，包括write、writev、send、sendto和sendmsg共5个函数。对于一个TCP套接字我们已在2.11节说过，内核将从应用进程的缓冲区到该套接字的发送缓冲区复制数据。对于阻塞的套接字，如果其发送缓冲区中没有空间，进程将被投入睡眠，直到有空间为止。
对于一个非阻塞的TCP套接字，如果其发送缓冲区中根本没有空间，输出函数调用将立即返回一个EWOULDBLOCK错误。如果其发送缓冲区中有一些空间，返回值将是内核能够复制到该缓冲区中的字节数。这个字节数也称为不足计数（short count）。
3. 接受外来连接，即accept函数。如果对一个阻塞的套接字调用accept函数，并且尚无新的连接到达，调用进程将被投入睡眠。如果对一个非阻塞的套接字调用accept函数，并且尚无新的连接到达，accept调用将立即返回一个EWOULDBLOCK错误。
4. 发起外出连接，即用于TCP的connect函数。（回顾一下，我们知道connect同样可用于UDP，不过它不能使一个“真正”的连接建立起来，它只是使内核保存对端的IP地址和端口号。）我们已在2.6节展示过，TCP连接的建立涉及一个三路握手过程，而且connect函数一直要等到客户收到对于自己的SYN的ACK为止才返回。这意味着TCP的每个connect总会阻塞其调用进程至少一个到服务器的RTT时间。

## 16.2 非阻塞读和写: str_cli 函数

```c
/* include nonb1 */
#include	"unp.h"

void
str_cli(FILE *fp, int sockfd)
{
	int			maxfdp1, val, stdineof;
	ssize_t		n, nwritten;
	fd_set		rset, wset;
	char		to[MAXLINE], fr[MAXLINE];
	char		*toiptr, *tooptr, *friptr, *froptr;

	val = Fcntl(sockfd, F_GETFL, 0);
	Fcntl(sockfd, F_SETFL, val | O_NONBLOCK);

	val = Fcntl(STDIN_FILENO, F_GETFL, 0);
	Fcntl(STDIN_FILENO, F_SETFL, val | O_NONBLOCK);

	val = Fcntl(STDOUT_FILENO, F_GETFL, 0);
	Fcntl(STDOUT_FILENO, F_SETFL, val | O_NONBLOCK);

	toiptr = tooptr = to;	/* initialize buffer pointers */
	friptr = froptr = fr;
	stdineof = 0;

	maxfdp1 = max(max(STDIN_FILENO, STDOUT_FILENO), sockfd) + 1;
	for ( ; ; ) {
		FD_ZERO(&rset);
		FD_ZERO(&wset);
		if (stdineof == 0 && toiptr < &to[MAXLINE])
			FD_SET(STDIN_FILENO, &rset);	/* read from stdin */
		if (friptr < &fr[MAXLINE])
			FD_SET(sockfd, &rset);			/* read from socket */
		if (tooptr != toiptr)
			FD_SET(sockfd, &wset);			/* data to write to socket */
		if (froptr != friptr)
			FD_SET(STDOUT_FILENO, &wset);	/* data to write to stdout */

		Select(maxfdp1, &rset, &wset, NULL, NULL);
/* end nonb1 */
/* include nonb2 */
		if (FD_ISSET(STDIN_FILENO, &rset)) { // 如果标准输入可以读，就调用 read
			if ( (n = read(STDIN_FILENO, toiptr, &to[MAXLINE] - toiptr)) < 0) {
				if (errno != EWOULDBLOCK) // 忽略 EWOULDBLOCK 错误
					err_sys("read error on stdin");

			} else if (n == 0) {
#ifdef	VOL2
				fprintf(stderr, "%s: EOF on stdin\n", gf_time());
#endif
				stdineof = 1;			/* all done with stdin */
				if (tooptr == toiptr)
					Shutdown(sockfd, SHUT_WR);/* send FIN */

			} else {
#ifdef	VOL2
				fprintf(stderr, "%s: read %d bytes from stdin\n", gf_time(), n);
#endif
				toiptr += n;			/* # just read */
				FD_SET(sockfd, &wset);	/* try and write to socket below */
			}
		}

		if (FD_ISSET(sockfd, &rset)) {
			if ( (n = read(sockfd, friptr, &fr[MAXLINE] - friptr)) < 0) {
				if (errno != EWOULDBLOCK)
					err_sys("read error on socket");

			} else if (n == 0) {
#ifdef	VOL2
				fprintf(stderr, "%s: EOF on socket\n", gf_time());
#endif
				if (stdineof)
					return;		/* normal termination */
				else
					err_quit("str_cli: server terminated prematurely");

			} else {
#ifdef	VOL2
				fprintf(stderr, "%s: read %d bytes from socket\n",
								gf_time(), n);
#endif
				friptr += n;		/* # just read */
				FD_SET(STDOUT_FILENO, &wset);	/* try and write below */
			}
		}
/* end nonb2 */
/* include nonb3 */
		if (FD_ISSET(STDOUT_FILENO, &wset) && ( (n = friptr - froptr) > 0)) {
			if ( (nwritten = write(STDOUT_FILENO, froptr, n)) < 0) {
				if (errno != EWOULDBLOCK)
					err_sys("write error to stdout");

			} else {
#ifdef	VOL2
				fprintf(stderr, "%s: wrote %d bytes to stdout\n",
								gf_time(), nwritten);
#endif
				froptr += nwritten;		/* # just written */
				if (froptr == friptr)
					froptr = friptr = fr;	/* back to beginning of buffer */
			}
		}

		if (FD_ISSET(sockfd, &wset) && ( (n = toiptr - tooptr) > 0)) {
			if ( (nwritten = write(sockfd, tooptr, n)) < 0) {
				if (errno != EWOULDBLOCK)
					err_sys("write error to socket");

			} else {
#ifdef	VOL2
				fprintf(stderr, "%s: wrote %d bytes to socket\n",
								gf_time(), nwritten);
#endif
				tooptr += nwritten;	/* # just written */
				if (tooptr == toiptr) {
					toiptr = tooptr = to;	/* back to beginning of buffer */
					if (stdineof)
						Shutdown(sockfd, SHUT_WR);	/* send FIN */
				}
			}
		}
	}
}
/* end nonb3 */
```

## 16.3 非阻塞 connect


```c
#include	"unp.h"

int connect_nonb(int sockfd, const SA *saptr, socklen_t salen, int nsec)
{
	int				flags, n, error;
	socklen_t		len;
	fd_set			rset, wset;
	struct timeval	tval;

	flags = Fcntl(sockfd, F_GETFL, 0);
	Fcntl(sockfd, F_SETFL, flags | O_NONBLOCK);

	error = 0;
	if ( (n = connect(sockfd, saptr, salen)) < 0)
		if (errno != EINPROGRESS)
			return(-1);

	/* Do whatever we want while the connect is taking place. */

	if (n == 0)
		goto done;	/* connect completed immediately */

	FD_ZERO(&rset);
	FD_SET(sockfd, &rset);
	wset = rset;
	tval.tv_sec = nsec;
	tval.tv_usec = 0;

	if ( (n = Select(sockfd+1, &rset, &wset, NULL,
					 nsec ? &tval : NULL)) == 0) {
		close(sockfd);		/* timeout */
		errno = ETIMEDOUT;
		return(-1);
	}

	if (FD_ISSET(sockfd, &rset) || FD_ISSET(sockfd, &wset)) {
		len = sizeof(error);
		if (getsockopt(sockfd, SOL_SOCKET, SO_ERROR, &error, &len) < 0)
			return(-1);			/* Solaris pending error */
	} else
		err_quit("select error: sockfd not set");

done:
	Fcntl(sockfd, F_SETFL, flags);	/* restore file status flags */

	if (error) {
		close(sockfd);		/* just in case */
		errno = error;
		return(-1);
	}
	return(0);
}
```

# 17 ioctl 操作

```c
#include	"unp.h"
#include	<net/if.h>

int
main(int argc, char **argv)
{
	int		i, sockfd, numif;
	char	buf[1024];

	sockfd = Socket(AF_INET, SOCK_DGRAM, 0);

	numif = 999;
	Ioctl(sockfd, SIOCGIFNUM, &numif);
	printf("numif = %d\n", numif);

	i = ioctl(sockfd, SIOCGHIWAT, &buf);
	printf("i = %d, errno = %d\n", i, errno);
	exit(0);
}
```

# 18. 路由套接口

```c
// 检查 udp 校验和是否开启
#include	"unproute.h"
#include	<netinet/udp.h>
#include	<netinet/ip_var.h>
#include	<netinet/udp_var.h>		/* for UDPCTL_xxx constants */

int
main(int argc, char **argv)
{
	int		mib[4], val;
	size_t	len;

	mib[0] = CTL_NET;
	mib[1] = AF_INET;
	mib[2] = IPPROTO_UDP;
	mib[3] = UDPCTL_CHECKSUM;

	len = sizeof(val);
	Sysctl(mib, 4, &val, &len, NULL, 0);
	printf("udp checksum flag: %d\n", val);

	exit(0);
}
```

# 19. 密钥管理套接口


# 20. 广播(broadcasting)

广播发送的数据报由发送主机某个所在子网上的所有主机接收。广播的劣势在于同一子网上的所有主机都必须处理数据报，
若是UDP数据报则需沿协议栈向上一直处理到UDP层，即使不参与广播应用的主机也不能幸免。要是运行诸如音频、视频等以较高数据速率工作的应用，
这些非必要的处理会给这些主机带来过度的处理负担。我们将在下一章看到多播可以解决本问题，因为多播发送的数据报只会由对相应多播应用感兴趣的主机接收。

竞争状态：多个进程访问共享数据，正确结果取决于进程的执行顺序时，称这些进程处于竞争状态。

```c
// bcast/udpcli06.c
#include	"unp.h"

int
main(int argc, char **argv)
{
	int					sockfd;
	struct sockaddr_in	servaddr;

	if (argc != 2)
		err_quit("usage: udpcli02 <IPaddress>");

	bzero(&servaddr, sizeof(servaddr));
	servaddr.sin_family = AF_INET;
	servaddr.sin_port = htons(13);		/* standard daytime server */
	Inet_pton(AF_INET, argv[1], &servaddr.sin_addr);

	sockfd = Socket(AF_INET, SOCK_DGRAM, 0);

	dg_cli(stdin, sockfd, (SA *) &servaddr, sizeof(servaddr));

	exit(0);
}

// bcast/dgclibcast6.c
#include	"unp.h"

static void	recvfrom_alarm(int);
static int	pipefd[2];

void
dg_cli(FILE *fp, int sockfd, const SA *pservaddr, socklen_t servlen)
{
	int				n, maxfdp1;
	const int		on = 1;
	char			sendline[MAXLINE], recvline[MAXLINE + 1];
	fd_set			rset;
	socklen_t		len;
	struct sockaddr	*preply_addr;

	preply_addr = Malloc(servlen);

	Setsockopt(sockfd, SOL_SOCKET, SO_BROADCAST, &on, sizeof(on)); // SO_BROADCAST 多播选项

	Pipe(pipefd);
	maxfdp1 = max(sockfd, pipefd[0]) + 1;

	FD_ZERO(&rset);

	Signal(SIGALRM, recvfrom_alarm);

	while (Fgets(sendline, MAXLINE, fp) != NULL) {
		Sendto(sockfd, sendline, strlen(sendline), 0, pservaddr, servlen);

		alarm(5);
		for ( ; ; ) {
			FD_SET(sockfd, &rset);
			FD_SET(pipefd[0], &rset);
			if ( (n = select(maxfdp1, &rset, NULL, NULL, NULL)) < 0) {
				if (errno == EINTR)
					continue;
				else
					err_sys("select error");
			}

			if (FD_ISSET(sockfd, &rset)) {
				len = servlen;
				n = Recvfrom(sockfd, recvline, MAXLINE, 0, preply_addr, &len);
				recvline[n] = 0;	/* null terminate */
				printf("from %s: %s",
						Sock_ntop_host(preply_addr, len), recvline);
			}

			if (FD_ISSET(pipefd[0], &rset)) {
				Read(pipefd[0], &n, 1);		/* timer expired */ // 定时器到时
				break;
			}
		}
	}
	free(preply_addr);
}

static void
recvfrom_alarm(int signo)
{
	Write(pipefd[1], "", 1);	/* write one null byte to pipe */
	return;
}
```

# 21. 多播 (multiplcasting)


```c

//mcast/main.c
#include	"unp.h"

void	recv_all(int, socklen_t);
void	send_all(int, SA *, socklen_t);

int
main(int argc, char **argv)
{
	int					sendfd, recvfd;
	const int			on = 1;
	socklen_t			salen;
	struct sockaddr		*sasend, *sarecv;

	if (argc != 3)
		err_quit("usage: sendrecv <IP-multicast-address> <port#>");

	sendfd = Udp_client(argv[1], argv[2], (void **) &sasend, &salen);

	recvfd = Socket(sasend->sa_family, SOCK_DGRAM, 0);

	Setsockopt(recvfd, SOL_SOCKET, SO_REUSEADDR, &on, sizeof(on));

	sarecv = Malloc(salen);
	memcpy(sarecv, sasend, salen);
	Bind(recvfd, sarecv, salen);

	Mcast_join(recvfd, sasend, salen, NULL, 0);
	Mcast_set_loop(sendfd, 0);

	if (Fork() == 0)
		recv_all(recvfd, salen);		/* child -> receives */

	send_all(sendfd, sasend, salen);	/* parent -> sends */
}


// ssntp/sntp_proc.c
#include	"sntp.h"

void
sntp_proc(char *buf, ssize_t n, struct timeval *nowptr)
{
	int				version, mode;
	uint32_t		nsec, useci;
	double			usecf;
	struct timeval	diff;
	struct ntpdata	*ntp;

	if (n < (ssize_t)sizeof(struct ntpdata)) {
		printf("\npacket too small: %d bytes\n", n);
		return;
	}

	ntp = (struct ntpdata *) buf;
	version = (ntp->status & VERSION_MASK) >> 3;
	mode = ntp->status & MODE_MASK;
	printf("\nv%d, mode %d, strat %d, ", version, mode, ntp->stratum);
	if (mode == MODE_CLIENT) {
		printf("client\n");
		return;
	}

	nsec = ntohl(ntp->xmt.int_part) - JAN_1970;
	useci = ntohl(ntp->xmt.fraction);	/* 32-bit integer fraction */
	usecf = useci;				/* integer fraction -> double */
	usecf /= 4294967296.0;		/* divide by 2**32 -> [0, 1.0) */
	useci = usecf * 1000000.0;	/* fraction -> parts per million */

	diff.tv_sec = nowptr->tv_sec - nsec;
	if ( (diff.tv_usec = nowptr->tv_usec - useci) < 0) {
		diff.tv_usec += 1000000;
		diff.tv_sec--;
	}
	useci = (diff.tv_sec * 1000000) + diff.tv_usec;	/* diff in microsec */
	printf("clock difference = %d usec\n", useci);
}
```

# 22 高级 udp 套接口编程

给 udp 增加可靠性：

- 超时和重传: 用户处理丢失的数据报 (比如通过携带时间戳计算 rtt 判断是否需要重传)
- 序列号：供客户验证一个应答是否匹配相应的需求 (发送报文生成序列号)

```c
// rtt/dg_send_recv.c
/* include dgsendrecv1 */
#include	"unprtt.h"
#include	<setjmp.h>

#define	RTT_DEBUG

static struct rtt_info   rttinfo;
static int	rttinit = 0;
static struct msghdr	msgsend, msgrecv;	/* assumed init to 0 */
static struct hdr {
  uint32_t	seq;	/* sequence # */
  uint32_t	ts;		/* timestamp when sent */
} sendhdr, recvhdr;

static void	sig_alrm(int signo);
static sigjmp_buf	jmpbuf;

ssize_t
dg_send_recv(int fd, const void *outbuff, size_t outbytes,
			 void *inbuff, size_t inbytes,
			 const SA *destaddr, socklen_t destlen)
{
	ssize_t			n;
	struct iovec	iovsend[2], iovrecv[2];

	if (rttinit == 0) {
		rtt_init(&rttinfo);		/* first time we're called */
		rttinit = 1;
		rtt_d_flag = 1;
	}

	sendhdr.seq++;
	msgsend.msg_name = destaddr;
	msgsend.msg_namelen = destlen;
	msgsend.msg_iov = iovsend;
	msgsend.msg_iovlen = 2;
	iovsend[0].iov_base = &sendhdr;
	iovsend[0].iov_len = sizeof(struct hdr);
	iovsend[1].iov_base = outbuff;
	iovsend[1].iov_len = outbytes;

	msgrecv.msg_name = NULL;
	msgrecv.msg_namelen = 0;
	msgrecv.msg_iov = iovrecv;
	msgrecv.msg_iovlen = 2;
	iovrecv[0].iov_base = &recvhdr;
	iovrecv[0].iov_len = sizeof(struct hdr);
	iovrecv[1].iov_base = inbuff;
	iovrecv[1].iov_len = inbytes;
/* end dgsendrecv1 */

/* include dgsendrecv2 */
	Signal(SIGALRM, sig_alrm);
	rtt_newpack(&rttinfo);		/* initialize for this packet */

sendagain:
#ifdef	RTT_DEBUG
	fprintf(stderr, "send %4d: ", sendhdr.seq);
#endif
	sendhdr.ts = rtt_ts(&rttinfo);
	Sendmsg(fd, &msgsend, 0);

	alarm(rtt_start(&rttinfo));	/* calc timeout value & start timer */
#ifdef	RTT_DEBUG
	rtt_debug(&rttinfo);
#endif

	if (sigsetjmp(jmpbuf, 1) != 0) {
		if (rtt_timeout(&rttinfo) < 0) {
			err_msg("dg_send_recv: no response from server, giving up");
			rttinit = 0;	/* reinit in case we're called again */
			errno = ETIMEDOUT;
			return(-1);
		}
#ifdef	RTT_DEBUG
		err_msg("dg_send_recv: timeout, retransmitting");
#endif
		goto sendagain;
	}

	do {
		n = Recvmsg(fd, &msgrecv, 0);
#ifdef	RTT_DEBUG
		fprintf(stderr, "recv %4d\n", recvhdr.seq);
#endif
	} while (n < sizeof(struct hdr) || recvhdr.seq != sendhdr.seq);

	alarm(0);			/* stop SIGALRM timer */
		/* 4calculate & store new RTT estimator values */
	rtt_stop(&rttinfo, rtt_ts(&rttinfo) - recvhdr.ts);

	return(n - sizeof(struct hdr));	/* return size of received datagram */
}

static void
sig_alrm(int signo)
{
	siglongjmp(jmpbuf, 1);
}
/* end dgsendrecv2 */

ssize_t
Dg_send_recv(int fd, const void *outbuff, size_t outbytes,
			 void *inbuff, size_t inbytes,
			 const SA *destaddr, socklen_t destlen)
{
	ssize_t	n;

	n = dg_send_recv(fd, outbuff, outbytes, inbuff, inbytes,
					 destaddr, destlen);
	if (n < 0)
		err_quit("dg_send_recv error");

	return(n);
}
```

# 24. 带外数据

带外数据(out-of-band data): 比普通数据具有更高的优先级。tcp 没有真正的带外数据，不过提供紧急模式和紧急指针。

```c
// oob/tcpsend01.c
#include	"unp.h"

int
main(int argc, char **argv)
{
	int		sockfd;

	if (argc != 3)
		err_quit("usage: tcpsend01 <host> <port#>");

	sockfd = Tcp_connect(argv[1], argv[2]);

	Write(sockfd, "123", 3);
	printf("wrote 3 bytes of normal data\n");
	sleep(1);

	Send(sockfd, "4", 1, MSG_OOB); // MSG_OOB 带外标志
	printf("wrote 1 byte of OOB data\n");
	sleep(1);

	Write(sockfd, "56", 2);
	printf("wrote 2 bytes of normal data\n");
	sleep(1);

	Send(sockfd, "7", 1, MSG_OOB);
	printf("wrote 1 byte of OOB data\n");
	sleep(1);

	Write(sockfd, "89", 2);
	printf("wrote 2 bytes of normal data\n");
	sleep(1);

	exit(0);
}


// oob/tcprecv01.c
#include	"unp.h"

int		listenfd, connfd;

void	sig_urg(int);

int
main(int argc, char **argv)
{
	int		n;
	char	buff[100];

	if (argc == 2)
		listenfd = Tcp_listen(NULL, argv[1], NULL);
	else if (argc == 3)
		listenfd = Tcp_listen(argv[1], argv[2], NULL);
	else
		err_quit("usage: tcprecv01 [ <host> ] <port#>");

	connfd = Accept(listenfd, NULL, NULL);

	Signal(SIGURG, sig_urg);
	Fcntl(connfd, F_SETOWN, getpid());

	for ( ; ; ) {
		if ( (n = Read(connfd, buff, sizeof(buff)-1)) == 0) {
			printf("received EOF\n");
			exit(0);
		}
		buff[n] = 0;	/* null terminate */
		printf("read %d bytes: %s\n", n, buff);
	}
}

void
sig_urg(int signo)
{
	int		n;
	char	buff[100];

	printf("SIGURG received\n");
	n = Recv(connfd, buff, sizeof(buff)-1, MSG_OOB);
	buff[n] = 0;		/* null terminate */
	printf("read %d OOB byte: %s\n", n, buff);
}
```

# 25 信号驱动 I/O

信号驱动 IO 是指进程预先告知内核，使得当某个描述字上发生某事时，内核使用信号通知相关进程。
对 TCP 没啥用，没有告诉发生了什么事件。
作者找到的唯一用途：基于 UDP 的 NTP 服务器。

```c
/* include dgecho1 */
#include	"unp.h"

static int		sockfd;

#define	QSIZE	   8		/* size of input queue */
#define	MAXDG	4096		/* max datagram size */

typedef struct {
  void		*dg_data;		/* ptr to actual datagram */
  size_t	dg_len;			/* length of datagram */
  struct sockaddr  *dg_sa;	/* ptr to sockaddr{} w/client's address */
  socklen_t	dg_salen;		/* length of sockaddr{} */
} DG;
static DG	dg[QSIZE];			/* queue of datagrams to process */
static long	cntread[QSIZE+1];	/* diagnostic counter */

static int	iget;		/* next one for main loop to process */
static int	iput;		/* next one for signal handler to read into */
static int	nqueue;		/* # on queue for main loop to process */
static socklen_t clilen;/* max length of sockaddr{} */

static void	sig_io(int);
static void	sig_hup(int);
/* end dgecho1 */

/* include dgecho2 */
void
dg_echo(int sockfd_arg, SA *pcliaddr, socklen_t clilen_arg)
{
	int			i;
	const int	on = 1;
	sigset_t	zeromask, newmask, oldmask;

	sockfd = sockfd_arg;
	clilen = clilen_arg;

	for (i = 0; i < QSIZE; i++) {	/* init queue of buffers */
		dg[i].dg_data = Malloc(MAXDG);
		dg[i].dg_sa = Malloc(clilen);
		dg[i].dg_salen = clilen;
	}
	iget = iput = nqueue = 0;

	Signal(SIGHUP, sig_hup);
	Signal(SIGIO, sig_io);
	Fcntl(sockfd, F_SETOWN, getpid());
	Ioctl(sockfd, FIOASYNC, &on);
	Ioctl(sockfd, FIONBIO, &on);

	Sigemptyset(&zeromask);		/* init three signal sets */
	Sigemptyset(&oldmask);
	Sigemptyset(&newmask);
	Sigaddset(&newmask, SIGIO);	/* signal we want to block */

	Sigprocmask(SIG_BLOCK, &newmask, &oldmask);
	for ( ; ; ) {
		while (nqueue == 0)
			sigsuspend(&zeromask);	/* wait for datagram to process */

			/* 4unblock SIGIO */
		Sigprocmask(SIG_SETMASK, &oldmask, NULL);

		Sendto(sockfd, dg[iget].dg_data, dg[iget].dg_len, 0,
			   dg[iget].dg_sa, dg[iget].dg_salen);

		if (++iget >= QSIZE)
			iget = 0;

			/* 4block SIGIO */
		Sigprocmask(SIG_BLOCK, &newmask, &oldmask);
		nqueue--;
	}
}
/* end dgecho2 */

/* include sig_io */
static void
sig_io(int signo)
{
	ssize_t		len;
	int			nread;
	DG			*ptr;

	for (nread = 0; ; ) {
		if (nqueue >= QSIZE)
			err_quit("receive overflow");

		ptr = &dg[iput];
		ptr->dg_salen = clilen;
		len = recvfrom(sockfd, ptr->dg_data, MAXDG, 0,
					   ptr->dg_sa, &ptr->dg_salen);
		if (len < 0) {
			if (errno == EWOULDBLOCK)
				break;		/* all done; no more queued to read */
			else
				err_sys("recvfrom error");
		}
		ptr->dg_len = len;

		nread++;
		nqueue++;
		if (++iput >= QSIZE)
			iput = 0;

	}
	cntread[nread]++;		/* histogram of # datagrams read per signal */
}
/* end sig_io */

/* include sig_hup */
static void
sig_hup(int signo)
{
	int		i;

	for (i = 0; i <= QSIZE; i++)
		printf("cntread[%d] = %ld\n", i, cntread[i]);
}
/* end sig_hup */
```

# 26 线程

fork的问题：

- 开销大。尽管有copy-on-write
- 父子进程之间传递信息需要 IPC 机制

使用线程开销小，不过需要考虑“同步”问题。

```c
// threads/strclithread2.c
#include	"unpthread.h"

void	*copyto(void *);

static int	sockfd;
static FILE	*fp;
static int	done;

void
str_cli(FILE *fp_arg, int sockfd_arg)
{
	char		recvline[MAXLINE];
	pthread_t	tid;

	sockfd = sockfd_arg;	/* copy arguments to externals */
	fp = fp_arg;

	Pthread_create(&tid, NULL, copyto, NULL);

	while (Readline(sockfd, recvline, MAXLINE) > 0)
		Fputs(recvline, stdout);

	if (done == 0)
		err_quit("server terminated prematurely");
}

void *
copyto(void *arg)
{
	char	sendline[MAXLINE];

	while (Fgets(sendline, MAXLINE, fp) != NULL)
		Writen(sockfd, sendline, strlen(sendline));

	Shutdown(sockfd, SHUT_WR);	/* EOF on stdin, send FIN */

	done = 1;
	return(NULL);
	/* return (i.e., thread terminates) when end-of-file on stdin */
}


// threads/tcpserv02.c
#include	"unpthread.h"

static void	*doit(void *);		/* each thread executes this function */

int
main(int argc, char **argv)
{
	int				listenfd, *iptr;
	thread_t		tid;
	socklen_t		addrlen, len;
	struct sockaddr	*cliaddr;

	if (argc == 2)
		listenfd = Tcp_listen(NULL, argv[1], &addrlen);
	else if (argc == 3)
		listenfd = Tcp_listen(argv[1], argv[2], &addrlen);
	else
		err_quit("usage: tcpserv01 [ <host> ] <service or port>");

	cliaddr = Malloc(addrlen);

	for ( ; ; ) {
		len = addrlen;
		iptr = Malloc(sizeof(int));
		*iptr = Accept(listenfd, cliaddr, &len);
		Pthread_create(&tid, NULL, &doit, iptr);
	}
}

static void *
doit(void *arg)
{
	int		connfd;

	connfd = *((int *) arg);
	free(arg);

	Pthread_detach(pthread_self());
	str_echo(connfd);		/* same function as before */
	Close(connfd);			/* done with connected socket */
	return(NULL);
}
```
