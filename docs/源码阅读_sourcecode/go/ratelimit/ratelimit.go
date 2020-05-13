package myratelimit

import (
	"fmt"
	"sync/atomic"
	myclock "test/source_code_read/ratelimit/internal/clock"
	"time"
	"unsafe"
)

type Limiter interface {
	Take() time.Time
}

type Clock interface {
	Now() time.Time
	Sleep(time.Duration)
}

type state struct {
	last     time.Time     //
	sleepFor time.Duration // 需要 sleep 的时间
}

type limiter struct {
	state unsafe.Pointer
	// padding 搜了下源码没用到啊？
	padding [56]byte // cache line size - state pointer size = 64 - 8; created to avoid false sharing。

	perRequest time.Duration // perRequest = 1s / rate，每个请求的时间=1s/每秒执行个数
	maxSlack   time.Duration
	clock      Clock
}

// 配置 limiter
type Option func(l *limiter)

// limit to rate RPS (request per second，每秒请求数)
func New(rate int, opts ...Option) Limiter {
	l := &limiter{
		perRequest: time.Second / time.Duration(rate),
		maxSlack:   -10 * time.Second / time.Duration(rate),
	}
	for _, opt := range opts {
		opt(l)
	}
	if l.clock == nil {
		l.clock = myclock.New()
	}
	initialState := state{
		last:     time.Time{},
		sleepFor: 0,
	}
	atomic.StorePointer(&l.state, unsafe.Pointer(&initialState))
	return l
}

// 方便 mock
func withClock(clock Clock) Option {
	return func(l *limiter) {
		l.clock = clock
	}
}

var WithoutSlack Option = withoutSlackOption

func withoutSlackOption(l *limiter) {
	l.maxSlack = 0
}

// take 阻塞，保证每次不同的 call 之间间隔是 time.Second/rate。每秒操作次数
func (t *limiter) Take() time.Time {
	newState := state{}
	taken := false
	for !taken {
		fmt.Println("++++")
		now := t.clock.Now()

		preStatePtr := atomic.LoadPointer(&t.state)
		oldState := (*state)(preStatePtr)

		newState = state{} // NOTE: 已开始写 shadow 了
		newState.last = now

		// 如果是第一个请求，放行
		if oldState.last.IsZero() {
			taken = atomic.CompareAndSwapPointer(&t.state, preStatePtr, unsafe.Pointer(&newState))
			continue
		}
		newState.sleepFor += t.perRequest - now.Sub(oldState.last)
		if newState.sleepFor < t.maxSlack { // 校正
			newState.sleepFor = t.maxSlack
		}
		if newState.sleepFor > 0 {
			newState.last = newState.last.Add(newState.sleepFor)
		}
		taken = atomic.CompareAndSwapPointer(&t.state, preStatePtr, unsafe.Pointer(&newState))
	}
	t.clock.Sleep(newState.sleepFor)
	return newState.last
}

type unlimited struct{}

func NewUnlimited() Limiter {
	return unlimited{}
}

func (unlimited) Take() time.Time {
	return time.Now()
}
