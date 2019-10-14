package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
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

func cloneTLSConfig(cfg *tls.Config) *tls.Config {
	return cfg.Clone()
}

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

func (c *conn) Close() error { // 被 ConnPoll 调用
	c.mu.Lock()
	err := c.err
	if c.err == nil {
		c.err = errors.New("redigo: closed")
		err = c.conn.Close()
	}
	c.mu.Unlock()
	return err
}

func (c *conn) Err() error {
	c.mu.Lock()
	err := c.err
	c.mu.Unlock()
	return err
}

func (c *conn) fatal(err error) error {
	c.mu.Lock()
	if c.err == nil {
		c.err = err
		// Close connection to force errors on subsequent calls and to unblock
		// other reader or writer.
		c.conn.Close()
	}
	c.mu.Unlock()
	return err
}

/***** 接下来是 conn 实现接口(Conn) 的方法，重要的是 Do/Send/Receive *****/
// 从关键方法『自顶向下』看代码，涉及到的地方可以跳转快速了解实现

// 先来看发送Send，NOTE： 这里虽然叫做 send，其实并没有真正发送，而是写到了缓冲区, Flush() 才发送
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
	_, err := c.bw.WriteString("\r\n") //写入开头和结尾
	return err
}

// 以下三个方法写入不同的数据类型
func (c *conn) writeBytes(p []byte) error {
	c.writeLen('$', len(p))
	c.bw.Write(p)
	_, err := c.bw.WriteString("\r\n")
	return err
}

func (c *conn) writeInt64(n int64) error {
	return c.writeBytes(strconv.AppendInt(c.numScratch[:0], n, 10))
}

func (c *conn) writeFloat64(n float64) error {
	return c.writeBytes(strconv.AppendFloat(c.numScratch[:0], n, 'g', -1, 64))
}

// 接下来是写入命令。调用各种 write 来写入参数
func (c *conn) writeArg(arg interface{}, argumentTypeOK bool) (err error) {
	switch arg := arg.(type) {
	case string:
		return c.writeString(arg)
	case []byte:
		return c.writeBytes(arg)
	case int:
		return c.writeInt64(int64(arg))
	case int64:
		return c.writeInt64(arg)
	case float64:
		return c.writeFloat64(arg)
	case bool:
		if arg {
			return c.writeString("1")
		}
		return c.writeString("0")
	case nil:
		return c.writeString("")
	case Argument:
		if argumentTypeOK {
			return c.writeArg(arg.RedisArg(), false) // 递归写入
		}
		// See comment in default clause below.
		var buf bytes.Buffer
		fmt.Fprint(&buf, arg)
		return c.writeBytes(buf.Bytes())
	default:
		// This default clause is intended to handle builtin numeric types.
		// The function should return an error for other types, but this is not
		// done for compatibility with previous versions of the package.
		var buf bytes.Buffer
		fmt.Fprint(&buf, arg)
		return c.writeBytes(buf.Bytes())
	}
}

// 上边看完了如何写入数据，然后就是 发送，之前的 send不是真正 send，而是写入到缓冲区，直到调用 Flush 才发送到 client
// 一般使用方式就是  client.send() （可以多次send) 之后调用 Flush 方法写入放到 client。之后就可以使用 Receive
// 读取返回结果了。 send() -> flush() -> receive()
func (c *conn) Flush() error {
	if c.writeTimeout != 0 {
		c.conn.SetWriteDeadline(time.Now().Add(c.writeTimeout))
	}
	if err := c.bw.Flush(); err != nil {
		return c.fatal(err)
	}
	return nil
}

// 读取redis 返回的结果
func (c *conn) Receive() (interface{}, error) {
	return c.ReceiveWithTimeout(c.readTimeout)
}

func (c *conn) ReceiveWithTimeout(timeout time.Duration) (reply interface{}, err error) {
	var deadline time.Time
	if timeout != 0 {
		deadline = time.Now().Add(timeout)
	}
	c.conn.SetReadDeadline(deadline)

	if reply, err = c.readReply(); err != nil { // 主要来看下 readReply 方法如何获取返回数据
		return nil, c.fatal(err)
	}
	// When using pub/sub, the number of receives can be greater than the
	// number of sends. To enable normal use of the connection after
	// unsubscribing from all channels, we do not decrement pending to a
	// negative value.
	//
	// The pending field is decremented after the reply is read to handle the
	// case where Receive is called before Send.
	c.mu.Lock()
	if c.pending > 0 { // NOTE: 注意 send 的时候增加1，这里减去1
		c.pending--
	}
	c.mu.Unlock()
	if err, ok := reply.(Error); ok {
		return nil, err
	}
	return
}

