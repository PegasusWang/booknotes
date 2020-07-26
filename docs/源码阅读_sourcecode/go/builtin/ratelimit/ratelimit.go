package main

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

/*
"golang.org/x/time/rate" 源码
https://en.wikipedia.org/wiki/Token_bucket

令牌桶算法：
Algorithm:

The token bucket algorithm can be conceptually understood as follows:

- A token is added to the bucket every 1/r seconds.
- The bucket can hold at the most b tokens. If a token arrives when the bucket is full, it is discarded.
- When a packet (network layer PDU) of n bytes arrives,
	- if at least n tokens are in the bucket, n tokens are removed from the bucket, and the packet is sent to the network.
	- if fewer than n tokens are available, no tokens are removed from the bucket, and the packet is considered to be non-conformant.

令牌桶算法 (token bucket)

- 一个 token 以 1/r 秒添加到桶
- 桶最多可以装 b 个 token。如果添加 token 的时候桶已经满了，就直接丢弃要加入的令牌。
- 当需要获取令牌的时候：
	- 如果桶里至少有还有 n 个令牌，就从桶里移除这 n 个令牌，然后放行这 n 个请求。
	- 如果桶里少于 n 个令牌，不会移除令牌。请求需要排队，或者丢弃

https://www.kancloud.cn/mutouzhang/go/596856
*/

type Limit float64

const Inf = Limit(math.MaxFloat64) // 无限制

func Every(interval time.Duration) Limit {
	if interval < 0 {
		return Inf
	}
	return 1 / Limit(interval.Seconds())
}

// 由于令牌池中最多有b个令牌，所以一次最多只能允许b个事件发生，一个事件花费掉一个令牌。
// Limter限制时间的发生频率，采用令牌桶的算法实现。这个池子一开始容量为b，装满b个令牌，然后每秒往里面填充r个令牌。
type Limiter struct {
	limit Limit // NewLimiter(limit, burst) 中的参数
	burst int   // 单词 allow 请求获取的最大 token 数 NewLimiter(r, burst)

	mu     sync.Mutex // 保证并发安全
	tokens float64    // 当前桶的 token 数目
	last   time.Time  // 最后一次 token 更新时间，用来计算每次请求的时间差，乘以每秒可以增加的 token
	// lastEvent is the latest time of a rate-limited event (past or future)
	lastEvent time.Time
}

func (lim *Limiter) Limit() Limit {
	lim.mu.Lock()
	defer lim.mu.Unlock()
	return lim.limit
}

func (lim *Limiter) Burst() int {
	return lim.burst
}

// r 补充速率， b 桶深度（支持一开始  爆发数目)。 r 是 qps， b 是允许突发流量大小
func NewLimiter(r Limit, b int) *Limiter {
	return &Limiter{limit: r, burst: b}
}

// 等价于 AllowN(time.Now(), 1)。一般是调用这个 allow 函数判断是否允许执行。
func (lim *Limiter) Allow() bool {
	return lim.AllowN(time.Now(), 1)
}

func (lim *Limiter) AllowN(now time.Time, n int) bool {
	return lim.reserveN(now, n, 0).ok
}

type Reservation struct {
	ok     bool // 是否允许
	lim    *Limiter
	tokens int

	timeToAct time.Time
	limit     Limit
}

func (r *Reservation) OK() bool {
	return r.ok
}

func (r *Reservation) Delay() time.Duration {
	return r.DelayFrom(time.Now())
}

const InfDuration = time.Duration(1<<63 - 1)

func (r *Reservation) DelayFrom(now time.Time) time.Duration {
	if !r.ok {
		return InfDuration
	}
	delay := r.timeToAct.Sub(now)
	if delay < 0 {
		return 0
	}
	return delay
}

func (r *Reservation) Cancel() {
	r.CancelAt(time.Now())
	return
}

