package main

import (
	"bufio"
	"crypto/tls"
	"errors"
	"net"
	"sync"
	"time"
)

/*
redigo 源码阅读和仿写。redigo 和 go-redis 实现差异还蛮大

预备知识：

- redis resp 协议，如何解析和包装
- net, net/url 内置 package 使用方式，socket 编程

redis 协议官方描述：
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


RESP(Redis Serialization Protocol): redis序列化协议

resp 把传输的数据结构分成5种最小单元类型，单元结束后统一加上回车换行符 \r\n

- 单行字符串以+开头
  - +hello\r\n
- 多行字符串以$开头，后缀字符串长度
  - $11\r\nhello world\r\n
- 整数值以: 开头，跟上字符串形式
  - :1024\r\n
- 错误消息，以-开头
  - WRONGTYPE Operation against a key holding the wrong kind of value
- 数组，以 * 开头，后跟数组的长度
  - *3\r\n:1\r\n:2\r\n:3\r\n

特殊：

- NULL 用多行字符串，长度-1
  - $-1\r\n
- 空串，用多行字符串表示，长度0。注意这里的俩\r\n 中间隔的是空串
  - $0\r\n\r\n

### 客户端->服务器

只有一种格式，多行字符串数组

### 服务器->客户端

也是5种基本类型组合

- 单行字符串响应
- 错误响应
- 整数响应
- 多行字符串响应
- 数组响应
- 嵌套



```
$ (printf "PING\r\nPING\r\nPING\r\n"; sleep 1) | nc localhost 6379
+PONG
+PONG
+PONG
```
server 端需要占用内存让命令入队，所以不要一次性发太多命令。



实现重点/疑问：
- 连接池如何实现，如何获取和关闭(归还)连接? Pool
- Conn 如何实现，怎么和 redis-server 交互的？ Conn
- 如何发送和解析 redis resp 协议？ Send/Receive

阅读成果：自己实现（哪怕抄过来）一个最小可用实现
*/

/***** 首先是 redis.go 定义了一系列接口，主要的就是 Conn/ConnWithTimeout，先把关键的抄过来 *****/

// Error represents an error returned in a command reply.
type Error string

func (err Error) Error() string { return string(err) }

// Conn 表示和 redis 服务器的一个 socket 连接。关键方法就这么几个。底层直接借助 net package 实现
type Conn interface {
	// Close closes the connection.
	Close() error
	// Err returns a non-nil value when the connection is not usable.
	Err() error
	// Do sends a command to the server and returns the received reply.
	Do(commandName string, args ...interface{}) (reply interface{}, err error)
	// Send writes the command to the client's output buffer.
	Send(commandName string, args ...interface{}) error
	// Flush flushes the output buffer to the Redis server.
	Flush() error
	// Receive receives a single reply from the Redis server
	Receive() (reply interface{}, err error)
}

// Argument is the interface implemented by an object which wants to control how
// the object is converted to Redis bulk strings.
type Argument interface {
	// RedisArg returns a value to be encoded as a bulk string per the
	// conversions listed in the section 'Executing Commands'.
	// Implementations should typically return a []byte or string.
	RedisArg() interface{}
}

// ConnWithTimeout is an optional interface that allows the caller to override
// a connection's default read timeout. This interface is useful for executing
// the BLPOP, BRPOP, BRPOPLPUSH, XREAD and other commands that block at the
// server.
//
// A connection's default read timeout is set with the DialReadTimeout dial
// option. Applications should rely on the default timeout for commands that do
// not block at the server.
//
// All of the Conn implementations in this package satisfy the ConnWithTimeout
// interface.
//
// Use the DoWithTimeout and ReceiveWithTimeout helper functions to simplify
// use of this interface.
type ConnWithTimeout interface { //增加了超时时间
	Conn
	// Do sends a command to the server and returns the received reply.
	// The timeout overrides the read timeout set when dialing the
	// connection.
	DoWithTimeout(timeout time.Duration, commandName string, args ...interface{}) (reply interface{}, err error)
	// Receive receives a single reply from the Redis server. The timeout
	// overrides the read timeout set when dialing the connection.
	ReceiveWithTimeout(timeout time.Duration) (reply interface{}, err error)
}

// redis 错误处理都用的简单，直接 errors.New
var errTimeoutNotSupported = errors.New("redis: connection does not support ConnWithTimeout")

