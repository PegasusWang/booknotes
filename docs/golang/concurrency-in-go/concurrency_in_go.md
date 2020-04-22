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
func CalculatePi(begin,end int64) <-chan uint
```

### Simplicity in the Face of Complexity


# 2 Modeling Your Code: Communicating Sequential Processes

### Concurrency vs Parallelism

Concurrency is a property of the code; parallelism is a property of the runnning programm.

### What Is CSP?

### Go's Philosophy on Concurrency

CSP primitives or Memory access synchronizations.
Use whichever is most expressive and/or most simple.

![](./decision_tree.png)

Aim for simplicity, use channels when possible, and treat goroutines like a free resource.


# 3 Go's Concurrency Building Blocks

Coroutines are simply concurrent subroutines(functions, clousures or methods in Go) that are nonpreemptive.(非抢占式)

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
wg.Wait() # this is the join point，等 child 执行完毕
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

	wg.Add(1) // 注意 Add 必须在外边而不是在 go func() 函数体里边
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

#### Pool(并发安全对象池)

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

Go will implicitly convert bidirectional channels to unidirectional channels when needed.

```
var receiveChan <- chan interface{}
var sendChan chan <- interface{}
dataStream := make(chan interface{})
//valid statements, 双向的可以赋值给单向的
receiveChan = dataStream
sendChan = dataStream
```

Unbufferd channel in Go are said to be blocking.

```
val, ok := <-stringStream // ok 判断还有没有值，从一个已经 closed 的 channel 依然会获取默认值
```

The second return value is a way for a read operation to indicate whether the read off the channel
was a value generated by a write elsewhere in the process, or a default value generated from a closed channel.

We can use "for range" iter channel, it will automatically break the loop when cahnnel is closed.

Buffred channels, even if no reads are performed on the channel, a goroutine can still perform n writes.
You can treat buffered channels are an inmemory FIFO queue for concurrent processes to communicate over.

![](./channel1.png)
![](./channel2.png)

The first thing we should do to put channels in the right context is to assign channel ownership.

- Instaniate the channel.
- Perform writes, os pass ownership to another goroutine.
- Close the channel
- Ecapsulate the previous three things in this list and expose them via a reader channel.

As a consumer of a channel, I only have to worry about two things.

- Knowing when a channel is closed (use val, ok)
- Responsibly handling blocking for any reason

```
// keep the scope of channel ownership small
func main() {

	chanOwner := func() <-chan int {
		resultStream := make(chan int, 5)
		go func() {
			defer close(resultStream)
			for i := 0; i <= 5; i++ {
				resultStream <- i
			}
		}()
		return resultStream // will implicityly converted to read-only for consumers
	}

	resultStream := chanOwner()
	for result := range resultStream {
		fmt.Printf("received : %d\n", result)
	}

	fmt.Println("Done receiving!")
}
```

### The select Statement

The select statement is the glue that binds channels together,
it can help safely bring channels together with concepts like cancellations, timeouts, waiting, and default values.


```
func main() {
	// a bit like switch
	var c1, c2 <-chan interface{}   // receive only chan
	var c3 chan<- interface{}  // send only chan
	// all channels reads and writes are considered simultaneously to see if any of them are ready
	// populated or closed channels in the case of reads, and channels that are not at capacity in the case of writes
	// if none of the channels are ready, the entire select statement blocks
	select {
	case <-c1:
		// do something
	case <-c2:
		// do something
	case c3 <- struct{}{}:
		// do something
	}
}
```

- What happens when multiple channels have something to read?

  - random select, each has an equal  chance of being selected as all the others


```
func main() {
	c1 := make(chan interface{})
	close(c1)
	c2 := make(chan interface{})
	close(c2)
	var c1Count, c2Count int
	for i := 1000; i >= 0; i-- {
		select {
		case <-c1:
			c1Count++
		case <-c2:
			c2Count++
		}
	}
	fmt.Printf("c1: %d\n, c2: %d\n", c1Count, c2Count)
}
```

- What if there are never any channels that become ready?

  - you may want to timeout

```
func main() {
	var c <-chan int
	select {
	case <-c:
	case <-time.After(1 * time.Second):
		fmt.Println("Time out")
	}
}
```

- what if we want to do something but no channels are currently ready?

  - use default
  - this allows a goroutine to make progress on work while waiting for another goroutine to report a result

```
func main() {
	start := time.Now()
	var c1, c2 <-chan int
	select {
	case <-c1:
	case <-c2:
	default:
		fmt.Printf("In default after %v\n\n", time.Since(start))
	}
}
```


```
func main() {
	done := make(chan interface{})
	go func() {
		time.Sleep(5 * time.Second)
		close(done)
	}()
	workCounter := 0
loop:
	for {
		select {
		case <-done:
			break loop
		default:
		}
		// simulate work
		workCounter++
		time.Sleep(1 * time.Second)
	}
	fmt.Printf("Achieved %v cycles of work before signalled to stop.\n", workCounter)
}
```
###### empty select
`seelct {}` will block forever.


### The GOMAXPROCS lever

this function controls the number of OS threads that will host so-called "work queues"


# 4. Concurrency Patterns in Go

### Confinement(限制)

When working with concurrent code

- Synchronization primitives for sharing memory (sync.Mutex)
- Synchronization via Communicating (channels)
- Immutable Data (copy of values)
- Data protected by confinement (ad hoc and lexical)


```
func main() {
	// confines the write aspect of this channel to prevent other goroutines from writing to it
	// 限制向 channel 写数据到一个 func 里
	chanOnwer := func() <-chan int {
		results := make(chan int, 5)
		go func() {
			defer close(results)
			for i := 0; i <= 5; i++ {
				results <- i
			}
		}()
		return results
	}

	consumer := func(results <-chan int) {
		for result := range results {
			fmt.Printf("Received : %d\n", result)
		}
		fmt.Println("Done receiving!")
	}

	results := chanOnwer()
	consumer(results)
}
```
### The for-select Loop

- Sending iteration variables out on a channel

```
	for _, s := range []string{"a", "b", "c"} {
		select {
		case <-done:
			return
		case stringStream <- s:
		}
	}