// 下边是读取 redis 返回数据，并且解析的代码。先再去熟悉下 resp 协议
var (
	okReply   interface{} = "OK"
	pongReply interface{} = "PONG"
)

type protocolError string

func (pe protocolError) Error() string {
	return fmt.Sprintf("redigo: %s (possible server error or unsupported concurrent read by application)", string(pe))
}

func (c *conn) readReply() (interface{}, error) {
	line, err := c.readLine() // 一会再去看下 readline 方法
	if err != nil {
		return nil, err
	}
	if len(line) == 0 {
		return nil, protocolError("short response line")
	}
	switch line[0] { // 以下是解析 resp 返回协议的代码，5种类型
	case '+': // - 单行字符串以+开头
		switch {
		case len(line) == 3 && line[1] == 'O' && line[2] == 'K':
			// Avoid allocation for frequent "+OK" response.
			return okReply, nil
		case len(line) == 5 && line[1] == 'P' && line[2] == 'O' && line[3] == 'N' && line[4] == 'G':
			// Avoid allocation in PING command benchmarks :)
			return pongReply, nil
		default:
			return string(line[1:]), nil
		}
	case '-': // - 错误消息，以-开头
		return Error(string(line[1:])), nil
	case ':': // - 整数值以: 开头，跟上字符串形式
		return parseInt(line[1:])
	case '$': // - 多行字符串以$开头，后缀字符串长度
		n, err := parseLen(line[1:])
		if n < 0 || err != nil {
			return nil, err
		}
		p := make([]byte, n)
		_, err = io.ReadFull(c.br, p)
		if err != nil {
			return nil, err
		}
		if line, err := c.readLine(); err != nil {
			return nil, err
		} else if len(line) != 0 {
			return nil, protocolError("bad bulk string format")
		}
		return p, nil
	case '*': // - 数组，以 * 开头，后跟数组的长度
		n, err := parseLen(line[1:])
		if n < 0 || err != nil {
			return nil, err
		}
		r := make([]interface{}, n)
		for i := range r {
			r[i], err = c.readReply()
			if err != nil {
				return nil, err
			}
		}
		return r, nil
	}
	return nil, protocolError("unexpected response line")
}

// 看下上边用到的 readLine 方法
// readLine reads a line of input from the RESP stream.
func (c *conn) readLine() ([]byte, error) {
	// To avoid allocations, attempt to read the line using ReadSlice. This
	// call typically succeeds. The known case where the call fails is when
	// reading the output from the MONITOR command.
	p, err := c.br.ReadSlice('\n')
	if err == bufio.ErrBufferFull {
		// The line does not fit in the bufio.Reader's buffer. Fall back to
		// allocating a buffer for the line.
		buf := append([]byte{}, p...)
		for err == bufio.ErrBufferFull {
			p, err = c.br.ReadSlice('\n')
			buf = append(buf, p...)
		}
		p = buf
	}
	if err != nil {
		return nil, err
	}
	i := len(p) - 2
	if i < 0 || p[i] != '\r' { // 不合法的 redis 协议
		return nil, protocolError("bad response line terminator")
	}
	return p[:i], nil // 去掉了后缀的 \r\n
}

// 注意在readReply 中用到了很多 parse 函数，用来解析返回值，下边看下几个函数
// parseLen parses bulk string and array lengths.
func parseLen(p []byte) (int, error) {
	if len(p) == 0 {
		return -1, protocolError("malformed length")
	}

	if p[0] == '-' && len(p) == 2 && p[1] == '1' {
		// handle $-1 and $-1 null replies.
		return -1, nil
	}

	var n int
	for _, b := range p { // 类似这种代码都是用来 转 bytes -> 数字的
		n *= 10
		if b < '0' || b > '9' {
			return -1, protocolError("illegal bytes in length")
		}
		n += int(b - '0')
	}

	return n, nil
}

