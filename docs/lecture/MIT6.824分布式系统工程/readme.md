# MIT 6.824 分布式系统工程

- http://nil.csail.mit.edu/6.824/2020/schedule.html 课程表，课程表包含 youbute 视频链接，讲义和论文pdf地址，可以自行下载
- https://www.youtube.com/watch?v=cQP8WApzIQQ&list=PLrw6a1wE39_tb2fErI4-WkMbsvGQk9_UB
- https://www.bilibili.com/video/av87684880/ 2020 年最新官方视频

# 其他资料

- https://zhuanlan.zhihu.com/p/34680235
- https://www.zhihu.com/question/29597104
- https://www.bilibili.com/video/av38073607/
- https://pdos.csail.mit.edu/6.824/index.html
- https://github.com/ty4z2008/Qix/blob/master/ds.md#
- https://github.com/chaozh/MIT-6.824
- https://www.v2ex.com/t/574537

# 1. Introduction

why?

- parallelism 并行
- fault tolerance 容错
- physical
- security / isolated

challenges:

- concurrency
- partial failure
- performance

lectures + papers + exams + labs + project(optional)

- Lab1 - MapReduce
- Lab2 - Raft for fault tolerance
- Lab3 - k/v server
- Lab4 - shared k/v service

Infrastructure - Abstractions

- Storage
- Communication
- Computation

Implementation

- RPC
- Threads
- Concurrency

Performance

- Scalability -> 2x computers -> 2x throughput

Fault Tolerance

- Availability
- Recoverability: non-volatile storage; replication

Topic - consistency

- Put(k,v)
- Get(k) -> v
- Strong / Weak

MapReduce (word count):

```
input1 -> Map a,1     b,1
input2 -> Map         b,1
input3 -> Map a,1             c,1

              reduce ------------------------a,2
                     reduce -----------------b,2
                              reduce --------c,1


Map(k, v): k[filename], v[content of this map]
  split v into words
  for w in each word:
    emit(w, "1")

Reduce(k, v):
  emit(len(v))
```

# 2. RPC and Threads

Why Go?

- simple
- type/memory safe
- GC

Threads (or event driven)

- IO concurrency
- Parallelism
- Convenience

Thread challenges:

- race (use lock) `mu.Lock(); n++; mu.UnLock()`
- coordination: channels, sync.Cond, waitGroup
- deadlock

# 3. GFS(Google File System)

paper link: https://pdos.csail.mit.edu/6.824/papers/gfs.pdf

Big Storage. Why Hard

- Performance -> Sharding
- Faults -> Tolerance
- Tolerance -> Replication
- Replication -> Inconsistency
- Consistency -> Low performance

Strong Consistency

Bad Replication Design


GFS:
Big, Fat
Global
Sharding
Automatic recovery

Single data center
Internal use
Big sequential access

![](./3_gfs.png)

- https://github.com/merrymercy/goGFS
- https://github.com/Bereket-G/Google-File-System-Implementation-with-Python

# 4. Primary-Backup Replication

![](./4-0.jpeg)
![](./4-1.jpeg)


# 5. Go, Threads, and Raft

