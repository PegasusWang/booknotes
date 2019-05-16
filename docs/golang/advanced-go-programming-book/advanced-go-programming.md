## 1.5.3 顺序一致性内存模型

go 中同一个 goroutine 保证顺序一致性内存模型，但是不同 goroutine 之间不保证，需要定义明确的同步事件。

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	go fmt.Println("hello")
}

// 使用 done 同步
func main() {
	done := make(chan int, 1)

	go func() {
		fmt.Println("hello")
		done <- 1 // 同一个 goroutine 满足顺序一致性，这个时候已经打印
	}()

	<-done
}

// 或者用mutex同步
func main() {
	var mu sync.Mutex
	mu.Lock()
	go func() {
		fmt.Println("hello")
		mu.Unlock() // UhLck一定在 print 之后发生
	}()
	mu.Lock() // Lock一定在unlock之后发生，通过sync.Mutex保证
}
```

## 1.5.4 初始化顺序

![](./init_order.png)

## 1.5.6 基于Channel通信

```go
// 利用channel缓存大小控制并发执行的goroutine最大数
var limit = make(chan int, 3)

func main() {
	for _, w := range work {
		go func() {
			limit <- 1
			w()
			<-limit
		}()
	}
	select {} // 阻塞 main，避免过早退出
}
```

# 1.6 常见并发模式

CSP: 同步通信

> Do not communicate by sharing memory; instead, share memory by communicating.
> 不要通过共享内存来通信，而应通过通信共享内存。

## 1.6.1 并发Hello World

并发编程核心概念是同步通信。

```go
// 错误的示例
package main

import (
	"fmt"
	"sync"
)

func main() {
	var mu sync.Mutex
	mu.Lock()
	go func() {
		fmt.Println("hello")
		mu.Lock()
	}()
	mu.Unlock()  //这里的Lock/Unlock无法保证顺序，必须先Lock之后才能Unlock
}
```

修复方式用main 中使用两次 lock

```go
func main() {
	var mu sync.Mutex
	mu.Lock()
	go func() {
		fmt.Println("hello")
		mu.Unlock() // UhLck一定在 print 之后发生
	}()
	mu.Lock() // Lock一定在unlock之后发生，通过sync.Mutex保证
}
```

使用无缓冲管道实现同步:

```go
func main() {
	done := make(chan int)

	go func() {
		fmt.Println("hello")
		<- done
	}()
	done <- 1   // 发送操作完成才有可能接受，所以会等待先执行 <-done

}
```

对于无缓存Channel 进行的接收，发生在对该 channel 进行发送完成之前。

更好的做法是将管道的发送和接受方调换，使用有缓存的管道。避免同步受到管道缓存大小影响。


```go
func main() {
	done := make(chan int, 1)

	go func() {
		fmt.Println("hello")
		done <- 1
	}()
	<-done
}
```

扩展到 n 个：

```go
package main

import (
	"fmt"
)

func main() {
	done := make(chan int, 10)

	for i := 0; i < cap(done); i++ {
		go func() {
			fmt.Println("hello")
			done <- 1
		}()
	}

	for i := 0; i < cap(done); i++ {
		<-done
	}

}
```
当然最简单的方式使用 sync.WaitGroup , wg.Add wg.Done wg.Wait


### 1.6.2 生产者消费者

