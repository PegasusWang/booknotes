package main

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"
)

/*
内置的 context 源码

参考：

https://geektutu.com/post/quick-go-context.html 先看看用法
https://www.cnblogs.com/li-peng/p/11249478.html 再看源码
*/

type Context interface {
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
	Value(key interface{}) interface{}
}

// context.Background() 底层实现
type emptyCtx int

func (*emptyCtx) Deadline() (deadline time.Time, ok bool) {
	return
}

func (*emptyCtx) Done() <-chan struct{} {
	return nil
}

func (*emptyCtx) Err() error {
	return nil
}

func (*emptyCtx) Value(key interface{}) interface{} {
	return nil
}

var (
	background = new(emptyCtx)
	todo       = new(emptyCtx)
)

func (e *emptyCtx) String() string {
	switch e {
	case background:
		return "context.Background"
	case todo:
		return "context.TODO"
	}
	return "unknown empty Context"
}

// never canceled, no values, has no deadline 。平常调用 Background  返回的就是这个
func Background() Context {
	return background
}

func TODO() Context {
	return todo
}

// 几个错误的定义
// Canceled is the error returned by Context.Err when the context is canceled.
var Canceled = errors.New("context canceled")

// DeadlineExceeded is the error returned by Context.Err when the context's
// deadline passes.
var DeadlineExceeded error = deadlineExceededError{}

type deadlineExceededError struct{}

func (deadlineExceededError) Error() string   { return "context deadline exceeded" }
func (deadlineExceededError) Timeout() bool   { return true }
func (deadlineExceededError) Temporary() bool { return true }

type CancelFunc func()

func WithCancel(parent Context) (ctx Context, cancel CancelFunc) {
	c := newCancelCtx(parent)
	propagateCancel(parent, &c)
	return &c, func() { c.cancel(true, Canceled) }
}

func newCancelCtx(parent Context) cancelCtx {
	return cancelCtx{Context: parent}
}

func propagateCancel(parent Context, child canceler) {
	if parent.Done() == nil { // 检查 parent 是否可以 cancel
		return // parent is never canceld
	}
	if p, ok := parentCancelCtx(parent); ok { // 检查 parent 是否是 cancelCtx 类型
		p.mu.Lock()
		if p.err != nil {
			child.cancel(false, p.err) // parent has already been canceld，如果 cancel 了就 cancel 掉 child
		} else {
			if p.children == nil {
				p.children = make(map[canceler]struct{})
			}
			p.children[child] = struct{}{} // 加入 child
		}
		p.mu.Unlock()
	} else { // 如果不是 cancelCtx，监控 parent 和 child 的 Done
		go func() {
			select {
			case <-parent.Done():
				child.cancel(false, parent.Err())
			case <-child.Done():
			}
		}()
	}
}

func parentCancelCtx(parent Context) (*cancelCtx, bool) {
	for {
		switch c := parent.(type) {
		case *cancelCtx:
			return c, true
		case *timerCtx:
			return &c.cancelCtx, true
		case *valueCtx:
			parent = c.Context
		default:
			return nil, false
		}
	}
}

// 从 parent 里删除 child
func removeChild(parent Context, child canceler) {
	p, ok := parentCancelCtx(parent)
	if !ok {
		return
	}
	p.mu.Lock()
	if p.children != nil {
		// children map[canceler]struct{} set to nil by first cancel call
		delete(p.children, child)
	}
	p.mu.Unlock()
}

// 可以直接取消的 context 类型
// implementations are *cancelCtx and *timerCtx.
type canceler interface {
	cancel(removeFromParent bool, err error)
	Done() <-chan struct{}
}

// 复用的 closed channel
var closedchan = make(chan struct{})

func init() {
	close(closedchan)
}

// 可以取消的 ctx，取消的时候同时会取消所有实现了 canceler 的 children。实现了 cancler
type cancelCtx struct {
	Context

	mu       sync.Mutex            // 保护以下字段
	done     chan struct{}         // 延迟创建，closed by first cancel call
	children map[canceler]struct{} // set to nil by first cancel call
	err      error                 // set to non-nil by first cancel call
}

func (c *cancelCtx) Done() <-chan struct{} {
	c.mu.Lock()
	if c.done == nil {
		c.done = make(chan struct{}) //  created lazily
	}
	d := c.done
	c.mu.Unlock()
	return d
}

