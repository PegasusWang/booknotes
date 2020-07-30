package myclock

import "time"

type clock struct{}

func New() Clock {
	return &clock{}
}

func (c *clock) After(d time.Duration) <-chan time.Time { return time.After(d) }

func (c *clock) AfterFunc(d time.Duration, f func()) {
	// TODO maybe return timer interface
	time.AfterFunc(d, f)
}

func (c *clock) Now() time.Time { return time.Now() }

func (c *clock) Sleep(d time.Duration) { time.Sleep(d) }