// parseInt parses an integer reply.
func parseInt(p []byte) (interface{}, error) {
	if len(p) == 0 {
		return 0, protocolError("malformed integer")
	}

	var negate bool
	if p[0] == '-' {
		negate = true
		p = p[1:]
		if len(p) == 0 {
			return 0, protocolError("malformed integer")
		}
	}

	var n int64
	for _, b := range p {
		n *= 10
		if b < '0' || b > '9' {
			return 0, protocolError("illegal bytes in length")
		}
		n += int64(b - '0')
	}

	if negate {
		n = -n
	}
	return n, nil
}

/*******
到这里流程就清楚了：

client.Send() 写入缓冲区，按照resp格式编码
client.Flush() 才写入到socket，这个时候真正发送给客户端
client.Receive() 接受并且解析客户端命令，主要是resp 5种协议格式解析，或者对应的数据

这里需要注意：如果多次 Send 了 Receive 的次数要和 Send 次数对应


每次都 send/flush/receive 比较麻烦，为了简化，redigo 还提供了 DO 方法来串起来这些操作，你可以看到 Do方法就是三者结合
NOTE：注意 go 的方法都提供了 DoWithTimeout 类似的方法，默认的socket没有设置超时线上可能有问题，实际使用必须设置超时毫秒数
*******/

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

/************
到这里关于 conn 的操作实际上就了解差不多了，现在写一段代码来测试以下。
通过内置的 net.Dial 返回一个 tcp 的 socket Conn，传给 NewConn，然后调用 send/flush/receive 测试一下
************/

func testRedigoConn() {
	// 搜了下代码发现 NewConn 函数其实并没有用到，好像只有 Dial 用到了
	conn, err := net.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		return
	}
	redisConn := NewConn(conn, time.Second, time.Second)
	defer redisConn.Close()
	redisConn.Send("SET", "test", "hehe")
	redisConn.Send("GET", "test")
	redisConn.Flush()
	res, err := redisConn.Receive()
	fmt.Println(res, err)
	res, err = redisConn.Receive()
	fmt.Println(string(res.([]byte)), err)

	// ^_^: 测试可以用（废话，都是copy 的代码)
}

// func main() {
// 	testRedigoConn()
// }

/******
上边看完了一个 tcp conn 如何和 redis server 交互的，如何解析协议的。之后看下如何实现一个 连接池(socket conn pool)
pool 的主要作用是减少频繁的 tcp 创建和开销，实现 tcp socket 被不同客户端复用，从而提升redis 交互效率
pool 的实现一般是使用队列/链表等
******/

// Pool 先来看下 redigo Pool 的实现定义，redigo 使用的是双链表来实现的。这里是个 struct 而不是接口
type Pool struct {
	// Dial is an application supplied function for creating and configuring a
	// connection.
	//
	// The connection returned from Dial must not be in a special state
	// (subscribed to pubsub channel, transaction started, ...).
	Dial func() (Conn, error)

	// TestOnBorrow is an optional application supplied function for checking
	// the health of an idle connection before the connection is used again by
	// the application. Argument t is the time that the connection was returned
	// to the pool. If the function returns an error, then the connection is
	// closed.
	TestOnBorrow func(c Conn, t time.Time) error

	// Maximum number of idle connections in the pool.
	MaxIdle int

	// Maximum number of connections allocated by the pool at a given time.
	// When zero, there is no limit on the number of connections in the pool.
	MaxActive int

	// Close connections after remaining idle for this duration. If the value
	// is zero, then idle connections are not closed. Applications should set
	// the timeout to a value less than the server's timeout.
	IdleTimeout time.Duration

	// If Wait is true and the pool is at the MaxActive limit, then Get() waits
	// for a connection to be returned to the pool before returning.
	Wait bool

	// Close connections older than this duration. If the value is zero, then
	// the pool does not close connections based on age.
	MaxConnLifetime time.Duration

	chInitialized uint32 // set to 1 when field ch is initialized

	mu     sync.Mutex    // mu protects the following fields
	closed bool          // set to true when the pool is closed.
	active int           // the number of open connections in the pool
	ch     chan struct{} // limits open connections when p.Wait is true。用 bufferd channel 来限制连接数的
	idle   idleList      // idle connections
}

// 代码给了一个示例来演示 Pool 的用法， 初始化函数不推荐使用了，直接用 struct 构造
// pool := &redis.Pool{
//   // Other pool configuration not shown in this example.
//   Dial: func () (redis.Conn, error) {
//     c, err := redis.Dial("tcp", server)
//     if err != nil {
//       return nil, err
//     }
//     if _, err := c.Do("AUTH", password); err != nil {
//       c.Close()
//       return nil, err
//     }
//     if _, err := c.Do("SELECT", db); err != nil {
//       c.Close()
//       return nil, err
//     }
//     return c, nil
//   },
// }
//