```

- Looping infinitely waiting to be stopped

```
	for {
		select {
		case <-done:
			return
		default:
		}
		// do non-preemptable work
	}
```

	for {
		select {
		case <-done:
			return
		default:
		}
		// do non-preemptable work
	}


### Preventing Goroutine Leaks(goroutine 泄露)

**Goroutine are not garbage collected by the runtime.**
The goroutine has a few paths to terminiation:

- when it has completed its work
- when it cannnot continue its work due to an unrecoverable error
- when it's told to stop working

A simple example of a goroutine leak:

```
func main() {
	doWork := func(strings <-chan string) <-chan interface{} {
		completed := make(chan interface{})
		go func() {
			defer fmt.Println("doWork exited.")
			defer close(completed)
			for s := range strings {
				// do something interesting
				fmt.Println(s)
			}
		}()
		return completed
	}
	//pass nil, the strings channel will never actually gets any strings written onto it
	//and the goroutine doWork will remain in memory for the lifetime of this process
	doWork(nil)
	// perhaps more work is done here
	fmt.Println("Done.")
}
```
Use signal done. The parent goroutine passed this channel to child goroutine,
and then closes the channel when it wants to cancel the child goroutine.

```
package main

import (
	"fmt"
	"time"
)

func main() {

	doWork := func(
		done <-chan interface{},
		strings <-chan string,
	) <-chan interface{} {
		terminated := make(chan interface{})
		go func() {
			defer fmt.Println("doWork exited")
			defer close(terminated)
			for {
				select {
				case s := <-strings:
					//do somthing
					fmt.Println(s)
				case <-done:
					return
				}
			}
		}()

		done := make(chan interface{})
		terminated := doWork(done, nil)

		go func() {
			// cancel operation after 1 second
			time.Sleep(1 * time.Second)
			fmt.Println("canceling doWork goroutine...")
			close(done)
		}()

		<-terminated // joined the goroutine
		fmt.Println("Done")
	}
}
```

What if a goroutine blocked on attempting wo write a value to a channel?

```
func main() {
	newRandStream := func() <-chan int {
		randStream := make(chan int)
		go func() {
			defer fmt.Println("newRandStream closure exited.") //never run
			defer close(randStream)
			for {
				randStream <- rand.Int()
			}
		}()
		return randStream
	}

	randStream := newRandStream()
	fmt.Println("3 random ints:")
	for i := 1; i <= 3; i++ {
		fmt.Printf("%d:%d\n", i, <-randStream)
	}
	// 依然退出，输出1到3
}
```

We can also use a done channel:

```
func main() {
	newRandStream := func(done <-chan interface{}) <-chan int {
		randStream := make(chan int)
		go func() {
			defer fmt.Println("newRandStream closure exited.") //never run
			defer close(randStream)
			for {
				select {
				case randStream <- rand.Int():
				case <-done:
					return
				}
			}
		}()
		return randStream
	}

	done :=make(chan interface{})
	randStream := newRandStream()
	fmt.Println("3 random ints:")
	for i := 1; i <= 3; i++ {
		fmt.Printf("%d:%d\n", i, <-randStream)
	}
	close(done)
}
```

NOTE: If a goroutine is responsible for creating a goroutine, it is also responsible for ensuring
it can stop the gorutine.

### The or-channel pattern

At times you may find yourself wanting to combine one or more done channels into a single done channel
that closes if any of its component channels close.
This pattern creates a composite done channel through recursion and goroutines.

```
func main() {
	var or func(channels ...<-chan interface{}) <-chan interface{}
	or = func(channels ...<-chan interface{}) <-chan interface{} {
		switch len(channels) {
		case 0:
			return nil
		case 1:
			return channels[0]
		}

		orDone := make(chan interface{})
		go func() {
			defer close(orDone)
			switch len(channels) {
			case 2:
				select {
				case <-channels[0]:
				case <-channels[1]:
				}
			default:
				select {
				case <-channels[0]:
				case <-channels[1]:
				case <-channels[2]:
				case <-or(append(channels[3:], orDone)...):
				}
			}
		}()
		return orDone
	}
}

