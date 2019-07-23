《Concurrency In Go》

# 1 An Introduction to Concurrency


critical section（临界区）：for a section of your program needs exclusive access to a shared resource.


### Deadlocks, Livelocks, and Starvation

Deadlocks conditions:

- Mutual Exclusion
- Wait For Condition
- No Preemption
- Circular Wait

Livelocks: 活锁。想象两个人迎面走来，一个人向一边转，然后另一个人也同方向转，如此循环一直僵持谁都过不去。

Starvation: 饥饿 ，一个并发的进程无法获取所有需要工作的资源。一个贪心的进程阻止其他进程获取执行资源。

### Determining Concurrency Safety

- who is responsible for the Concurrency?
- how is the problem space mapped onto concurrency primitives?
- who is responsible for the synchronization?

```go
func CalculatePi(begin,end int64, pi *Pi)
func CalculatePi(begin,end int64) []int64
func CalculatePi(bengin,end int64) <-chan uint
```

### Simplicity in the Face of Complexity


# 2 Modeling Your Code: Communicating Sequential Processes

### Concurrency vs Parallelism

Concurrency is a property of the code; parallelism is a property of the runnning programm.

### What Is CAP?

### Go's Philosophy on Concurrency

CSP primitives or Memory access synchronizations.
Use whichever is most expressive and/or most simple.

![](./decision_tree.png)

Aim for simplicity, use channels when possible, and treat goroutines like a free resource.


# 3 Go's Concurrency Building Blocks

Coroutines are simply concurrent subroutines(functions, clousures or methods in Go) that are nonpreemptive.

Go follows a model of concurrency called the fork-join model.


![](./join_point.png)

```go
var wg sync.WaitGroup

sayHello := func() {
	defer wg.Done()
	fmt.Println("hello")
}
wg.Add(1)
go sayHello()
wg.Wait() # this is the join point
```

### The sync Package

The sync package contains the concurrency primitives that are most useful for low-level memory access synchronization.

#### WaitGroup

Wait for a set of concurrent operations to complete when you either don't care about the result of the concurrent 
operations , or you have other means of collecting their results.
```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("1st goroutine sleeping...")
		time.Sleep(1)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("2nd goroutine sleeping...")
		time.Sleep(2)
	}()

	wg.Wait()
	fmt.Println("All goroutines complete")
}
```

#### Mutex and RWMutex

```go
// Mutex demo 
package main

import (
	"fmt"
	"sync"
)

func main() {
	var count int
	var lock sync.Mutex

	incr := func() {
		lock.Lock()
		defer lock.Unlock()
		count++
		fmt.Printf("Incr: %d\n", count)
	}

	decr := func() {
		lock.Lock()
		defer lock.Unlock()
		count--
		fmt.Printf("Decr: %d\n", count)
	}

	var arithmetic sync.WaitGroup
	for i := 0; i <= 5; i++ {
		arithmetic.Add(1)
		go func() {
			defer arithmetic.Done()
			incr()
		}()
	}

	for i := 0; i <= 5; i++ {
		arithmetic.Add(1)
		go func() {
			defer arithmetic.Done()
			decr()
		}()
	}

	arithmetic.Wait()
	fmt.Println("Done")
}
```

RWMutex: 适合读多写少场景。可以获取多个读锁，除非锁用来持有写入。


#### Cond

A rendezvous point for goroutine waiting for or announcing the occurrence of an event.

```
for conditoinTrue() == false{
	time.Sleep(1*time.Millisecond) //sleep多久是个问题，太久效率低下，太快消耗 cpu
}
```
use Cond, we cloud write like this:

```
c := sync.NewCond(&sync.Mutex{})
c.L.Lock()
for conditoinTrue() == false {
	c.Wait() // blocking call, will suspend
}
c.L.UnLock()
```
考虑有一个固定长度为2的队列，10个 items 想要 push 进去。
只要有空间我们就希望入队，被尽早通知。

```
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	c := sync.NewCond(&sync.Mutex{})
	queue := make([]interface{}, 0, 10)

	removeFromQueue := func(delay time.Duration) {
		time.Sleep(delay)
		c.L.Lock() // enter the critical section
		queue = queue[1:] //simulate dequeuing an item
		fmt.Println("removed from queue")
		c.L.Unlock() // exit critical section
		c.Signal() // let a goroutine waiting on the condition know that something has occured
	}

	for i := 0; i < 10; i++ {
		c.L.Lock() //进入临界区, critical section
		for len(queue) == 2 {
			c.Wait() //will suspend the main goroutine until a signal on the condition has been sent
		}
		fmt.Println("Adding to queue")
		queue = append(queue, struct{}{})
		go removeFromQueue(1 * time.Second)
		c.L.Unlock() //exit critical section
	}
}
```
#### Once

only one call to "Do", even on different goroutines.

```
package main

import (
	"fmt"
	"sync"
)

func main() {
	var count int

	increment := func() {
		count++
	}

	var once sync.Once
	var wg sync.WaitGroup
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			once.Do(increment)
		}()
	}
	wg.Wait()
	fmt.Printf("count is %d\n", count) //1
}
```
NOTES: sync.Once only counts the number of times Do is Called, not 
how many times unique functions passed into Do are called.

```
package main

import (
	"fmt"
	"sync"
)

func main() {
	var count int
	increment := func() { count++ }
	decremet := func() { count-- }

	var once sync.Once
	once.Do(increment)
	once.Do(decremet)
	fmt.Printf("count is %d\n", count) //1
}
```

#### Pool

Pool is a concurrent-safe implementation of the object pool pattern.

```
myPool := &sync.Pool{
	New: func() interface{}{
		fmt.Println("Creating new instance.")
		return struct{}{}
	},
}

myPool.Get()
instance:=myPool.Get()
myPool.Put(instance)
myPool.Get()
```

Another common situation where a Pool is useful is for warming a cache of pre-allocated
objects for operations that must run as quickly as possible.

When working with sync.Pool:

- give it a New member variable that is thread-safe when called
- when receive an instance from Get,make no assumptions regarding the state of the object you receive back
- Make sure to call Put when you're finished with the object you pulled out of the pool.(with defer)
- Objects in the pool must be roughly uniform in makeup


### Channels

Go weill implicitly convert bidirectional channels to unidirectional channels when needed.

```
var receiveChan <- chan interface{}
var sendChan chan <- interface{}
dataStream := make(chan interface{})
//valid statements
receiveChan = dataStream
sendChan = dataStream
```

Unbufferd channel in Go are said to be blocking.

```
val, ok := <-stringStream
```

The second return value is a way for a read operation to indicate whether the read off the channel
was a value generated by a write elsewhere in the process, or a default value generated from a closed channel.

We can use "for range" iter channel, it will automatically break the loop when cahnnel is closed.

Buffred channels, even if no reads are performed on the channel, a goroutine can still perform n writes.
You can treat buffered channels are an inmemory FIFO queue for concurrent processes to communicate over.