func (r *Reservation) CancelAt(now time.Time) {
	if !r.ok {
		return
	}
	r.lim.mu.Lock()
	defer r.lim.mu.Unlock()

	if r.lim.limit == Inf || r.tokens == 0 || r.timeToAct.Before(now) {
		return
	}

	// 计算需要恢复的 tokens
	restoreTokens := float64(r.tokens) - r.limit.tokensFromDuration(r.lim.lastEvent.Sub(r.timeToAct))
	if restoreTokens <= 0 {
		return
	}
	now, _, tokens := r.lim.advance(now)
	tokens += restoreTokens
	if burst := float64(r.lim.burst); tokens > burst {
		tokens = burst
	}
	r.lim.last = now
	r.lim.tokens = tokens
	if r.timeToAct == r.lim.lastEvent {
		prevEvent := r.timeToAct.Add(r.limit.durationFromTokens(float64(-r.tokens)))
		if !prevEvent.Before(now) {
			r.lim.lastEvent = prevEvent
		}
	}
	return
}

func (lim *Limiter) Reserve() *Reservation {
	return lim.ReserveN(time.Now(), 1)
}

func (lim *Limiter) ReserveN(now time.Time, n int) *Reservation {
	r := lim.reserveN(now, n, InfDuration)
	return &r
}

func (lim *Limiter) Wait(ctx context.Context) (err error) {
	return lim.WaitN(ctx, 1)
}

// block 直到 允许 n 个事件发生
func (lim *Limiter) WaitN(ctx context.Context, n int) (err error) {
	if n > lim.burst && lim.limit != Inf {
		return fmt.Errorf("rate: Wait(n=%d) exceeds limiter's burst %d", n, lim.burst)
	}
	// check if ctx is already canceled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	now := time.Now()
	waitLimit := InfDuration
	if deadline, ok := ctx.Deadline(); ok {
		waitLimit = deadline.Sub(now)
	}

	r := lim.reserveN(now, n, waitLimit)
	if !r.ok {
		return fmt.Errorf("rate: Wait(n=%d) would exceed context deadline", n)
	}
	// wait if necessary
	delay := r.DelayFrom(now)
	if delay == 0 {
		return nil
	}
	t := time.NewTimer(delay)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		r.Cancel()
		return ctx.Err()
	}

}

// SetLimit is shorthand for SetLimitAt(time.Now(), newLimit).
func (lim *Limiter) SetLimit(newLimit Limit) {
	lim.SetLimitAt(time.Now(), newLimit)
}

func (lim *Limiter) SetLimitAt(now time.Time, newLimit Limit) {
	lim.mu.Lock()
	defer lim.mu.Unlock()

	now, _, tokens := lim.advance(now)

	lim.last = now
	lim.tokens = tokens
	lim.limit = newLimit
}

// n 是需要的 token 数量，如果不够 n 个，需要等待
func (lim *Limiter) reserveN(now time.Time, n int, maxFutureReserve time.Duration) Reservation {
	lim.mu.Lock()
	if lim.limit == Inf { // 无限大，直接放行
		lim.mu.Unlock()
		return Reservation{
			ok:        true, // ok 为 true，allowN 始终返回 True。 Allow 调用的是  (now,1,0)
			lim:       lim,
			tokens:    n,
			timeToAct: now,
		}
	}

	// lazy compute
	//这个函数返回指定到预约时间后当时Limiter会处于的状态。last表示其数据对应的最新时间；tokens表示到预约的时间点上能有多少令牌存在。
	now, last, tokens := lim.advance(now)
	tokens -= float64(n) //去掉预约的数量后，Limiter上还会剩余多少令牌。

	var waitDuration time.Duration
	if tokens < 0 { // 令牌不够用了，说明需要限流了
		waitDuration = lim.limit.durationFromTokens(-tokens) // 如果 token 不够用了，需要计算等待多久
	}

	ok := n <= lim.burst && waitDuration <= maxFutureReserve

	// 构造 Reservation
	r := Reservation{
		ok:    ok,
		lim:   lim,
		limit: lim.limit,
	}
	if ok {
		r.tokens = n
		r.timeToAct = now.Add(waitDuration)
	}

	// 更新状态
	if ok {
		lim.last = now
		lim.tokens = tokens // 扣除 n 个之后的令牌数量
		lim.lastEvent = r.timeToAct
	} else {
		lim.last = last
	}
	lim.mu.Unlock()
	return r
}