func test() {
	sig := func(after time.Duration) <-chan interface{} {
		c := make(chan interface{})
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}
	start := time.Now()
	<-or(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(1*time.Second),
	)
	fmt.Printf("done after %v\n", time.Since(start))
}
```
### Error Handling

```
package main

import (
	"fmt"
	"net/http"
)

// Result is http result
type Result struct {
	Error    error
	Response *http.Response
}

func main() {
	checkStatus := func(done <-chan interface{}, urls ...string) <-chan Result {
		results := make(chan Result)
		go func() {
			defer close(results)
			for _, url := range urls {
				var result Result
				resp, err := http.Get(url)
				result = Result{Error: err, Response: resp}
				select {
				case <-done:
					return
				case results <- result:
				}
			}
		}()
		return results
	}

	done := make(chan interface{})
	defer close(done)
	urls := []string{"https://www.baidu.com", "https://badhost"}
	for result := range checkStatus(done, urls...) {
		if result.Error != nil {
			fmt.Printf("error:%v", result.Error)
			continue
		}
		fmt.Printf("Response: %\n", result.Response.Status)
	}
}
```

Errors should be considered first-class citizens when constructing values to return from goroutines.
If your goroutine can produce errors, those errors should be tightly coupled with your result type,
and passed along through the same lines of communication.

### Pipelines

It is a very powerful tool to use when your program needs to process streams, or batched of data.

Stage: taskes data in , performs a transformation on it, and sends the data back out


```
	multiply := func(values []int, multiplier int) []int {
		multipliedValues := make([]int, len(values))
		for i, v := range values {
			multipliedValues[i] = v * multiplier
		}
		return multipliedValues
	}
```

#### Best practice for constructing Pipelines

```
package main

import "fmt"

