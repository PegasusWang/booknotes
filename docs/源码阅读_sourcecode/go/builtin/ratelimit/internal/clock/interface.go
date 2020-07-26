package myclock

import "time"

// 方便做 time mock
type Clock interface {
	AfterFunc(d time.Duration, f func())
	Now() time.Time
	Sleep(d time.Duration)
}
