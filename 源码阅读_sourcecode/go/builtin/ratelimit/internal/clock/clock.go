package myclock

import (
	"container/heap"
	"sync"
	"time"
)

type Mock struct {
	sync.Mutex
	now    time.Time
	timers Timers
}

func NewMock() *Mock {
	return &Mock{now: time.Unix(0, 0)}
}

func (m *Mock) Add(d time.Duration) {
	m.Lock()

	end := m.now.Add(d)
	for len(m.timers) > 0 && m.now.Before(end) {
		t := heap.Pop(&m.timers).(*Timer)
		m.now = t.next
		m.Unlock()
		t.Tick()
		m.Lock()
	}
	m.Unlock()
	nap()
}

func (m *Mock) Timer(d time.Duration) *Timer {
	ch := make(chan time.Time)
	t := &Timer{
		C:    ch,
		c:    ch,
		mock: m,
		next: m.now.Add(d),
	}
	m.addTimer(t)
	return t
}

func (m *Mock) addTimer(t *Timer) {
	m.Lock()
	defer m.Unlock()
	heap.Push(&m.timers, t)
}

func (m *Mock) After(d time.Duration) <- chan time.Time {
	return m.Timer(d).C
}

func (m *Mock) AfterFunc( d time.Duration, f func() ) *Timer {
	t := m.Timer(d)
	go func() {
		<-t.c
		f()
	}()
	nap()
	return t
}

func (m *Mock) Now() time.Time {
	m.Lock()
	defer m.Unlock()
	return m.now
}

func (m *Mock) Sleep( d time.Duration ) {
	<-m.After(d)
}

// Timer represents a single event.
type Timer struct {
	C    <-chan time.Time
	c    chan time.Time
	next time.Time
	mock *Mock
}

func (t *Timer) Next() time.Time {
	return t.next
}

func (t *Timer) Tick() {
	select {
	case t.c <- t.next:
	default:
	}
	nap()
}

// Sleep momentarily so that other goroutines can process.
func nap() { time.Sleep(1 * time.Millisecond) }
