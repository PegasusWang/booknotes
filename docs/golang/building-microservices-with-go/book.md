code: https://github.com/building-microservices-with-go/

# 3. Docker

 Containers are immutable instances of images, and the data volumes are by default non-persistent.

# 4. Testing

use httptest

```
func TestSearchHandler( t *testing.T) {
	handler := SearchHandler{}
	request := httptest.NewRequest("GET", "/search", nil)
	response := httptest.NewRecorder()

	handler.ServeHTP(response, request)
	if response.Code != http.StatusBadRequest{
		t.Errorf("Expected BadRequest got %v", response.Code)
	}
}
// httptest generate Mock versions of the dependent objects http.Request and http.ResponseWriter
```

### Dependency injection and mocking

github.com/stretchr/testify

### Code coverage

```
go test -cover ./...
```

### Behavioral Driven Developent(BDD)

github.com/DATA-DOG/godog/cmd/godog

### Testing with Docker Compose


### Benchmarking and profiling

search_bench_test.go
go test -bench=. -benchmem


Go supports three different types of profiling:

- CPU, Identifies the tasks which require the most CPU time
- Heap: Identifies the statements responsible for allocating the most memory
- Blocking: Identifies the operations responsible for blocking Goroutines for the longest time

Add it to the beginning of your main Go file and, if you are not already running an HTTP web server,
start one:

```
import (
	"log"
	_ "net/http/pprof"
)

go func() {
	log.Println(http.ListenAndServe("localhost:6060", nil))
}
```

# 5. Commong Patterns

### Design for failure
Anything that can go wrong will go wrong.
想象一个场景，你使用了同步方式调用第三方邮件发送API，某一天用户量因为打了广告大涨，结果却因为调用
邮件API的频率限制导致应用一直失败。

### Patterns
##### Event Processing
The first question we should ask ourselves is "Does this call need to be synchronous?"

Event processing with at least once delivery.

##### Hanlding Errors
Append the error every time we fail to process a message as it gives us the history of what went wrong.

##### Dead Letter Queue
we can examine the failed messages on this queue to assist us with debugging the system

##### Idempotent transactions and message order

##### Atomic transactions

try to avoid distributed transactons. Use message queue, when somethiing fails, keep retrying

##### Timeouts
The key feature of a timeout is to fail fast and to notify the caller of this failure.

```
// github.com/eapache/go-resiliency/tree/master/deadline
func makeTimeoutRequest() {
	dl := deadline.New(1 * time.Second)
	err := dl.Run(func(stopper <-chan struct{}) error {
		slowFunction()
		return nil
	})
	switch err {
	case deadline.ErrTimeOut:
		fmt.Println("Timeout")
	default:
		fmt.Println(err)
	}
}
```

##### Back off

A backoff algorithm waits for a set period before retrying after the first failure, this then increments
with subsequent failures up to a maximum duration.

go-resiliency package and the retryier package

##### Circuit breaking
Circuit breaking is all about failing fast, automatically degrade functionality when the system is under stress.

how it works:
Under normal operations, like a circuit breaker in your electricity switch box, the breaker is closed and traffic
flows normally. However, once the pre-determined error threshold has been exceeded, the breaker enters the open state,
and all requests immediately fail without even being attempted. After a period, a further request would be allowed
and the circuit enters a half-open state, in this state a failure immediately returns to the open state regardless
of the errorThreshold. Once some requests have been processed without any error, then the circuit again returns
to the closed state, and only if the number of failures exceeded the error threshold would the circuit open again.

![](./circuit.png)

```go
// go-resilience

// Threshold: number of times a request can fail before the circuit opens
// successThreshold: number of times that we need a successful reqeust in the half-open state before we move back to open
// timeout: the time that circuit will stay in the open state before chaning to half-open
func New (error Threshold, successThreshold int, timeout time.Duration) *Breaker
package main

import (
	"fmt"
	"time"
)

func main() {
	b := breaker.New(3, 1, 5*time.Second)
	for {
		result := b.Run(func() error {
			// call some service
			time.Sleep(2 * time.Second)
			return fmt.Errorf("Timeout")
		})

		switch result {
		case nil:
			// success
		case breaker.ErrBreadkerOpen:
			// our function wasn't run because the breaker was open
			fmt.Println("Breaker open")
		default:
			fmt.Println(result)
		}
		time.Sleep(500 * time.Millisecond)
	}

}
```

One of more modern implementations of circuit breaking and timeouts is the Hystrix library from Netflix.

- github.com/Netflix/Hystrix
- github.com/afex/hystrix-go

##### Health checks

Every services should expose a health check endpoint which can be accessed by the consul or another server monitor.
Recommend you look at implementing these features:

- Data store connections status(general connection state, connection pool status)
- current response time (rolling average)
- burrent connections
- bad requests(running average)

##### Throttling (限流)
Throttling is a pattern where you restrict the number of connections that a service can handle,
returning an HTTP error code when this threshold has been exceeded.

### Service discovery
> Microservices are easy, building microservice is hard

The solution is service discovery and the use of a dynamic sevice registery, like Consul or Etcd.
There two main patterns for service discovery:

##### server-side service discovery
typically, there will be a resverse proxy which acts as a gateway to your services, it contains the dynamic
service registry and forwards your request on to the backend services.
The reverse proxy my become a bottleneck.

##### client-side service discovery

Prefer client-side, this gives you greater control over what happens when a failure occurs.
Client is responsible for the service discovery and load balancing.

##### Load balancing
Implement in Go

`func NewLoadBalancer(strategy Strategy, endpoints []url.URL) * loadBalancer`

### Caching

you should be talking about consistency and the tradeoffs with performance and cost.
