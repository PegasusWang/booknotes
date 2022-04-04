// 断路器实现逻辑
package myhystrix

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type CircuitBreaker struct {
	Name                   string
	open                   bool
	forceOpen              bool
	mutex                  *sync.RWMutex
	openedOrLastTestedTime int64

	executorPool *executorPool
	metrics      *metricExchange
}

var (
	circuitBreakerMutex = &sync.RWMutex{}
	circuitBreakers     = make(map[string]*CircuitBreaker)
)

func GetCircuit(name string) (*CircuitBreaker, bool, error) {
	circuitBreakerMutex.RLock()
	_, ok := circuitBreakers[name]
	if !ok {
		circuitBreakerMutex.RUnlock()
		circuitBreakerMutex.Lock() // 写锁
		defer circuitBreakerMutex.Unlock()

		// double check
		if cb, ok := circuitBreakers[name]; ok {
			return cb, false, nil
		}
		circuitBreakers[name] = newCircuitBreaker(name)
	} else {
		defer circuitBreakerMutex.RUnlock()
	}
	return circuitBreakers[name], !ok, nil
}

func newCircuitBreaker(name string) *CircuitBreaker {
	c := &CircuitBreaker{}
	c.Name = name
	c.metrics = newMetricExchange(name)
	c.executorPool = newExecutorPool(name)
	c.mutex = &sync.RWMutex{}
	return c
}

// 清理所有指标和断路器
func Flush() {
	circuitBreakerMutex.Lock()
	defer circuitBreakerMutex.Unlock()
	for name, cb := range circuitBreakers {
		cb.metrics.Reset()
		cb.executorPool.Metrics.Reset()
		delete(circuitBreakers, name)
	}
}

func (cb *CircuitBreaker) toggleForceOpen(toggle bool) error {
	circuit, _, err := GetCircuit(cb.Name)
	if err != nil {
		return err
	}
	circuit.forceOpen = toggle
	return nil
}

func (cb *CircuitBreaker) IsOpen() bool {
	cb.mutex.RLock()
	o := cb.forceOpen || cb.open
	cb.mutex.RUnlock()
	if o {
		return true
	}

	if uint64(cb.metrics.Requests().Sum(time.Now())) < getSettings(cb.Name).RequestVolumeThreshold {
		return false
	}

	if !cb.metrics.IsHealthy(time.Now()) {
		// too many failures, open the cb
		cb.setOpen()
		return true
	}

	return false
}

func (cb *CircuitBreaker) setOpen() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	if cb.open {
		return
	}
	cb.openedOrLastTestedTime = time.Now().UnixNano()
	cb.open = true
}

func (cb *CircuitBreaker) AllowRequest() bool {
	return !cb.IsOpen() || cb.allowSingleTest()
}

func (cb *CircuitBreaker) allowSingleTest() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	now := time.Now().UnixNano()
	openedOrLastTestedTime := atomic.LoadInt64(&cb.openedOrLastTestedTime)
	// TODO why
	if cb.open && now > openedOrLastTestedTime+getSettings(cb.Name).SleepWindow.Nanoseconds() {
		swapped := atomic.CompareAndSwapInt64(&cb.openedOrLastTestedTime, openedOrLastTestedTime, now)
		if swapped {
			log.Printf("hystrix-go: allowing single test to possibly close circuit %v", cb.Name)
		}
		return swapped
	}
	return false
}

func (cb *CircuitBreaker) setClose() {
	cb.mutex.Lock()
	cb.mutex.Unlock()

	if !cb.open {
		return
	}
	cb.open = false
	cb.metrics.Reset()
}

// ReportEvent records command metrics for tracking recent error rates and exposing data to the dashboard.
func (cb *CircuitBreaker) ReportEvent(eventTypes []string, start time.Time, runDuration time.Duration) error {
	if len(eventTypes) == 0 {
		return fmt.Errorf("no event types sent for metrics")
	}
	cb.mutex.RLock()
	o := cb.open
	cb.mutex.RUnlock()
	if eventTypes[0] == "success" && o {
		cb.setClose()
	}

	var concurrencyInUse float64
	if cb.executorPool.Max > 0 {
		concurrencyInUse = float64(cb.executorPool.ActiveCount()) / float64(cb.executorPool.Max)
	}
	select {
	case cb.metrics.Updates <- &commandExecution{
		Types:            eventTypes,
		Start:            start,
		RunDuration:      runDuration,
		ConcurrencyInUse: concurrencyInUse,
	}:
	default:
		return CircuitError{Message: fmt.Sprintf("metrics channel (%v) is at capacity", cb.Name)}
	}
	return nil
}