func main() {
	generator := func(done <-chan interface{}, integers ...int) <-chan int {
		intStream := make(chan int)
		go func() {
			defer close(intStream)
			for _, i := range integers {
				select {
				case <-done:
					return
				case intStream <- i:
				}
			}
		}()
		return intStream
	}

	multiply := func(done <-chan interface{}, intStream <-chan int, multiplier int) <-chan int {
		multipliedStream := make(chan int)
		go func() {
			defer close(multipliedStream)
			for i := range intStream {
				select {
				case <-done:
					return
				case multipliedStream <- i * multiplier:
				}
			}
		}()
		return multipliedStream
	}

	add := func(done <-chan interface{}, intStream <-chan int, additive int) <-chan int {
		addedStream := make(chan int)
		go func() {
			defer close(addedStream)
			for i := range intStream {
				select {
				case <-done:
					return
				case addedStream <- i + additive:
				}
			}
		}()
		return addedStream
	}

	done := make(chan interface{})
	defer close(done)

	intStream := generator(done, 1, 2, 3, 4)
	pipeline := multiply(done, add(done, multiply(done, intStream, 2), 1), 2)
	for v := range pipeline {
		fmt.Println(v)
	}

}
```
#### Some Handy Generators


```
package main

import "fmt"

func main() {

	repeat := func(done <-chan interface{}, values ...interface{}) <-chan interface{} {
		valueStream := make(chan interface{})
		go func() {
			defer close(valueStream)
			for {
				for _, v := range values {
					select {
					case <-done:
						return
					case valueStream <- v:
					}
				}
			}
		}()
		return valueStream
	}

	take := func(done <-chan interface{}, valueStream <-chan interface{}, num int) <-chan interface{} {
		takeStream := make(chan interface{})
		go func() {
			defer close(takeStream)
			for i := 0; i < num; i++ {
				select {
				case <-done:
					return
				case takeStream <- <-valueStream:
				}
			}

		}()
		return takeStream
	}

	done := make(chan interface{})
	defer close(done)

	for num := range take(done, repeat(done, 1), 10) {
		fmt.Printf("%v ", num)
	}
	// 1 1 1 1 1 1 1 1 1 1

}
```


### Fan-Out, Fan-In (扇出，扇入)


Fan-out is a term to describe the process of starting multiple gortouine to handle input from the pipeline,
and Fan-in is a term to describe the process of combining results into one channel.

Fanout one of your stages if both of the following apply:

- It doesn't rely on values that the stage had calulated before.
- It takes a long time to run


### The or-done-channel

```
	orDone := func(done, c <-chan interface{}) <-chan interface{} {
		valStream := make(chan interface{})
		go func() {
			defer close(valStream)
			for {
				select {
				case <-done:
					return
				case v, ok := <-c:
					if ok == false {
						return
					}
					select {
					case valStream <- v:
					case <-done:
					}
				}
			}
		}()
		return valStream
	}

	for val := range orDone(done, myChan) {
		// Do something with val
	}
```

### The tee-channel

Sometimes you may want to split values coming in from a channel so that you can send them off
into two separate areas of your codebase.

You can pass it a channel to read from , and it will return two seperate channels that will get the same value:

```
package main

import "fmt"

func main() {
	tee := func(done <-chan interface{}, in <-chan interface{}) (<-chan interface{}, <-chan interface{}) {
		out1 := make(chan interface{})
		out2 := make(chan interface{})
		go func() {
			defer close(out1)
			defer close(out2)
			for val := range orDone(done, in) {
				var out1, out2 = out1, out2 //shadow the variables
				for i := 0; i < 2; i++ {
					select {
					case <-done:
					case out1 <- val:
						// set its shadowed copy to nil,so further writes will block and the other channel may continue
						out1 = nil
					case out2 <- val:
						out2 = nil
					}
				}
			}
		}()
		return out1, out2
	}

	done := make(chan interface{})
	defer close(done)
	out1, out2 := tee(done, take(done, repeat(done, 1, 2), 4))

	for val1 := range out1 {
		fmt.Printf("out1: %v, out2:%v\n", val1, <-out2)
	}
}
```

### The bridge-channel

define a function that can destructure the channel of channels into s simple channel(bridging the channels)


```
package main