// DoWithTimeout executes a Redis command with the specified read timeout. If
// the connection does not satisfy the ConnWithTimeout interface, then an error
// is returned.
func DoWithTimeout(c Conn, timeout time.Duration, cmd string, args ...interface{}) (interface{}, error) {
	cwt, ok := c.(ConnWithTimeout)
	if !ok {
		return nil, errTimeoutNotSupported
	}
	return cwt.DoWithTimeout(timeout, cmd, args...)
}

// ReceiveWithTimeout receives a reply with the specified read timeout. If the
// connection does not satisfy the ConnWithTimeout interface, then an error is
// returned.
func ReceiveWithTimeout(c Conn, timeout time.Duration) (interface{}, error) {
	cwt, ok := c.(ConnWithTimeout)
	if !ok {
		return nil, errTimeoutNotSupported
	}
	return cwt.ReceiveWithTimeout(timeout)
}

/***** 既然上边看完了 Conn定义，来看下 conn.go 如何实现 Conn 定义的接口*****/
/* conn.go 里主要就是两个 struct : conn和DialOption 但是函数还是很多，文件挺长 */

var _ ConnWithTimeout = (*conn)(nil) // 确保 conn 实现了 ConnWithTimeout interface

// conn is the low-level implementation of Conn
// 注意所有的『内部』方法使用小写开头，但是没有保护作用，golang 的导出是以 package 为单位的
type conn struct {
	// Shared
	mu      sync.Mutex
	pending int
	err     error
	conn    net.Conn // 底层使用内置的 net 实现，具体先参考  net, net/url
	// Read
	readTimeout time.Duration
	br          *bufio.Reader
	// Write
	writeTimeout time.Duration
	bw           *bufio.Writer
	// Scratch space for formatting argument length.
	// '*' or '$', length, "\r\n"
	lenScratch [32]byte
	// Scratch space for formatting integers and floats.
	numScratch [40]byte
}

// DialOption specifies an option for dialing a Redis server.
// 主要就是一些 Dial redis server 方法的参数配置
type DialOption struct {
	f func(*dialOptions) // 函数作为 field
}

type dialOptions struct {
	readTimeout  time.Duration
	writeTimeout time.Duration
	dialer       *net.Dialer
	dial         func(network, addr string) (net.Conn, error)
	db           int // 使用哪个 redis db
	password     string
	useTLS       bool
	skipVerify   bool
	tlsConfig    *tls.Config
}

// Dial connects to the Redis server at the given network and address using the specified options.
// 这个方法用来 dial redis 服务器，然后返回 socket 连接。虽然很长，但是每个步骤都很清楚
// 此方法可以直接暴露给用户，用来返回一个 redis 连接，NOTE：在后端服务中一般需要使用连接池减少频繁连接开销
func Dial(network, address string, options ...DialOption) (Conn, error) {
	do := dialOptions{
		dialer: &net.Dialer{
			KeepAlive: time.Minute * 5,
		},
	}
	for _, option := range options {
		option.f(&do)
	}
	if do.dial == nil {
		do.dial = do.dialer.Dial
	}

	netConn, err := do.dial(network, address) // redis/conn_test.go|664 col 24| c, err := redis.Dial("tcp", "example.com:6379",
	if err != nil {
		return nil, err
	}

	if do.useTLS { // 建立 tls 握手
		var tlsConfig *tls.Config
		if do.tlsConfig == nil {
			tlsConfig = &tls.Config{InsecureSkipVerify: do.skipVerify}
		} else {
			tlsConfig = cloneTLSConfig(do.tlsConfig)
		}
		if tlsConfig.ServerName == "" {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				netConn.Close()
				return nil, err
			}
			tlsConfig.ServerName = host
		}

		tlsConn := tls.Client(netConn, tlsConfig)
		if err := tlsConn.Handshake(); err != nil {
			netConn.Close()
			return nil, err
		}
		netConn = tlsConn
	}

	c := &conn{ // 包装 内置返回的 netConn 为 上边的 conn struct
		conn:         netConn,
		bw:           bufio.NewWriter(netConn),
		br:           bufio.NewReader(netConn),
		readTimeout:  do.readTimeout, // NOTE: 这里默认是没有超时比较危险的，连接池里一定要增加超时参数
		writeTimeout: do.writeTimeout,
	}

	if do.password != "" { // 连接以后如果需要认证 发送 AUTH
		if _, err := c.Do("AUTH", do.password); err != nil {
			netConn.Close()
			return nil, err
		}
	}

	if do.db != 0 { // 选择使用 redis 哪个 db，默认都是 0
		if _, err := c.Do("SELECT", do.db); err != nil {
			netConn.Close()
			return nil, err
		}
	}

	return c, nil
}

