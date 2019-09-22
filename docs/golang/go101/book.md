
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