import "fmt"

func main() {
	bridge := func(done <-chan interface{}, chanStream <-chan <-chan interface{}) <-chan interface{} {
		valStream := make(chan interface{})
		go func() {
			defer close(valStream)
			for {
				var stream <-chan interface{} //this is the channel that will return all values from bridge
				select {
				case maybeStream, ok := <-chanStream:
					if ok == false {
						return
					}
					stream = maybeStream
				case <-done:
					return
				}
				for val := range orDone(done, stream) {
					select {
					case valStream <- val:
					case <-done:
					}
				}

			}
		}()
		return valStream
	}

	genVals := func() <-chan <-chan interface{} {
		chanStream := make(chan (<-chan interface{}))
		go func() {
			defer close(chanStream)
			for i := 0; i < 10; i++ {
				stream := make(chan interface{}, 1)
				stream <- i
				close(stream)
				chanStream <- stream
			}
		}()
		return chanStream
	}

	for v := range bridge(nil, genVals()) {
		fmt.Printf("%v ", v) //0 1 2 3 4 5 6 7 8 9
	}
}
```


### Queuing

queuing should be implemented either:

- At the entrance to your pipeline
- In stages where batching will lead to higher efficiency

### The context Package

context package serves two primary purposes:

- to provide an API for canceling branches of you call-graph
- to provide a data-bag for transporting request-scoped data through your call-graph


# 5. Concurrency at Scale

### Error Progagation

critical information:

- What happend
- When and where it occured(stack track, UTC time, machine)
- A friendly usre-facing message
- How the user can get more information(track id)

### Timeouts and Cancellation

what the reasons we might want our concurrent processes to support timeouts?

- System saturation(饱和)
- Stale data
- Attempting to prevent deadlocks

There are a number of reasons why a concurrent process might be canceld:

- Timeouts, a timeout is an implicit cancellation
- User intervention(干预)
- Parent cancellation
- Replicated requests


### Heartbeats

Heartbeats are a way for concurrent processes to signal life to outside parties. Two diferent types:

- Heartbeats that occur on a time interval.
- Heartbeats that occur at the beginning of a unit of work.


```
package main

import (
	"fmt"
	"time"
)

func main() {
	doWork := func(done <-chan interface{}, pulseInterval time.Duration) (chan interface{}, <-chan time.Time) {
		heartbeat := make(chan interface{})
		results := make(chan time.Time)
		go func() {
			defer close(heartbeat)
			defer close(results)

			pulse := time.Tick(pulseInterval)
			workGen := time.Tick(2 * pulseInterval) //another ticker to simulate work coming in

			sendPulse := func() {
				select {
				case heartbeat <- struct{}{}:
				default: // no one maybe listening to our hearteat
				}
			}
			sendResult := func(r time.Time) {
				for {
					select {
					case <-done:
						return
					case <-pulse:
						sendPulse()
					case results <- r:
						return
					}
				}
			}

			for {
				select {
				case <-done:
					return
				case <-pulse:
					sendPulse()
				case r := <-workGen:
					sendResult(r)
				}
			}
		}()
		return heartbeat, results
	}

	done := make(chan interface{})
	time.AfterFunc(10*time.Second, func() { close(done) })

	const timeout = 2 * time.Second
	heartbeat, results := doWork(done, timeout/2)
	for {
		select {
		case _, ok := <-heartbeat:
			if ok == false {
				return
			}
			fmt.Println("pulse")
		case r, ok := <-results:
			if ok == false {
				return
			}
			fmt.Printf("results %v\n", r.Second())
		case <-time.After(timeout):
			return
		}
	}

}
```

For any long-running goroutines, or goroutines that need to be tested, highly recommend this pattern.

### Replicated Request

You can replicate the request to mulitple handlers(whether those be gorutines, processes, or servers),
and one of them will run faster than other ones, you can then immediately return the result.
The downside is that you'll have to utilize resources to keep multiple copies of the handlers running.


```
package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	doWork := func(done <-chan interface{}, id int, wg *sync.WaitGroup, result chan<- int) {
		started := time.Now()
		defer wg.Done()

		//Simulate random load
		simulateLoadTime := time.Duration(1+rand.Intn(5)) * time.Second
		select {
		case <-done:
		case <-time.After(simulateLoadTime):
		}
		select {
		case <-done:
		case result <- id:
		}

		took := time.Since(started)
		if took < simulateLoadTime {
			took = simulateLoadTime
		}
		fmt.Printf("%v took %v\n", id, took)
	}

	done := make(chan interface{})
	result := make(chan int)
	var wg sync.WaitGroup
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go doWork(done, i, &wg, result) //start 10 handlers to handle our request
	}
	firstReturned := <-result // grabs the first return value from the group of handlers
	close(done)               //cancel all the remaining handlers
	wg.Wait()
	fmt.Printf("Received an answer from #%v\n", firstReturned)
}
```

### Rate Limiting

Most rate limitings is done by utilizing an algorithm called the "token bucket".

```
package main