/*
先来看下 pool 里边一个特殊的结构 idleList 是如何实现的，其他类型都是内置结构
这个类型用来实现连接池里 conn 的放入和拿出
*/

type idleList struct { //实现的一个双端链表，front,back 分别指向 头和尾。注意本身没加上线程安全控制，在调用点加上的
	count       int
	front, back *poolConn
}

type poolConn struct {
	c          Conn
	t          time.Time // 放入时间, put 方法会更新t
	created    time.Time
	next, prev *poolConn
}

func (l *idleList) Count() int {
	return l.count
}

func (l *idleList) pushFront(pc *poolConn) {
	pc.next = l.front
	pc.prev = nil
	if l.count == 0 {
		l.back = pc
	} else {
		l.front.prev = pc
	}
	l.front = pc
	l.count++
	return
}

func (l *idleList) popFront() {
	pc := l.front
	l.count--
	if l.count == 0 {
		l.front, l.back = nil, nil
	} else {
		pc.next.prev = nil
		l.front = pc.next
	}
	pc.next, pc.prev = nil, nil
}

func (l *idleList) popBack() {
	pc := l.back
	l.count--
	if l.count == 0 {
		l.front, l.back = nil, nil
	} else {
		pc.prev.next = nil
		l.back = pc.prev
	}
	pc.next, pc.prev = nil, nil
}

// Get gets a connection. The application must close the returned connection.
// This method always returns a valid connection so that applications can defer
// error handling to the first use of the connection. If there is an error
// getting an underlying connection, then the connection Err, Do, Send, Flush
// and Receive methods return that error.
func (p *Pool) Get() Conn { // 先来看下 get 方法从池里获取一个 conn
	pc, err := p.get(nil)
	if err != nil {
		fmt.Println(err)
		return errorConn{err}
	}
	return &activeConn{p: p, pc: pc}
}

type errorConn struct{ err error }

func (ec errorConn) Do(string, ...interface{}) (interface{}, error) { return nil, ec.err }
func (ec errorConn) DoWithTimeout(time.Duration, string, ...interface{}) (interface{}, error) {
	return nil, ec.err
}
func (ec errorConn) Send(string, ...interface{}) error                     { return ec.err }
func (ec errorConn) Err() error                                            { return ec.err }
func (ec errorConn) Close() error                                          { return nil }
func (ec errorConn) Flush() error                                          { return ec.err }
func (ec errorConn) Receive() (interface{}, error)                         { return nil, ec.err }
func (ec errorConn) ReceiveWithTimeout(time.Duration) (interface{}, error) { return nil, ec.err }

// 活动连接
type activeConn struct {
	p     *Pool // 活动连接所在的 Pool
	pc    *poolConn
	state int // 参考下边的 state 定义, 表示现在的链接使用哪个
}

var (
	sentinel     []byte
	sentinelOnce sync.Once
)
var ErrPoolExhausted = errors.New("redigo: connection pool exhausted")

var (
	errPoolClosed = errors.New("redigo: connection pool closed")
	errConnClosed = errors.New("redigo: connection closed")
)

func initSentinel() {
	p := make([]byte, 64)
	if _, err := rand.Read(p); err == nil {
		sentinel = p
	} else {
		h := sha1.New()
		io.WriteString(h, "Oops, rand failed. Use time instead.")
		io.WriteString(h, strconv.FormatInt(time.Now().UnixNano(), 10))
		sentinel = h.Sum(nil)
	}
}