func (c *cancelCtx) Err() error {
	c.mu.Lock()
	err := c.err
	c.mu.Unlock()
	return err
}

func (c *cancelCtx) String() string {
	return fmt.Sprintf("%v.WithCancel", c.Context)
}

// close c.done, cancels each of c's children, and, if removeFromParent is True, rmoves c from its parent's children
func (c *cancelCtx) cancel(removeFromParent bool, err error) {
	if err == nil {
		panic("context: internal error: missing cancel error")
	}
	c.mu.Lock()
	if c.err != nil {
		c.mu.Unlock()
		return // already canceled
	}
	c.err = err
	if c.done == nil {
		c.done = closedchan
	} else {
		close(c.done)
	}
	for child := range c.children {
		child.cancel(false, err)
	}
	c.children = nil
	c.mu.Unlock()

	if removeFromParent {
		removeChild(c.Context, c)
	}

}

func WithDeadline(parent Context, d time.Time) (Context, CancelFunc) {
	if cur, ok := parent.Deadline(); ok && cur.Before(d) {
		// parent 有 deadline 并且早于 deadline
		return WithCancel(parent)
	}
	c := &timerCtx{
		cancelCtx: newCancelCtx(parent),
		deadline:  d,
	}
	propagateCancel(parent, c)
	dur := time.Until(d)
	if dur <= 0 {
		c.cancel(true, DeadlineExceeded) // deadline has already passed
		return c, func() { c.cancel(false, Canceled) }
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.err == nil {
		c.timer = time.AfterFunc(dur, func() { // 在 dur 之后执行 func() c.cancel
			c.cancel(true, DeadlineExceeded)
		})
	}
	return c, func() { c.cancel(true, Canceled) }
}

// 带有过期时间的 timerCtx
type timerCtx struct {
	cancelCtx

	timer    *time.Timer
	deadline time.Time
}

func (c *timerCtx) Deadline() (deadline time.Time, ok bool) {
	return c.deadline, true
}

func (c *timerCtx) String() string {
	return fmt.Sprintf("%v.WithDeadline(%s [%s])", c.cancelCtx.Context, c.deadline, time.Until(c.deadline))
}

// 定时器 ctx 的超时
func (c *timerCtx) cancel(removeFromParent bool, err error) {
	c.cancelCtx.cancel(false, err)
	if removeFromParent {
		// Remove this timerCtx from its parent cancelCtx's children.
		removeChild(c.cancelCtx.Context, c)
	}
	c.mu.Lock()
	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}
	c.mu.Unlock()
}

func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc) {
	return WithDeadline(parent, time.Now().Add(timeout))
}

func WithValue(parent Context, key, val interface{}) Context {
	if key == nil {
		panic("nil key")
	}
	if !reflect.TypeOf(key).Comparable() {
		panic("key is not comparable")
	}
	return &valueCtx{parent, key, val}
}

type valueCtx struct {
	Context
	key, val interface{}
}

func (c *valueCtx) String() string {
	return fmt.Sprintf("%v.WithValue(%#v, %#v)", c.Context, c.key, c.val)
}

// 查询 key 的时候，是一个向上递归的过程
func (c *valueCtx) Value(key interface{}) interface{} {
	if c.key == key {
		return c.val
	}
	return c.Context.Value(key)
}


/* 测试代码 */

func reqTaskDeadline(ctx Context, name string) {
	for {
		select {
		case <-ctx.Done(): // 到了截止时间就退出，在 cancel 之前就会结束
			fmt.Println("stop", name, ctx.Err()) // print ctx.Err
			return
		default:
			fmt.Println(name, "send request")
			time.Sleep(1 * time.Second)
		}
	}
}

func testCtxWithDeadlineMain() {
	// 截止时间一般用 当前时间+ 超时时间
	ctx, cancel := WithDeadline(Background(), time.Now().Add(1*time.Second))
	go reqTaskDeadline(ctx, "worker1")
	go reqTaskDeadline(ctx, "worker2")

	time.Sleep(3 * time.Second)
	fmt.Println("before cancel")
	cancel()
	time.Sleep(3 * time.Second)
	fmt.Println("main done")
}

func main() {
	testCtxWithDeadlineMain()
}