// NewConn returns a new Redigo connection for the given net connection.
// 然后是 Conn 相关接口的实现，这里 Conn 代表一个 redis tcp socket 连接，主要是发送和接收（包括解析 redis 协议)
func NewConn(netConn net.Conn, readTimeout, writeTimeout time.Duration) Conn {
	return &conn{
		conn:         netConn,
		bw:           bufio.NewWriter(netConn),
		br:           bufio.NewReader(netConn),
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
	}
}

/***** 接下来是 conn 实现接口(Conn) 的方法，重要的是 Do/Send/Receive *****/
// 从关键方法『自顶向下』看代码，涉及到的地方可以跳转快速了解实现

// 先来看发送Send，NOTE： 这里虽然叫做 send，其实并没有真正发送，而是写到了缓冲区
func (c *conn) Send(cmd string, args ...interface{}) error {
	c.mu.Lock()
	c.pending++ // c.pending += 1  这里我的 golint 报错了，把源代码 pending += 1 改成了 pending++
	c.mu.Unlock()
	if c.writeTimeout != 0 {
		c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
	}
	if err := c.writeCommand(cmd, args); err != nil {
		return c.fatal(err)
	}
	return nil
}

// 到这里先要熟悉 resp 协议才能看懂，最上边的注释已经列出来了
// 上边 Send 用到了 if err := c.writeCommand(cmd, args); err != nil {，来看看 这个方法
func (c *conn) writeCommand(cmd string, args []interface{}) error {
	// 又是三个 write 方法，一会一个个看下是啥意思 writeLen/writeString/writeArg
	// NOTE: Client -> Server 只有一种格式，就是多行字符串数组，数组以 * 开头，后面跟数组长度
	// - 数组，以 * 开头，后跟数组的长度  *3\r\n:1\r\n:2\r\n:3\r\n
	c.writeLen('*', 1+len(args))
	if err := c.writeString(cmd); err != nil {
		return err
	}
	for _, arg := range args {
		if err := c.writeArg(arg, true); err != nil {
			return err
		}
	}
	return nil
}

// 按照 resp 协议描述写入长度
func (c *conn) writeLen(prefix byte, n int) error {
	//resp 把传输的数据结构分成5种最小单元类型，单元结束后统一加上回车换行符 \r\n
	c.lenScratch[len(c.lenScratch)-1] = '\n'
	c.lenScratch[len(c.lenScratch)-2] = '\r'
	i := len(c.lenScratch) - 3
	for { // 这里是把数字转成字母的形式，然后填充，需要倒过来处理字符串
		c.lenScratch[i] = byte('0' + n%10)
		i-- // i -= 1
		n = n / 10
		if n == 0 {
			break
		}
	}
	c.lenScratch[i] = prefix
	_, err := c.bw.Write(c.lenScratch[i:]) // 首先写入长度
	return err
}

// send 第二个用到的是这个方法， resp 规定 发送多行字符串以 $ 开头
func (c *conn) writeString(s string) error {
	c.writeLen('$', len(s))
	c.bw.WriteString(s)
	_, err := c.bw.WriteString("\r\n")
	return err
}

func (c *conn) Do(cmd string, args ...interface{}) (interface{}, error) {
	return c.DoWithTimeout(c.readTimeout, cmd, args...)
}

func (c *conn) DoWithTimeout(readTimeout time.Duration, cmd string, args ...interface{}) (interface{}, error) {
	c.mu.Lock()
	pending := c.pending
	c.pending = 0
	c.mu.Unlock()

	if cmd == "" && pending == 0 {
		return nil, nil
	}

	if c.writeTimeout != 0 {
		c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
	}

	if cmd != "" {
		if err := c.writeCommand(cmd, args); err != nil {
			return nil, c.fatal(err)
		}
	}

	if err := c.bw.Flush(); err != nil {
		return nil, c.fatal(err)
	}

	var deadline time.Time
	if readTimeout != 0 {
		deadline = time.Now().Add(readTimeout)
	}
	c.conn.SetReadDeadline(deadline)

	if cmd == "" {
		reply := make([]interface{}, pending)
		for i := range reply {
			r, e := c.readReply()
			if e != nil {
				return nil, c.fatal(e)
			}
			reply[i] = r
		}
		return reply, nil
	}

	var err error
	var reply interface{}
	for i := 0; i <= pending; i++ {
		var e error
		if reply, e = c.readReply(); e != nil {
			return nil, c.fatal(e)
		}
		if e, ok := reply.(Error); ok && err == nil {
			err = e
		}
	}
	return reply, err
}
