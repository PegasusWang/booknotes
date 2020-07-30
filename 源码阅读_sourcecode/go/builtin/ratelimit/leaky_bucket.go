package limiter

import (
	"sync/atomic"
	"time"
	"unsafe"
)

// 参考：https://github.com/uber-go/ratelimit
type LeakyBucketLimiter interface {
	Take() time.Time
}

type state struct {
	last     time.Time
	sleepFor time.Duration
}

type leakyBucketLimiter struct {
	state      unsafe.Pointer
	perRequest time.Duration // perRequest = 1s / rate，每个请求间隔 1s/perRequest
	maxSlack   time.Duration
}

func NewLeakyBucketLimiter(rate int) LeakyBucketLimiter {
	l := &leakyBucketLimiter{
		perRequest: time.Second / time.Duration(rate),
		maxSlack:   -10 * time.Second / time.Duration(rate),
	}
	initialState := state{
		last:     time.Time{},
		sleepFor: 0,
	}
	atomic.StorePointer(&l.state, unsafe.Pointer(&initialState))
	return l
}

func (t *leakyBucketLimiter) Take() time.Time {
	newState := state{}
	taken := false
	for !taken {
		now := time.Now()

		preStatePtr := atomic.LoadPointer(&t.state)
		oldState := (*state)(preStatePtr)

		newState = state{}
		newState.last = now

		if oldState.last.IsZero() { // 首次请求
			taken = atomic.CompareAndSwapPointer(&t.state, preStatePtr, unsafe.Pointer(&newState))
			continue
		}
		newState.sleepFor += t.perRequest - now.Sub(oldState.last)
		if newState.sleepFor < t.maxSlack {
			newState.sleepFor = t.maxSlack
		}
		if newState.sleepFor > 0 {
			newState.last = newState.last.Add(newState.sleepFor)
		}
		taken = atomic.CompareAndSwapPointer(&t.state, preStatePtr, unsafe.Pointer(&newState))
	}

	time.Sleep(newState.sleepFor) // 阻塞sleep
	return newState.last
}