// 看下 activeConn 的几个方法
// Close 这个方法用来放回连接池，根据状态先向 redis 发送中断命令，然后放回到连接池，注意并不会关闭 socket
func (ac *activeConn) Close() error {
	pc := ac.pc
	if pc == nil {
		return nil
	}
	ac.pc = nil

	if ac.state&connectionMultiState != 0 { //这里需要根据状态来做一个清理，不能直接关闭 socket
		pc.c.Send("DISCARD")
		ac.state &^= (connectionMultiState | connectionWatchState)
	} else if ac.state&connectionWatchState != 0 {
		pc.c.Send("UNWATCH")
		ac.state &^= connectionWatchState
	}
	if ac.state&connectionSubscribeState != 0 {
		pc.c.Send("UNSUBSCRIBE")
		pc.c.Send("PUNSUBSCRIBE")
		// To detect the end of the message stream, ask the server to echo
		// a sentinel value and read until we see that value.
		sentinelOnce.Do(initSentinel)
		pc.c.Send("ECHO", sentinel)
		pc.c.Flush()
		for {
			p, err := pc.c.Receive()
			if err != nil {
				break
			}
			if p, ok := p.([]byte); ok && bytes.Equal(p, sentinel) {
				ac.state &^= connectionSubscribeState
				break
			}
		}
	}
	pc.c.Do("")
	ac.p.put(pc, ac.state != 0 || pc.c.Err() != nil) //NOTE: 这一步又重新放回连接池
	return nil
}

func (ac *activeConn) Err() error {
	pc := ac.pc
	if pc == nil {
		return errConnClosed
	}
	return pc.c.Err()
}

func (ac *activeConn) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	pc := ac.pc
	if pc == nil {
		return nil, errConnClosed
	}
	ci := lookupCommandInfo(commandName)
	ac.state = (ac.state | ci.Set) &^ ci.Clear
	return pc.c.Do(commandName, args...)
}

func (ac *activeConn) DoWithTimeout(timeout time.Duration, commandName string, args ...interface{}) (reply interface{}, err error) {
	pc := ac.pc
	if pc == nil {
		return nil, errConnClosed
	}
	cwt, ok := pc.c.(ConnWithTimeout)
	if !ok {
		return nil, errTimeoutNotSupported
	}
	ci := lookupCommandInfo(commandName)
	ac.state = (ac.state | ci.Set) &^ ci.Clear
	return cwt.DoWithTimeout(timeout, commandName, args...)
}

func (ac *activeConn) Send(commandName string, args ...interface{}) error {
	pc := ac.pc
	if pc == nil {
		return errConnClosed
	}
	ci := lookupCommandInfo(commandName)
	ac.state = (ac.state | ci.Set) &^ ci.Clear
	return pc.c.Send(commandName, args...)
}

func (ac *activeConn) Flush() error {
	pc := ac.pc
	if pc == nil {
		return errConnClosed
	}
	return pc.c.Flush()
}

func (ac *activeConn) Receive() (reply interface{}, err error) {
	pc := ac.pc
	if pc == nil {
		return nil, errConnClosed
	}
	return pc.c.Receive()
}

func (ac *activeConn) ReceiveWithTimeout(timeout time.Duration) (reply interface{}, err error) {
	pc := ac.pc
	if pc == nil {
		return nil, errConnClosed
	}
	cwt, ok := pc.c.(ConnWithTimeout)
	if !ok {
		return nil, errTimeoutNotSupported
	}
	return cwt.ReceiveWithTimeout(timeout)
}

var nowFunc = time.Now // for testing