```go
// closure.go
package main

import "sync"

func main() {

	var a string
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		a = "hello world"
		wg.Done()
	}()
	wg.Wait()
	println(a)
}

// loop.go
package main

import "sync"

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(x int) {
			sendRPC(x)
			wg.Done()
		}(i) // 注意闭包变量
	}
	wg.Wait()
}

func sendRPC(i int) {
	println(i)
}

// sleep.go
package main

import "time"

func main() {
	time.Sleep(1 * time.Second)
	println("started")
	go periodic()
	time.Sleep(5 * time.Second)
}

func periodic() {
	for {
		println("tick")
		time.Sleep(1 * time.Second)
	}
}

// sleep-cancel.go
package main

import (
	"sync"
	"time"
)

var done bool
var mu sync.Mutex

func main() {
	time.Sleep(1 * time.Second)
	println("started")
	go periodic()
	time.Sleep(5 * time.Second)

	mu.Lock()
	done = true
	mu.Unlock()
	println("cancelled")
	time.Sleep(3 * time.Second)
}

func periodic() {
	for {
		println("tick")
		time.Sleep(1 * time.Second)
		mu.Lock()
		if done {
			mu.Unlock()
			return
		}
		mu.Unlock()
	}
}


// bank.go
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	alice := 10000
	bob := 10000

	var mu sync.Mutex

	total := alice + bob

	go func() {
		for i := 0; i < 1000; i++ {
            // 这代码演示目的就是 锁的粒度，不能单独给两个操作分别加锁，而是放到一起
            // 这样才能保证不变式 total := alice+bob 始终成立
			mu.Lock()
			alice -= 1
			bob += 1
			mu.Unlock()
		}
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			mu.Lock()
			bob -= 1
			alice += 1
			mu.Unlock()
		}
	}()

	start := time.Now()
	for time.Since(start) < 1*time.Second {
		mu.Lock()
		if alice+bob != total {
			fmt.Printf("observed violation, alice=%v,bob=%v,sum=%v\n", alice, bob, alice+bob)
		}
		mu.Unlock()
	}
}


// vote-count-2.go
package main

import (
	"math/rand"
	"sync"
	"time"
)

// 同步原语演示：演示条件变量 cond var
func main() {
	rand.Seed(time.Now().UnixNano())

	count := 0
	finished := 0
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		go func() {
			vote := requestVote()
			mu.Lock()
			defer mu.Unlock()
			if vote {
				count++
			}
			finished++
		}()
	}
	// bad: 这里浪费 cpu，一个比较low 的实现是在 for 里 time.Sleep(50 * time.Millisecond)
	for {
		mu.Lock()
		if count >= 5 || finished == 10 {
			break
		}
		mu.Unlock()
		// time.Sleep(50 * time.Millisecond)
	}

	if count >= 5 {
		println("recived 5+ votes!")
	} else {
		println("lost")
	}

	mu.Unlock() // NOTE: 注意上边 for 循环里如果 break 了这里需要 unlock
}

func requestVote() bool {
	return rand.Intn(100) < 50 // 50%概率返回 true
}


// vote-count-4.go
package main

import (
	"math/rand"
	"sync"
	"time"
)

// 同步原语演示：演示条件变量 cond var
func main() {
	rand.Seed(time.Now().UnixNano())

	count := 0
	finished := 0
	var mu sync.Mutex
	cond := sync.NewCond(&mu) // 使用条件变量

	for i := 0; i < 10; i++ {
		go func() {
			vote := requestVote()
			mu.Lock()
			defer mu.Unlock()
			if vote {
				count++
			}
			finished++
			cond.Broadcast() // Broadcast wakes all goroutines waiting on c.
		}()
	}

	mu.Lock()
	for count < 5 && finished != 10 {
		cond.Wait()
	}
	if count >= 5 {
		println("received 5+ votes!")
	} else {
		println("lost")
	}
	mu.Unlock()
}

func requestVote() bool {
	return rand.Intn(100) < 50 // 50%概率返回 true
}

/* cond.txt

mu.Lock()
// do something that might affect the condition
cond.Broadcast()
mu.Unlock()

---

mu.Lock()
for !condition {
	cond.Wait()
}
// now condition is true and we have the lock
mu.Unlock()
*/


// channel-unbuffered.go
package main

import (
	"fmt"
	"time"
)

func main() {
	c := make(chan bool)
	go func() {
		time.Sleep(1 * time.Second)
		<-c
	}()

	start := time.Now()
	c <- true // block until other goroutine receives 。
	fmt.Printf("send took %v\n\n", time.Since(start)) // 等待 1s 之后输出 时间
}

// unbuffered-deadlock.go
package main

import "fmt"

// 死锁啦！
func main() {
	c := make(chan bool) // 如果改成 c := make(chan bool, 1) 就不会死锁
	c <- true            // block until  reveive。 如果放到一个 单独 goroutine 里就可以了，可以继续执行以下代码

	// never execute
	fmt.Println("===")
	<-c // 因为上边一直阻塞住了，这里始终不会执行到，最终导致死锁
}


// producer-consumer.go
package main

import (
	"math/rand"
	"time"
)

func main() {
	c := make(chan int)
	for i := 0; i < 4; i++ {
		go doWork(c)
	}
	for {
		v := <-c
		println(v)
	}
}

func doWork(c chan int) {
	for {
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		c <- rand.Int()
	}
}


// waitgroup
package main

import "sync"

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(x int) {
			sendRPC(x)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func sendRPC(i int) {
	println(i)
}
```

lab: https://www.cnblogs.com/mignet/p/6824_Lab_2_Raft_2A.html


# 6. Fault Tolerance-Raft(1)

raft 资料：

- https://raft.github.io/
- https://www.bilibili.com/video/BV1R7411t71W?p=6
- http://thesecretlivesofdata.com/raft/ 这个动画非常清晰易懂
- https://github.com/vision9527/raft-demo
- http://kanaka.github.io/raft.js/
- https://observablehq.com/@stwind/raft-consensus-simulator

### Split Brain

### Majority Vote

2 of 3
2f+1 -> f failures
