# 21. Channels - The Go Way to do concurrency synchronizations

Don't (let computations) communicate by sharing memroy, (let them) share memory by communicating (through channels).

五种操作(all these operations are already synchronized)：

- close(ch), ch must not be a receive-only channel.
- `ch <- v`, send a value
- `<-ch`, receive a value from the channel
- cap(ch), value buffer capacity, return int
- len(ch), query current number of values in the value buffer

We can think of each channel as maintaining 3 queues:
- the receiving goroutine queue
- the sending goroutine queue
- the value buffer queue

![](./channel.png)

```go
// unbufferd demo
package main

import (
	"fmt"
	"time"
)

func main() {
	c := make(chan int)
	go func(ch chan<- int, x int) {
		time.Sleep(time.Second)
		ch <- x * x //block until the result is received.
	}(c, 3)

	done := make(chan struct{})

	go func(ch <-chan int) {
		n := <-ch //block until 9 is sent
		fmt.Println(n)

		time.Sleep(time.Second)
		done <- struct{}{}
	}(c)
	<-done //block until a value is sent to done
	fmt.Println("bye")
}

// buffered channel
package main

import (
	"fmt"
	"time"
)

// A never ending football game
func main() {
	var ball = make(chan string)
	kickBall := func(playerName string) {
		for {
			fmt.Println(<-ball, "kicked the ball.")
			time.Sleep(time.Second)
			ball <- playerName
		}
	}
	go kickBall("John")
	go kickBall("Alice")
	go kickBall("Bob")
	go kickBall("Emily")
	ball <- "referee" //kick off 开球
	var c chan bool   //nil
	<-c               // blocking here forever
}

```

- Channel Element Values are Transferred by Copy。If the passed value size too large, use a pointer element type instead.
- A goroutine can be garbage collected when it has already exited.
- Channel send and receive operatoins are simple statements.

empty select-case code block `select{}` will make current goroutine stay in blocking state forever.

# 22 Methods in Go

Should a method be declared with pointer receiver or value receiver ?

- Too many pointer copies my cause heavier workload for garbage collector
- if value receiver type is Large , should use pointer receiver.
- declaring methods of both value receivers ans pointer receivers for the same base type is more
	likely to cause data races if the declared methods are called concurrently in multiple goroutines.
- values of the types in sync standard package should not be copied.

If it is hard to make a decisoin , just choose the pointer receiver way.

# 23 Interfaces in Go

type assert and type switch

- Values of []t can't be directly converted to []I, even if type T implements interface type I.
- Each method specified in a interface type corresponds to an implicit function.


# 37 Channel Use Cases

## Use channels as Futures/Promises

#### Return receive-only channels as results

```go
package main

import (
	"fmt"
	"math/rand"
	"time"
)

func longTimeRequest() <-chan int32 {
	r := make(chan int32)
	go func() {
		// simulate a workload
		time.Sleep(time.Second * 3)
		r <- rand.Int31n(100)
	}()
	return r
}

func sumSquares(a, b int32) int32 {
	return a*a + b*b
}

func main() {
	rand.Seed(time.Now().UnixNano())
	a, b := longTimeRequest(), longTimeRequest()
	fmt.Println(sumSquares(<-a, <-b)) // 3s 返回, a,b 并发执行的
}
```

#### Pass send-only channels as arguments

```go
package main

import (
	"fmt"
	"math/rand"
	"time"
)

func longTiemRequest(r chan<- int32) {
	// simulate a worklaod
	time.Sleep(time.Second * 3)
	r <- rand.Int31n(100)
}

func sumSquares(a, b int32) int32 {
	return a*a + b*b
}

func main() {
	rand.Seed(time.Now().UnixNano())
	ra, rb := make(chan int32), make(chan int32)
	go longTiemRequest(ra)
	go longTiemRequest(rb)
	fmt.Println(sumSquares(<-ra, <-rb))
}
```

#### The first response wins

```go
package main

import (
	"fmt"
	"math/rand"
	"time"
)

func source(c chan<- int32) {
	ra, rb := rand.Int31(), rand.Intn(3)+1
	time.Sleep(time.Duration(rb) * time.Second)
	c <- ra
}

func main() {
	rand.Seed(time.Now().UnixNano())
	beg := time.Now()
	c := make(chan int32, 5) // must bufferd channel

	for i := 0; i < cap(c); i++ {
		go source(c)
	}
	// only frist resposne will be used
	rnd := <-c
	fmt.Println(time.Since(beg))
	fmt.Println(rnd)
}
```

## Use Channels for Notifications

Use blank struct{} as element types of the notification channels, size of type struct{} is zero, doesn't consume memory.

#### 1-To-1 notification by sending a value to a channel


```go
package main

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
)

func main() {

	values := make([]byte, 32*1024*1024)
	if _, err := rand.Read(values); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	done := make(chan struct{})
	go func() {
		sort.Slice(values, func(i, j int) bool {
			return values[i] < values[j]
		})
		// notify sorting is done
		done <- struct{}{}
	}()

	// do some other things
	fmt.Println("other thing")
	fmt.Println(values[0], values[len(values)-1])
}
```

#### 1-To-1 notification by receiving a value from a channel

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	done := make(chan struct{})
	go func() {
		fmt.Print("hello")
		time.Sleep(time.Second * 2)
		<-done
	}()
	// blocked here, wait for a notification
	done <- struct{}{}
	fmt.Println(" world")
}
```

slowers notify the faster waiting for notifications.

#### N-to-1 and 1-to-N notifications


```go
package main

import (
	"log"
	"time"
)

// T type
type T = struct{}

func worker(id int, ready <-chan T, done chan<- T) {
	<-ready // block here and wait a notification
	log.Print("Worker#", id, " starts.")
	// simulate a workload
	time.Sleep(time.Second * time.Duration(id+1))
	log.Print("Worker#", id, " job done.")
	// notify main goroutine (n-to-1)
	done <- T{}
}

func main() {
	log.SetFlags(0)
	ready, done := make(chan T), make(chan T)
	go worker(0, ready, done)
	go worker(1, ready, done)
	go worker(2, ready, done)

	// simulate an initialization phase
	time.Sleep(time.Second * 3 / 2)
	// 1-to-n notifications
	ready <- T{}
	ready <- T{}
	ready <- T{}
	// Being N-to-1 notified
	<-done
	<-done
	<-done
}
```

更常见的使用 sync.WaitGroup 做 N-to-1，通过 close channels 实现 1-to-N.

#### Broadcast (1-To-N) notifications by closing a channel

上例中的三个发送 ready 可以直接换成一个 close(ready)

#### Timer: scheduled notification