// 上边看到 Get 方法调用了内部方法 get 来获取。先清理过期conn，然后返回一个连接池的
// conn，如果连接池空了则新建一个conn返回
// get prunes stale connections and returns a connection from the idle list or
// creates a new connection.
func (p *Pool) get(ctx interface { // 类似匿名结构体那种定义方式,匿名接口的定义直接写到了参数里
	Done() <-chan struct{}
	Err() error
}) (*poolConn, error) {

	// Handle limit for p.Wait == true.
	if p.Wait && p.MaxActive > 0 { // wait字段如果是 true，会限制达到最大MaxActive连接数以后，Get 方法等待
		p.lazyInit() //lazyInit 用来设置这个字段 p.chInitialized 和初始化 p.ch (buffed channel 限制连接数)。
		if ctx == nil {
			<-p.ch // 这里直接用 buffed channel 来做限制连接数，可以直接使用channel 的 block 功能，而不是轮询重试
		} else {
			select {
			case <-p.ch:
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	p.mu.Lock() // 这里开始加锁，看起 Lock 范围还挺大的

	// Prune stale connections at the back of the idle list. 清理长期不用的闲置连接，根据 IdleTimeout 判断
	if p.IdleTimeout > 0 {
		n := p.idle.count // 连接池连接数量。之后遍历n 次，从最后删除。链表尾部的 最后使用时间 t 小，闲置越久
		for i := 0; i < n && p.idle.back != nil && p.idle.back.t.Add(p.IdleTimeout).Before(nowFunc()); i++ {
			pc := p.idle.back
			p.idle.popBack()
			p.mu.Unlock()
			pc.c.Close()
			p.mu.Lock()
			p.active--
		}
	}

	// Get idle connection from the front of idle list.
	for p.idle.front != nil {
		pc := p.idle.front
		p.idle.popFront() // 每次从头部取一个链接
		p.mu.Unlock()
		if (p.TestOnBorrow == nil || p.TestOnBorrow(pc.c, pc.t) == nil) &&
			(p.MaxConnLifetime == 0 || nowFunc().Sub(pc.created) < p.MaxConnLifetime) {
			return pc, nil
		}
		pc.c.Close()
		p.mu.Lock()
		p.active--
	}

	// Check for pool closed before dialing a new connection.
	if p.closed {
		p.mu.Unlock()
		return nil, errors.New("redigo: get on closed pool")
	}

	// Handle limit for p.Wait == false.
	if !p.Wait && p.MaxActive > 0 && p.active >= p.MaxActive {
		p.mu.Unlock()
		return nil, ErrPoolExhausted
	}

	p.active++
	p.mu.Unlock()
	c, err := p.Dial()
	if err != nil {
		c = nil
		p.mu.Lock()
		p.active--
		if p.ch != nil && !p.closed {
			p.ch <- struct{}{} // 用来限制连接数大小
		}
		p.mu.Unlock()
	}
	fmt.Println("After GET count is", p.idle.Count(), p.active)
	return &poolConn{c: c, created: nowFunc()}, err
}

// get 用到这个方法，主要作用是往 p.ch 中塞入 MaxActive 个 struct{}{}
func (p *Pool) lazyInit() {
	// Fast path.
	if atomic.LoadUint32(&p.chInitialized) == 1 {
		return
	}
	// Slow path.
	p.mu.Lock()
	if p.chInitialized == 0 {
		p.ch = make(chan struct{}, p.MaxActive) // bufferd channel
		if p.closed {
			close(p.ch)
		} else {
			for i := 0; i < p.MaxActive; i++ {
				p.ch <- struct{}{}
			}
		}
		atomic.StoreUint32(&p.chInitialized, 1)
	}
	p.mu.Unlock()
}

// 看完 get 一个 conn的，再来看下放回一个 conn的。注意这个是私有方法，不对外暴露。看下一个 Close 如何调用 put
func (p *Pool) put(pc *poolConn, forceClose bool) error {
	p.mu.Lock()
	if !p.closed && !forceClose {
		pc.t = nowFunc()
		p.idle.pushFront(pc) // 注意每次是放入头(get也是从头部获取)，更新 tc.t。所有 from -> back ，头部到尾部 t 依次减小
		if p.idle.count > p.MaxIdle {
			pc = p.idle.back
			p.idle.popBack() // 如果超过了最大闲置数量，就从尾部踢出一个最久没用过的
		} else {
			pc = nil //没有满
		}
	}

	if pc != nil { // 关闭被踢出的连接
		p.mu.Unlock()
		pc.c.Close()
		p.mu.Lock()
		p.active--
	}

	if p.ch != nil && !p.closed {
		p.ch <- struct{}{}
	}
	p.mu.Unlock()
	fmt.Println("After PUT count is", p.idle.Count(), p.active)
	return nil
}

// Close releases the resources used by the pool.  TODO(wnn) 如何放回连接池的？-> 看 activeConn.Close
func (p *Pool) Close() error { // 清理链表；关闭连接
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	p.active -= p.idle.count
	pc := p.idle.front
	p.idle.count = 0
	p.idle.front, p.idle.back = nil, nil
	if p.ch != nil {
		close(p.ch)
	}
	p.mu.Unlock()
	for ; pc != nil; pc = pc.next { // 注意 activeConn.Close 是放回连接池，Pool 的 close 是真的关闭连接
		pc.c.Close()
	}
	return nil
}

func testPool() {
	pool := Pool{
		// Other pool configuration not shown in this example.
		Dial: func() (Conn, error) {
			c, err := Dial("tcp", ":6379")
			if err != nil {
				return nil, err
			}
			return c, nil
		},
	}
	conn := pool.Get()
	defer conn.Close()
	conn2 := pool.Get()
	defer conn2.Close()
}

func main() {
	testPool()
}
