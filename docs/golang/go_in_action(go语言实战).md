# 5 Go 语言的类型系统


值得类型给编译器提供两个信息:

- 需要分配多少内存（规模）
- 这段内存表示什么

```
type Duration int64


func main() {
	var dur Duration
	dur = int64(100)  // cannot not use64(1000) (type int64) as type Duration
  // 编译器不会做隐式类型转换
}
```

go 语言有两种类型的接受者：值接受者和指针接受者。
值接受者获取的是副本（如果是指针也是指针指向的值的副本）

### 5.3 类型的本质

- 内置类型：数值类型、字符串、布尔类型。传递的副本
- 引用类型: 切片、映射、通道、接口和函数类型。通过复制传递应用类型值的副本，本质上就是共享底层数据结构
- 结构类型：非原始值应该总是用共享传递，而不是复制。

使用值接收还是指针接收不应该由该方法是否修改了接受到的值来决定，而应该基于该类型的本质。
一个例外是需要让类型值符合某个接口的时候。

接口是用来定义行为的类型，被定义的行为不由接口直接实现，而是通过方法由用户定义的类型实现。
如果用户定义的类型实现了某个接口声明的一组方法，那么这个用户定义的类型的值就可以赋值给这个接口类型的值。
这个赋值会把用户定义的类型的值存入接口类型的值。接口值的方法调用是一种多态。

嵌入类型：已有类型嵌入到新的类型里。内部类型和外部类型。如果没有重名的话，外部类型可以直接调用内部类型的方法。


# 6. 并发


GO并发同步模型来自通信顺序进程(Communicating Sequential Processes, CSP)的泛型(paradigm)。
CSP 是一种消息传递模型，通过在goroutine之间传递数据来传递消息，而不是通过对数据加锁来实现同步访问。
用于在 goroutine 之间同步和传递数据的关键数据类型叫做通道(channel)。

进程和线程：进程维护了应用程序运行时的内存地址空间、文件和设备的句柄以及线程。


```
package main

import (
	"fmt"
	"runtime"
	"sync"
)

func main() {
	runtime.GOMAXPROCS(1) // 只能使用一个逻辑处理器

	var wg sync.WaitGroup
	wg.Add(2)
	fmt.Println("Start Goroutines")

	go func() {
		defer wg.Done()
		for count := 0; count < 3; count++ {
			for char := 'a'; char < 'a'+26; char++ {
				fmt.Printf("%c", char)
			}
		}
	}()

	go func() {
		defer wg.Done()
		for count := 0; count < 3; count++ {
			for char := 'A'; char < 'A'+26; char++ {
				fmt.Printf("%c", char)
			}
		}
	}()

	wg.Wait()
	fmt.Println("Waiting To Finish")

}
```

竞争状态：两个或者多个 goroutine 在没有同步的情况的下，访问某个共享的资源，并试图同时读写这个资源，就会处于
相互竞争状态。

```
// 竞争状态演示
package main

import (
	"fmt"
	"runtime"
	"sync"
)

var (
	counter int
	wg      sync.WaitGroup
)

func main() {
	wg.Add(2)
	go incCounter(1)
	go incCounter(2)
	wg.Wait()
	fmt.Println("final counter:", counter)
}

func incCounter(id int) {
	defer wg.Done()
	for count := 0; count < 2; count++ {
		value := counter
		// 用于当前 goroutine 从线程退出，并放回到队列，给其他 gorouine 运行机会
		//这里是为了强制调度器切换两个 goroutine，让竞争状态的效果更明显
		// go build -race 可以用竞争检测器标志来 编译程序
		runtime.Gosched()
		value++
		counter = value
	}
}
```

使用锁来锁住共享资源： 

- 原子函数(atomic)
- 互斥锁(mutex)
- 通道(channel)


```
// 愿子函数
package main

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
)

var (
	counter int64
	wg      sync.WaitGroup
)

func main() {
	wg.Add(2)
	go incCounter(1)
	go incCounter(2)
	wg.Wait()
	fmt.Println("final counter:", counter)
}

func incCounter(id int) {
	defer wg.Done()
	for count := 0; count < 2; count++ {
		// StoreInt64, LoadInt64
		atomic.AddInt64(&counter, 1)
		runtime.Gosched()
	}
}
```

使用互斥锁:


```
// 使用 互斥锁 mutex
package main

import (
	"fmt"
	"runtime"
	"sync"
)

var (
	counter int
	wg      sync.WaitGroup
	mutex   sync.Mutex
)

func main() {
	wg.Add(2)
	go incCounter(1)
	go incCounter(2)
	wg.Wait()
	fmt.Println("final counter:", counter)
}

func incCounter(id int) {
	defer wg.Done()
	for count := 0; count < 2; count++ {
		// 同一时刻只允许一个 goroutine 进入临界区
		mutex.Lock()
		{
			value := counter
			runtime.Gosched()
			value++
			counter = value
		}
		mutex.Unlock()
	}
}
```


使用通道，通过发送和接收需要共享的资源，在 goroutine 之间做同步。
可以通过 Channel 共享内置类型、命名类型、结构类型、和引用类型的值或者指针。

unbuffered channel: 接收前没有能力保存任何值的通道。要求发送和接收的goroutine 同时准备好，才能完成发送和接收。
如果两个 goroutine 没有同时准备好，通道会导致先执行发送或者接收的 goroutine 阻塞等待。行为本身就是同步的。


```
//使用unbufferd channel 模拟网球比赛

package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var wg sync.WaitGroup

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	court := make(chan int)
	wg.Add(2)

	// 启动俩选手
	go player("Nadal", court)
	go player("Djokovic", court)

	// 发球
	court <- 1

	// 等待游戏结束
	wg.Wait()
}

func player(name string, court chan int) {
	defer wg.Done()

	for {
		// 等待球被打回来
		ball, ok := <-court
		if !ok {
			// 如果通道被关闭，我们就 赢了
			fmt.Printf("Player %s Won\n", name)
			return
		}
		n := rand.Intn(100) //随机数判断是否丢球
		if n%13 == 0 {
			fmt.Printf("Player %s Missed\n", name)
			close(court)
			return
		}

		fmt.Printf("Player %s Hit %d\n", name, ball)
		ball++

		court <- ball //把球打到对手

	}
}
```


buffered channel: 在接收前能够存储一个或者多个值的通道。只在通道中没有要接收的值时，接收动作才会阻塞。
只有通道没有可用缓冲区容纳被发送的值，发送动作才会阻塞。
无缓冲通道保证进行发送和接收的 goroutine 会在同一时间进行数据交换；有缓冲通道没有这个保证。


```
// This sample program demonstrates how to use a buffered
// channel to work on multiple tasks with a predefined number
// of goroutines.
package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	numberGoroutines = 4  // Number of goroutines to use.
	taskLoad         = 10 // Amount of work to process.
)

// wg is used to wait for the program to finish.
var wg sync.WaitGroup

// init is called to initialize the package by the
// Go runtime prior to any other code being executed.
func init() {
	// Seed the random number generator.
	rand.Seed(time.Now().Unix())
}

// main is the entry point for all Go programs.
func main() {
	// Create a buffered channel to manage the task load.
	tasks := make(chan string, taskLoad)

	// Launch goroutines to handle the work.
	wg.Add(numberGoroutines)
	for gr := 1; gr <= numberGoroutines; gr++ {
		go worker(tasks, gr)
	}

	// Add a bunch of work to get done.
	for post := 1; post <= taskLoad; post++ {
		tasks <- fmt.Sprintf("Task : %d", post)
	}

	// Close the channel so the goroutines will quit
	// when all the work is done.
	close(tasks) // 关闭后 goroutine 依旧可以接收数据，但是不能再发送（要能够获取剩下的所有值）

	// Wait for all the work to get done.
	wg.Wait()
}

// worker is launched as a goroutine to process work from
// the buffered channel.
func worker(tasks chan string, worker int) {
	// Report that we just returned.
	defer wg.Done()

	for {
		// Wait for work to be assigned. 会阻塞在这里等待接收值
		task, ok := <-tasks
		if !ok {
			// This means the channel is empty and closed.
			fmt.Printf("Worker: %d : Shutting Down\n", worker)
			return
		}

		// Display we are starting the work.
		fmt.Printf("Worker: %d : Started %s\n", worker, task)

		// Randomly wait to simulate work time.
		sleep := rand.Int63n(100)
		time.Sleep(time.Duration(sleep) * time.Millisecond)

		// Display we finished the work.
		fmt.Printf("Worker: %d : Completed %s\n", worker, task)
	}
}
```