import (
	"context"
	"log"
	"os"
	"sync"

	"golang.org/x/time/rate" // go get github.com/golang/time/rate
)

func Open() *APIConnection {
	return &APIConnection{
		rateLimiter: rate.NewLimiter(rate.Limit(1), 1), //1 event per second
	}
}

type APIConnection struct {
	rateLimiter *rate.Limiter // https://godoc.org/golang.org/x/time/rate
}

func (a *APIConnection) ReadFile(ctx context.Context) error {
	if err := a.rateLimiter.Wait(ctx); err != nil {
		return err
	}
	return nil
}
func (a *APIConnection) ResoveAddress(ctx context.Context) error {
	if err := a.rateLimiter.Wait(ctx); err != nil {
		return err
	}
	return nil
}

func main() {
	defer log.Printf("Done")
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)
	apiConnection := Open()
	var wg sync.WaitGroup
	wg.Add(20)

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			err := apiConnection.ReadFile(context.Background())
			if err != nil {
				log.Printf("cannot ReadFile :%v", err)
			}
			log.Printf("ReadFile")
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			err := apiConnection.ResoveAddress(context.Background())
			if err != nil {
				log.Printf("cannot ResolveAddress :%v", err)
			}
			log.Printf("ResolveAddress")
		}()
	}
	wg.Wait()

}
```

It's eaiser to keep the limiters separate and then combine them into one rate limiter that manages the interaction for
you.


```
package main

import (
	"context"
	"sort"
	"time"

	"golang.org/x/time/rate" // go get github.com/golang/time/rate
)

type RateLimiter interface {
	Wait(context.Context) error
	Limit() rate.Limit
}

func MultiLimiter(limiters ...RateLimiter) *multiLimiter {
	byLimit := func(i, j int) bool {
		return limiters[i].Limit() < limiters[j].Limit()
	}
	sort.Slice(limiters, byLimit)
	return &multiLimiter{limiters: limiters}
}

type multiLimiter struct {
	limiters []RateLimiter
}