// advance calculates and returns an updated state for lim resulting from the passage of time.
// lim is not changed.
func (lim *Limiter) advance(now time.Time) (newNow time.Time, newLast time.Time, newTokens float64) {
	last := lim.last
	if now.Before(last) {
		last = now
	}
	// Avoid making delta overflow below when last is very old.
	maxElapsed := lim.limit.durationFromTokens(float64(lim.burst) - lim.tokens)
	elapsed := now.Sub(last)
	if elapsed > maxElapsed {
		elapsed = maxElapsed
	}

	delta := lim.limit.tokensFromDuration(elapsed)
	tokens := lim.tokens + delta
	if burst := float64(lim.burst); tokens > burst {
		tokens = burst // 桶的最大token 数量 burst
	}
	return now, last, tokens
}

// durationFromTokens is a unit conversion function from the number of tokens to the duration
// of time it takes to accumulate them at a rate of limit tokens per second.
func (limit Limit) durationFromTokens(tokens float64) time.Duration {
	seconds := tokens / float64(limit)
	return time.Nanosecond * time.Duration(1e9*seconds)
}

// tokensFromDuration is a unit conversion function from a time duration to the number of tokens
// which could be accumulated during that duration at a rate of limit tokens per second.
func (limit Limit) tokensFromDuration(d time.Duration) float64 {
	return d.Seconds() * float64(limit)
}

/***************************** **********************************/

// 按照令牌桶的的原理写一个简单的限流
type MyLimiter struct {
	limit float64
	burst int

	mu     sync.Mutex
	tokens float64
	last   time.Time // 上一次消耗令牌的时间
}

/*
令牌桶算法 (token bucket)

- 一个 token 以 1/r 秒添加到桶( 1s 添加 r 个 token)
- 桶最多可以装 b 个 token。如果添加 token 的时候桶已经满了，就直接丢弃要加入的令牌。
- 当需要获取令牌的时候：
	- 如果桶里至少有还有 n 个令牌，就从桶里移除这 n 个令牌，然后放行这 n 个请求。
	- 如果桶里少于 n 个令牌，不会移除令牌。请求需要排队，或者丢弃
*/

func NewMyLimiter(rate float64, burst int) *MyLimiter {
	return &MyLimiter{limit: rate, burst: burst}
}

func (lim *MyLimiter) Allow() bool {
	return lim.AllowN(time.Now(), 1)
}

// 同标准库的方式，只不过演示一下原理，写的很简单
func (lim *MyLimiter) AllowN(now time.Time, n int) bool {
	lim.mu.Lock()
	defer lim.mu.Unlock()

	delta := now.Sub(lim.last).Seconds() * lim.limit // 计算到上一次应该加多少个令牌
	lim.tokens += delta

	if lim.tokens > float64(lim.burst) { // 放入新增的令牌数目 之后超过了最大值就丢弃
		lim.tokens = float64(lim.burst) //超过的令牌丢弃
	}

	if lim.tokens < float64(n) { // 当前令牌数目不够要消耗的 n 个，返回不允许
		return false
	}

	// token 够用，减掉消耗的 n 个，然后更新时间
	lim.tokens -= float64(n)
	lim.last = now
	return true
}

type TestLimiter struct {
	Limiter // 嵌入。可以自己再封装一下，结合自己的业务与来用
}

func NewTestLimiter(r Limit, b int) *TestLimiter {
	l := Limiter{limit: r, burst: b}
	return &TestLimiter{Limiter: l}
}

func main() {
	// NewLimiter(rate, b)  每秒放行 rate 个，桶的大小是b ，允许突发流量是 b，突发之后 qps 是 rate
	// var limiter = NewLimiter(1, 5)
	// var limiter = NewMyLimiter(1, 5)
	var limiter = NewTestLimiter(1, 5)

	for {
		n := 4
		for i := 1; i < n; i++ {
			go func(i int) {
				if !limiter.Allow() {
					fmt.Printf("forbid [%d]\n", i)
				} else {
					fmt.Printf("allow  [%d]\n", i)
				}
			}(i)
		}
		time.Sleep(time.Second)
		fmt.Printf("=================\n")
	}
}
