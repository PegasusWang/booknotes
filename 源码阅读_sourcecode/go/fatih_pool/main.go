/*
"github.com/fatih/pool"
关于实现一个连接池"github.com/fatih/pool" 代码阅读

连接池：
conn pool,通过实现 tcp 连接复用，防止频繁创建和销毁连接的开销，
一般在 redis/mysql 实现的 client 中常见。

go 里常见的使用 channel, 双端链表等（redigo)，slice(go-redis use []*Conn) 来实现连接池。
*/

package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

var (
	// ErrClosed is the error resulting if the pool is closed via pool.Close().
	ErrClosed = errors.New("pool is closed")
)

// Pool interface describes a pool implementation. A pool should have maximum
// capacity. An ideal pool is threadsafe and easy to use.
// 先定义 Pool 接口，包括 get/close/len
type Pool interface {
	Get() (net.Conn, error)
	Close() //关闭连接池中的连接，不同于 PoolConn 的 close 方法是放回(或者关闭)连接池
	Len() int
}

// 基于 bufferd channel 实现 Pool interface
type channelPool struct {
	mu      sync.RWMutex
	conns   chan net.Conn
	factory Factory // 定义工厂方法，决定如何初始化conn
}

// Factory conn
type Factory func() (net.Conn, error)

// NewChannelPool create new channelPool
func NewChannelPool(initialCap, maxCap int, factory Factory) (Pool, error) {
	if initialCap < 0 || maxCap <= 0 || initialCap > maxCap {
		return nil, errors.New("invalid capacity")
	}
	c := &channelPool{
		conns:   make(chan net.Conn, maxCap),
		factory: factory,
	}
	// 创建 initialCap 个连接，如果出错，直接 close
	for i := 0; i < initialCap; i++ {
		conn, err := factory()
		if err != nil {
			c.Close()
			return nil, fmt.Errorf("factor is not able to fill the pool: %s", err)
		}
		c.conns <- conn
	}
	return c, nil
}

func (c *channelPool) getConnsAndFactory() (chan net.Conn, Factory) {
	c.mu.RLock()
	conns := c.conns
	factory := c.factory
	c.mu.RUnlock()
	return conns, factory
}
func (c *channelPool) Get() (net.Conn, error) {
	conns, factory := c.getConnsAndFactory()
	if conns == nil {
		return nil, ErrClosed
	}
	select {
	case conn := <-conns:
		if conn == nil {
			return nil, ErrClosed
		}
		return c.wrapConn(conn), nil // 包装 conn
	default:
		conn, err := factory() // 没有就创建
		if err != nil {
			return nil, err
		}
		return c.wrapConn(conn), nil
	}
}

func (c *channelPool) put(conn net.Conn) error {
	if conn == nil {
		return errors.New("connection is nil. rejecting")
	}
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conns == nil {
		return conn.Close() //如果连接池关闭了，直接 关闭conn
	}
	select {
	case c.conns <- conn:
		return nil
	default:
		return conn.Close() //连接池满了，直接关闭 conn
	}
}

// 关闭连接池中的连接，关闭channel
func (c *channelPool) Close() {
	c.mu.Lock()
	conns := c.conns
	c.conns = nil
	c.factory = nil
	c.mu.Unlock()
	if conns == nil {
		return
	}
	close(conns)              //close channel
	for conn := range conns { // close all conn
		conn.Close()
	}
}

func (c *channelPool) Len() int {
	conns, _ := c.getConnsAndFactory() // 可以直接返回 len(c.conns) 么?
	return len(conns)
}

// PoolConn 包装 net.Conn
type PoolConn struct {
	net.Conn
	mu       sync.RWMutex
	c        *channelPool //所在的连接池
	unusable bool
}

// Close 关闭 or 放回连接池
func (p *PoolConn) Close() error {
	fmt.Println("close===========")
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.unusable {
		if p.Conn != nil {
			return p.Conn.Close()
		}
		return nil
	}
	return p.c.put(p.Conn) // 放回连接池
}

// MarkUnusable 标记这个连接不在使用，关闭的时候直接 close，而不是放回连接池
func (p *PoolConn) MarkUnusable() {
	p.mu.Lock()
	p.unusable = true
	p.mu.Unlock()
}

// 包装一个 net.Conn 为  PoolConn，注意是 channelPool 而不是 PoolConn 的方法
func (c *channelPool) wrapConn(conn net.Conn) net.Conn {
	p := &PoolConn{c: c}
	p.Conn = conn
	return p
}

func main() {
	// 测试代码
	factor := func() (net.Conn, error) { return net.Dial("tcp", "127.0.0.1:6379") }
	pool, err := NewChannelPool(2, 10, factor)
	fmt.Println(err)
	conn, err := pool.Get()
	defer conn.Close()

	conn.Write([]byte("PING\r\n\r\n"))
	buff := make([]byte, 512)
	n, _ := conn.Read(buff)
	log.Printf("Receive: %s", buff[:n])
}
