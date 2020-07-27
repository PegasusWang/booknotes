package main

import (
	"time"

	"go.uber.org/atomic"
)

type Throttle struct {
	maxNum          atomic.Int64
	intervalSeconds int64
	throttleNum     atomic.Int64
}

func NewThrottle(maxNum, intervalSeconds int64) *Throttle {
	throttle := &Throttle{
		intervalSeconds: intervalSeconds,
	}
	throttle.maxNum.Store(maxNum)
	throttle.start()
	return throttle
}

func (t *Throttle) start() {
	go func() {
		ticker := time.NewTicker(time.Duration(t.intervalSeconds) * time.Second)
		for {
			select {
			case <-ticker.C:
				t.throttleNum.Store(0)
			}
		}
	}()
	return
}

func (t *Throttle) Throttle() bool {
	curNum := t.throttleNum.Add(1)
	if curNum > 0 && curNum <= t.maxNum.Load() {
		return true
	}
	return false
}

func (t *Throttle) SetMaxNum(maxNum int64) {
	t.maxNum.Store(maxNum)
}