func (l *multiLimiter) Wait(ctx context.Context) error {
	for _, l := range l.limiters {
		if err := l.Wait(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (l *multiLimiter) Limit() rate.Limit {
	// return the most restrictive limit
	return l.limiters[0].Limit()
}

func Open() *APIConnnection {
	secondLimit := rate.NewLimiter(Per(2, time.Second), 1)
	minuteLimit := rate.NewLimiter(Per(10, time.Minute), 10)
	return &APIConnnection{
		rateLimiter: MultiLimiter(secondLimit, minuteLimit),
	}
}

type APIConnnection struct {
	rateLimiter RateLimiter
}

func (a *APIConnnection) ReadFile(ctx context.Context) error {
	if err := a.rateLimiter.Wait(ctx); err != nil {
		return err
	}
	return nil
}

func (a *APIConnnection) ResolveAddress(ctx context.Context) errro {
	if err := a.rateLimiter.Wait(ctx); err != nil {
		return err
	}
	return nil
}
```

### Healing Unhealthy Goroutines

Ensure your long-lived goroutines stay up and healthy.

```
package main

import (
	"log"
	"time"
)

type startGoroutineFn func(done <-chan interface{}, pulseInterval time.Duration) (heartbeat <-chan interface{})

func main() {

	newSteward := func(timeout time.Duration, startGoroutine startGoroutineFn) startGoroutineFn {
		return func(done <-chan interface{}, pulseInterval time.Duration) <-chan interface{} {
			heartbeat := make(chan interface{})
			go func() {
				defer close(heartbeat)

				var wardDone chan interface{}
				var wardHeartbeat <-chan interface{}
				startWard := func() {
					wardDone = make(chan interface{})
					wardHeartbeat = startGoroutine(or(wardDone, done), timeout/2)
				}
				startWard()
				pulse := time.Tick(pulseInterval)

			monitorLoop:
				for {
					timeoutSignal := time.After(timeout)
					for {
						select {
						case <-pulse:
							select {
							case heartbeat <- struct{}{}:
							default:
							}
						case <-wardHeartbeat:
							continue monitorLoop
						case <-timeoutSignal:
							log.Println("steward: ward unhealthy: restarting")
							close(wardDone)
							startWard()
							continue monitorLoop
						case <-done:
							return
						}
					}

				}
			}()
			return heartbeat
		}
	}
}
```


# 6 Goroutines and the Go Runtime


### Work Stealing
work stealing strategy. Go follows a fork-join model for concurrency.
Forks are when goroutines are started, and join points are when two or more goroutines are
synchronized through channels or types in the sync package.

The work stealing algorithm follows a few basic rules, Given a thread of execution:

1. At A fork point, add tasks to the tail of the deque associated with the thread
2. If the thread is idle, steal work from the head of deque associated with some other random thread
3. At a join point that cannot be realized yet(i.e.,the goroutine it is synchronized with has not completed yet), pop
	 work off the tail of the thread's own deque.
4. if the thread's deque is empty, either:
  - stall at a join
  - steal work from the head of a random thread's associated deque

### Stealing Tasks or Continuations?

```
func main() {
	var fib func(n int) <-chan int
	fib = func(n int) <-chan int {
		result := make(chan int)
		go func() { // tasks
			defer close(result)
			if n <= 2 {
				result <- 1
				return
			}
			result <- <-fib(n-1) + <-fib(n-2)
		}()
		return result // continuation
	}
	fmt.Printf("fib = %d", <-fib(4))
}
```

- In go , goroutines are tasks
- Everything after a goroutine is called is the continuation

Continuation stealing is how Go's work-stealing algorithm is implemented

Go's scheduler has three main concepts:

- G: A Goroutine
- M: An OS thread(also referenced as a machine in the source code)
- P: A context(also referecned as a processor in the source code), GOMAXPROCS


# Appendix

### Anatomy of a Goroutine Error

Prior to Go1.6, when a goroutine panicked, the runtime would print stack traces of all the currently executing
goroutines.

```
package main

import (
	"fmt"
)

func main() {
	//go run -race main.go
	var data int
	go func() {
		data++
	}()
	if data == 0 {
		fmt.Printf("value is %v\n", data)
	}
}
```

### Race Detection

go run -race main.go

### pprof

`runtime/pporf` package has predefined profiles to hook into and display:

- goroutine
- heap
- threadcreate
- block
- primitives
- mutex

```
package main

import (
	"log"
	"os"
	"runtime/pprof"
	"time"
)

func main() {
	log.SetFlags(log.Ltime | log.LUTC)
	log.SetOutput(os.Stdout)

	//every second , log how many goroutines are currently running
	go func() {
		goroutines := pprof.Lookup("goroutine")
		for range time.Tick(1 * time.Second) {
			log.Printf("goroutine count:%d\n", goroutines.Count())
		}
	}()

	//create some goroutines which will never exit
	var blockForever chan struct{}
	for i := 0; i < 10; i++ {
		go func() { <-blockForever }()
		time.Sleep(500 * time.Millisecond)
	}
}
```
