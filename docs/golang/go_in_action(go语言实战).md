《Go 实战》 , https://github.com/goinaction/code

# 4 数组，切片和映射

### 数组

数组类型包括元素类型和长度，只有长度和类型一致才能互相赋值。
注意在函数之间传递数组变量可能有性能问题(copy 数组值)，使用切片(slice)更好。

### 切片

实现：指向底层数组的指针，元素个数，容量

区分 nil和空切片：

```
// nil 切片
var slice []int //nil切片，用于描述不存在的切片,比如函数返回切片但是发生异常了
// nil(pointer), 0(len), 0(cap)


// 空切片,比如数据库查询返回0个结果时表示空集合
slice := make([]int, 0)
// or
slice := []int{}

//NOTE: 注意不管是 nil 切片还是空切片，调用 append, len, cap 效果一样。
//NOTE: 但是如果是做 json 序列化，一个是序列化成 null，一个是空[]，要注意二者之间的区别


newSlice := slice[1:5] // 注意 newSlice, slice 共享了一个底层数组。这点和 py 不一样，py 切片会进行复制
// copy slice
arr := []int{1, 2, 3}
tmp := make([]int, len(arr))
copy(tmp, arr)
fmt.Println(tmp)
fmt.Println(arr)

// for 遍历
for idx, val := range slice {
    // NOTE: val 这里是元素的拷贝，而不是直接引用的切片元素。
    fmt.Println(val)
}
```

### 映射
map使用两个数据结构实现。一个是数组，存储的是选择桶的散列键的高八位，区分每个键值对存储在哪个桶里。
第二个是一个字节数组，用于存储键值对。
传递映射同样不会拷贝副本，所以传递成本很小。


# 5 Go 语言的类型系统

值的类型给编译器提供两个信息:

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

go 语言有两种类型的接收者：值接收者和指针接收者。
值接收者获取的是副本（如果是指针也是指针指向的值的副本）

如果想要修改值就需要用 pointer receivers，但是 pointer receivers 不是并发安全的。
Value receivers are concurrency safe, while pointer receivers are not concurrency safe.

value-receiver-vs-pointer-receiver-in-golang: https://stackoverflow.com/questions/27775376/value-receiver-vs-pointer-receiver-in-golang

The rule about pointers vs. values for receivers is that value methods can be invoked on pointers and values, but pointer methods can only be invoked on pointers


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
		{ //大括号只是为了让临界区看起来更清晰
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
		ball, ok := <-court // 注意 val,ok 语法
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


# 7 并发模式

### runner

使用通道监视程序运行时间，终止程序等。当开发需要后台处理任务程序的时候，比较有用。
```
// runner.go
// Example is provided with help by Gabriel Aszalos.
// Package runner manages the running and lifetime of a process.
package runner

import (
	"errors"
	"os"
	"os/signal"
	"time"
)

// Runner runs a set of tasks within a given timeout and can be
// shut down on an operating system interrupt.
type Runner struct {
	// interrupt channel reports a signal from the
	// operating system.
	interrupt chan os.Signal

	// complete channel reports that processing is done.
	complete chan error

	// timeout reports that time has run out.
	timeout <-chan time.Time

	// tasks holds a set of functions that are executed
	// synchronously in index order.
	tasks []func(int)
}

// ErrTimeout is returned when a value is received on the timeout channel.
var ErrTimeout = errors.New("received timeout")

// ErrInterrupt is returned when an event from the OS is received.
var ErrInterrupt = errors.New("received interrupt")

// New returns a new ready-to-use Runner.
func New(d time.Duration) *Runner {
	return &Runner{
		interrupt: make(chan os.Signal, 1),
		complete:  make(chan error),
		timeout:   time.After(d),
	}
}

// Add attaches tasks to the Runner. A task is a function that
// takes an int ID.
func (r *Runner) Add(tasks ...func(int)) {
	r.tasks = append(r.tasks, tasks...)
}

// Start runs all tasks and monitors channel events.
func (r *Runner) Start() error {
	// We want to receive all interrupt based signals.
	signal.Notify(r.interrupt, os.Interrupt)

	// Run the different tasks on a different goroutine.
	go func() {
		r.complete <- r.run()
	}()

	select {
	// Signaled when processing is done.
	case err := <-r.complete:
		return err

	// Signaled when we run out of time.
	case <-r.timeout:
		return ErrTimeout
	}
}

// run executes each registered task.
func (r *Runner) run() error {
	for id, task := range r.tasks {
		// Check for an interrupt signal from the OS.
		if r.gotInterrupt() {
			return ErrInterrupt
		}

		// Execute the registered task.
		task(id)
	}

	return nil
}

// gotInterrupt verifies if the interrupt signal has been issued.
func (r *Runner) gotInterrupt() bool {
	select {
	// Signaled when an interrupt event is sent.
	case <-r.interrupt:
		// Stop receiving any further signals.
		signal.Stop(r.interrupt)
		return true

	// Continue running as normal.
	default:
		return false
	}
}
```

支持终止方式：

- 程序在分配的时间之内完成工作，正常终止
- 没有及时完成，『自杀』
- 接收到 os 发送的中断事件，程序试图立刻清理状态并停止工作


测试代码如下：

```
// This sample program demonstrates how to use a channel to
// monitor the amount of time the program is running and terminate
// the program if it runs too long.
package main

import (
	"log"
	"os"
	"time"

	"github.com/goinaction/code/chapter7/patterns/runner"
)

// timeout is the number of second the program has to finish.
const timeout = 3 * time.Second

// main is the entry point for the program.
func main() {
	log.Println("Starting work.")

	// Create a new timer value for this run.
	r := runner.New(timeout)

	// Add the tasks to be run.
	r.Add(createTask(), createTask(), createTask())

	// Run the tasks and handle the result.
	if err := r.Start(); err != nil {
		switch err {
		case runner.ErrTimeout:
			log.Println("Terminating due to timeout.")
			os.Exit(1)
		case runner.ErrInterrupt:
			log.Println("Terminating due to interrupt.")
			os.Exit(2)
		}
	}

	log.Println("Process ended.")
}

// createTask returns an example task that sleeps for the specified
// number of seconds based on the id.
func createTask() func(int) {
	return func(id int) {
		log.Printf("Processor - Task #%d.", id)
		time.Sleep(time.Duration(id) * time.Second)
	}
}
```

### pool

使用有缓冲的通道实现资源池，来管理可以在任意数量的 goroutine 之间共享以及独立使用的资源。
pool模式在共享一组静态资源（数据库连接，内存缓冲区）非常有用。

```
// pool.go
// Example provided with help from Fatih Arslan and Gabriel Aszalos.
// Package pool manages a user defined set of resources.
package pool

import (
	"errors"
	"io"
	"log"
	"sync"
)

// Pool manages a set of resources that can be shared safely by
// multiple goroutines. The resource being managed must implement
// the io.Closer interface.
type Pool struct {
	m         sync.Mutex
	resources chan io.Closer
	factory   func() (io.Closer, error)
	closed    bool
}

// ErrPoolClosed is returned when an Acquire returns on a
// closed pool.
var ErrPoolClosed = errors.New("Pool has been closed.")

// New creates a pool that manages resources. A pool requires a
// function that can allocate a new resource and the size of
// the pool.
func New(fn func() (io.Closer, error), size uint) (*Pool, error) {
	if size <= 0 {
		return nil, errors.New("Size value too small.")
	}

	return &Pool{
		factory:   fn,
		resources: make(chan io.Closer, size),
	}, nil
}

// Acquire retrieves a resource	from the pool.
func (p *Pool) Acquire() (io.Closer, error) {
	select {
	// Check for a free resource.
	case r, ok := <-p.resources:
		log.Println("Acquire:", "Shared Resource")
		if !ok {
			return nil, ErrPoolClosed
		}
		return r, nil

	// Provide a new resource since there are none available.
	default:
		log.Println("Acquire:", "New Resource")
		return p.factory()
	}
}

// Release places a new resource onto the pool.
func (p *Pool) Release(r io.Closer) {
	// Secure this operation with the Close operation.
	p.m.Lock()
	defer p.m.Unlock()

	// If the pool is closed, discard the resource.
	if p.closed {
		r.Close()
		return
	}

	select {
	// Attempt to place the new resource on the queue.
	case p.resources <- r:
		log.Println("Release:", "In Queue")

	// If the queue is already at cap we close the resource.
	default:
		log.Println("Release:", "Closing")
		r.Close()
	}
}

// Close will shutdown the pool and close all existing resources.
func (p *Pool) Close() {
	// Secure this operation with the Release operation.
	p.m.Lock()
	defer p.m.Unlock()

	// If the pool is already close, don't do anything.
	if p.closed {
		return
	}

	// Set the pool as closed.
	p.closed = true

	// Close the channel before we drain the channel of its
	// resources. If we don't do this, we will have a deadlock.
	close(p.resources)

	// Close the resources
	for r := range p.resources {
		r.Close()
	}
}
```

测试代码如下，注意实现 closer:

```
// This sample program demonstrates how to use the pool package
// to share a simulated set of database connections.
package main

import (
	"io"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/goinaction/code/chapter7/patterns/pool"
)

const (
	maxGoroutines   = 25 // the number of routines to use.
	pooledResources = 2  // number of resources in the pool
)

// dbConnection simulates a resource to share.
type dbConnection struct {
	ID int32
}

// Close implements the io.Closer interface so dbConnection
// can be managed by the pool. Close performs any resource
// release management.
func (dbConn *dbConnection) Close() error {
	log.Println("Close: Connection", dbConn.ID)
	return nil
}

// idCounter provides support for giving each connection a unique id.
var idCounter int32

// createConnection is a factory method that will be called by
// the pool when a new connection is needed.
func createConnection() (io.Closer, error) {
	id := atomic.AddInt32(&idCounter, 1)
	log.Println("Create: New Connection", id)

	return &dbConnection{id}, nil
}

// main is the entry point for all Go programs.
func main() {
	var wg sync.WaitGroup
	wg.Add(maxGoroutines)

	// Create the pool to manage our connections.
	p, err := pool.New(createConnection, pooledResources)
	if err != nil {
		log.Println(err)
	}

	// Perform queries using connections from the pool.
	for query := 0; query < maxGoroutines; query++ {
		// Each goroutine needs its own copy of the query
		// value else they will all be sharing the same query
		// variable.
		go func(q int) {
			performQueries(q, p)
			wg.Done()
		}(query)
	}

	// Wait for the goroutines to finish.
	wg.Wait()

	// Close the pool.
	log.Println("Shutdown Program.")
	p.Close()
}

// performQueries tests the resource pool of connections.
func performQueries(query int, p *pool.Pool) {
	// Acquire a connection from the pool.
	conn, err := p.Acquire()
	if err != nil {
		log.Println(err)
		return
	}

	// Release the connection back to the pool.
	defer p.Release(conn)

	// Wait to simulate a query response.
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	log.Printf("Query: QID[%d] CID[%d]\n", query, conn.(*dbConnection).ID)
}
```

### work

使用无缓冲的通道创建一个goroutine池，这些goroutine执行并控制一组工作，让其并发执行。

```
// Example provided with help from Jason Waldrip.
// Package work manages a pool of goroutines to perform work.
package work

import "sync"

// Worker must be implemented by types that want to use
// the work pool.
type Worker interface {
	Task()
}

// Pool provides a pool of goroutines that can execute any Worker
// tasks that are submitted.
type Pool struct {
	work chan Worker
	wg   sync.WaitGroup
}

// New creates a new work pool.
func New(maxGoroutines int) *Pool {
	p := Pool{
		work: make(chan Worker),
	}

	p.wg.Add(maxGoroutines)
	for i := 0; i < maxGoroutines; i++ {
		go func() {
			for w := range p.work {
				w.Task()
			}
			p.wg.Done()
		}()
	}

	return &p
}

// Run submits work to the pool.
func (p *Pool) Run(w Worker) {
	p.work <- w
}

// Shutdown waits for all the goroutines to shutdown.
func (p *Pool) Shutdown() {
	close(p.work)
	p.wg.Wait()
}
```

测试代码：
```
// This sample program demonstrates how to use the work package
// to use a pool of goroutines to get work done.
package main

import (
	"log"
	"sync"
	"time"

	"github.com/goinaction/code/chapter7/patterns/work"
)

// names provides a set of names to display.
var names = []string{
	"steve",
	"bob",
	"mary",
	"therese",
	"jason",
}

// namePrinter provides special support for printing names.
type namePrinter struct {
	name string
}

// Task implements the Worker interface.
func (m *namePrinter) Task() {
	log.Println(m.name)
	time.Sleep(time.Second)
}

// main is the entry point for all Go programs.
func main() {
	// Create a work pool with 2 goroutines.
	p := work.New(2)

	var wg sync.WaitGroup
	wg.Add(100 * len(names))

	for i := 0; i < 100; i++ {
		// Iterate over the slice of names.
		for _, name := range names {
			// Create a namePrinter and provide the
			// specific name.
			np := namePrinter{
				name: name,
			}

			go func() {
				// Submit the task to be worked on. When RunTask
				// returns we know it is being handled.
				p.Run(&np)
				wg.Done()
			}()
		}
	}

	wg.Wait()

	// Shutdown the work pool and wait for all existing work
	// to be completed.
	p.Shutdown()
}
```

# 8 标准库

### 标注库
标准库的代码经过预编译的，这些预编译后的文件，称作归档文件(archive file, .a)，放在 pkg 下。

### log
unix 架构创建了  stderr 设备作为日志的默认输出地，把程序输出和日志分离开。
如果用户程序只有记录日志，更常用的方式是将一般的日志写到 stdout， 错误或者警告写到 stderr。

NOTE: 标准 log 记录是 goroutine 安全的。
```
// This sample program demonstrates how to create customized loggers.
package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
)

var (
	Trace   *log.Logger // Just about anything
	Info    *log.Logger // Important information
	Warning *log.Logger // Be concerned
	Error   *log.Logger // Critical problem
)

func init() {
	file, err := os.OpenFile("errors.txt",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open error log file:", err)
	}

	Trace = log.New(ioutil.Discard,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(os.Stdout,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(os.Stdout,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(io.MultiWriter(file, os.Stderr),
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
	Trace.Println("I have something standard to say")
	Info.Println("Special Information")
	Warning.Println("There is something you need to know about")
	Error.Println("Something has failed")
}
```

### 序列化和反序列化：marshal

- 序列化(marshal): 数据-> json
- 反序列化(unmarshal): json -> 数据

### 输入和输出：Writer/Reader

```
package main

import (
      "bytes"
      "fmt"
      "os"
)

func main() {
      var b bytes.Buffer
      b.Write([]byte("hello "))
      fmt.Fprintf(b, "world")
      b.WriteTo(os.Stdout)
}
```

```
package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	r, err := http.Get(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}
	file, err := os.Create(os.Args[2])
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	dest := io.MultiWriter(os.Stdout, file)
	io.Copy(dest, r.Body)
	if err := r.Body.Close(); err != nil {
		log.Println(err)
	}
}
```

```
// 实现一个简单的 curl 请求，同时把返回结果写到 stdout 和文件
package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

// main is the entry point for the application.
func main() {
	// r here is a response, and r.Body is an io.Reader.
	r, err := http.Get(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}

	// Create a file to persist the response.
	file, err := os.Create(os.Args[2])
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	// Use MultiWriter so we can write to stdout and
	// a file on the same write operation.
	dest := io.MultiWriter(os.Stdout, file)

	// Read the response and write to both destinations.
	io.Copy(dest, r.Body)
	if err := r.Body.Close(); err != nil {
		log.Println(err)
	}
}
```


# 9 测试和性能

### 单元测试
单测：用来测试包或者程序的一部分代码或者一组代码的函数。目的是确认目标代码在给定场景下，是否按照预期工作。

基础测试(basic test)和表组(table test)测试（多个测试用例）

```
package listing01

import (
	"net/http"
	"testing"
)

const checkMark = "\u2713"
const ballotX = "\u2717"

// TestDownload validates the http Get function can download content.
func TestDownload(t *testing.T) {
	url := "http://www.goinggo.net/feeds/posts/default?alt=rss"
	statusCode := 200

	t.Log("Given the need to test downloading content.")
	{
		t.Logf("\tWhen checking \"%s\" for status code \"%d\"",
			url, statusCode)
		{
			resp, err := http.Get(url)
			if err != nil {
				t.Fatal("\t\tShould be able to make the Get call.",
					ballotX, err)
			}
			t.Log("\t\tShould be able to make the Get call.",
				checkMark)

			defer resp.Body.Close()

			if resp.StatusCode == statusCode {
				t.Logf("\t\tShould receive a \"%d\" status. %v",
					statusCode, checkMark)
			} else {
				t.Errorf("\t\tShould receive a \"%d\" status. %v %v",
					statusCode, ballotX, resp.StatusCode)
			}
		}
	}
}
```
注意测试文件用 test_file 开头，测试函数需要 Test 开头并且接收 testing.T 指针，不能有返回值。

如果有多组测试用例，使用表组测试:

```
package listing08

import (
	"net/http"
	"testing"
)

const checkMark = "\u2713"   //输出对勾符号
const ballotX = "\u2717"

// TestDownload validates the http Get function can download
// content and handles different status conditions properly.
func TestDownload(t *testing.T) {
	var urls = []struct {
		url        string
		statusCode int
	}{
		{
			"http://www.goinggo.net/feeds/posts/default?alt=rss",
			http.StatusOK,
		},
		{
			"http://rss.cnn.com/rss/cnn_topstbadurl.rss",
			http.StatusNotFound,
		},
	}

	t.Log("Given the need to test downloading different content.")
	{
		for _, u := range urls {
			t.Logf("\tWhen checking \"%s\" for status code \"%d\"",
				u.url, u.statusCode)
			{
				resp, err := http.Get(u.url)
				if err != nil {
					t.Fatal("\t\tShould be able to Get the url.",
						ballotX, err)
				}
				t.Log("\t\tShould be able to Get the url.",
					checkMark)

				defer resp.Body.Close()

				if resp.StatusCode == u.statusCode {
					t.Logf("\t\tShould have a \"%d\" status. %v",
						u.statusCode, checkMark)
				} else {
					t.Errorf("\t\tShould have a \"%d\" status. %v %v",
						u.statusCode, ballotX, resp.StatusCode)
				}
			}
		}
	}
}
```

### mocking: 标准库包含一个 httptest 可以 mock 网络调用

一般单元测试不要依赖网络或者其他服务。mock 是一种常见的手段，用来模拟访问不可用的资源。
使用 httptest 来 mock 一个网络调用。(这个例子麻烦在处理这个返回的 xml)
http和httptest会帮你处理请求哪个端口之类的问题。

```
// Sample test to show how to mock an HTTP GET call internally.
// Differs slightly from the book to show more.
package listing12

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const checkMark = "\u2713"
const ballotX = "\u2717"

// feed is mocking the XML document we except to receive.
var feed = `<?xml version="1.0" encoding="UTF-8"?>
<rss>
<channel>
    <title>Going Go Programming</title>
    <description>Golang : https://github.com/goinggo</description>
    <link>http://www.goinggo.net/</link>
    <item>
        <pubDate>Sun, 15 Mar 2015 15:04:00 +0000</pubDate>
        <title>Object Oriented Programming Mechanics</title>
        <description>Go is an object oriented language.</description>
        <link>http://www.goinggo.net/2015/03/object-oriented</link>
    </item>
</channel>
</rss>`

// mockServer returns a pointer to a server to handle the get call.
func mockServer() *httptest.Server {
	f := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprintln(w, feed)
	}

	return httptest.NewServer(http.HandlerFunc(f))
}

// TestDownload validates the http Get function can download content
// and the content can be unmarshaled and clean.
func TestDownload(t *testing.T) {
	statusCode := http.StatusOK

	server := mockServer()
	defer server.Close()

	t.Log("Given the need to test downloading content.")
	{
		t.Logf("\tWhen checking \"%s\" for status code \"%d\"",
			server.URL, statusCode)
		{
			resp, err := http.Get(server.URL)
			if err != nil {
				t.Fatal("\t\tShould be able to make the Get call.",
					ballotX, err)
			}
			t.Log("\t\tShould be able to make the Get call.",
				checkMark)

			defer resp.Body.Close()

			if resp.StatusCode != statusCode {
				t.Fatalf("\t\tShould receive a \"%d\" status. %v %v",
					statusCode, ballotX, resp.StatusCode)
			}
			t.Logf("\t\tShould receive a \"%d\" status. %v",
				statusCode, checkMark)

			var d Document
			if err := xml.NewDecoder(resp.Body).Decode(&d); err != nil {
				t.Fatal("\t\tShould be able to unmarshal the response.",
					ballotX, err)
			}
			t.Log("\t\tShould be able to unmarshal the response.",
				checkMark)

			if len(d.Channel.Items) == 1 {
				t.Log("\t\tShould have \"1\" item in the feed.",
					checkMark)
			} else {
				t.Error("\t\tShould have \"1\" item in the feed.",
					ballotX, len(d.Channel.Items))
			}
		}
	}
}

// Item defines the fields associated with the item tag in
// the buoy RSS document.
type Item struct {
	XMLName     xml.Name `xml:"item"`
	Title       string   `xml:"title"`
	Description string   `xml:"description"`
	Link        string   `xml:"link"`
}

// Channel defines the fields associated with the channel tag in
// the buoy RSS document.
type Channel struct {
	XMLName     xml.Name `xml:"channel"`
	Title       string   `xml:"title"`
	Description string   `xml:"description"`
	Link        string   `xml:"link"`
	PubDate     string   `xml:"pubDate"`
	Items       []Item   `xml:"item"`
}

// Document defines the fields associated with the buoy RSS document.
type Document struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
	URI     string
}
```

### 测试服务端点(endpoint)
endpoint是指与服务宿主信息无关，用来分辨某个服务的地址，一般是不包含宿主的一个路径。
（感觉这个测试相比动态语言还是麻烦太多了)

```
// Sample test to show how to test the execution of an
// internal endpoint.
package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goinaction/code/chapter9/listing17/handlers"
)

const checkMark = "\u2713"
const ballotX = "\u2717"

func init() {
	handlers.Routes()
}

// TestSendJSON testing the sendjson internal endpoint.
func TestSendJSON(t *testing.T) {
	t.Log("Given the need to test the SendJSON endpoint.")
	{
		req, err := http.NewRequest("GET", "/sendjson", nil)
		if err != nil {
			t.Fatal("\tShould be able to create a request.",
				ballotX, err)
		}
		t.Log("\tShould be able to create a request.",
			checkMark)

		rw := httptest.NewRecorder()
		http.DefaultServeMux.ServeHTTP(rw, req)

		if rw.Code != 200 {
			t.Fatal("\tShould receive \"200\"", ballotX, rw.Code)
		}
		t.Log("\tShould receive \"200\"", checkMark)

		u := struct {
			Name  string
			Email string
		}{}

		if err := json.NewDecoder(rw.Body).Decode(&u); err != nil {
			t.Fatal("\tShould decode the response.", ballotX)
		}
		t.Log("\tShould decode the response.", checkMark)

		if u.Name == "Bill" {
			t.Log("\tShould have a Name.", checkMark)
		} else {
			t.Error("\tShould have a Name.", ballotX, u.Name)
		}

		if u.Email == "bill@ardanstudios.com" {
			t.Log("\tShould have an Email.", checkMark)
		} else {
			t.Error("\tShould have an for Email.", ballotX, u.Email)
		}
	}
}
```

### Example 示例文档
注意函数名需要时基于已经存在的公开的函数或者方法。
```
// Sample test to show how to write a basic example.
package handlers_test

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
)

// ExampleSendJSON provides a basic example.
func ExampleSendJSON() {
	r, _ := http.NewRequest("GET", "/sendjson", nil)
	w := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(w, r)

	var u struct {
		Name  string
		Email string
	}

	if err := json.NewDecoder(w.Body).Decode(&u); err != nil {
		log.Println("ERROR:", err)
	}

	fmt.Println(u)
	// Output:
	// {Bill bill@ardanstudios.com}
}
```
使用 `godoc -http=":3000"` 启动能看到 godoc 多了示例代码。
可以用 `go test -v -run="ExampleSendJSON"` 运行这个示例函数。支持正则和单元测试


### 基准测试(benchmark test)

测试代码性能，不同代码方案的性能。可以识别某段代码的 cpu/mem 效率
比如 go 里边有三种方式将一个整数转为字符串。来测试下哪种最快：

```
// Sample benchmarks to test which function is better for converting
// an integer into a string. First using the fmt.Sprintf function,
// then the strconv.FormatInt function and then strconv.Itoa.
package listing05_test

import (
	"fmt"
	"strconv"
	"testing"
)

// BenchmarkSprintf provides performance numbers for the
// fmt.Sprintf function.
func BenchmarkSprintf(b *testing.B) {  // benchmark 同样用 _test.go 结尾, testing.B 参数表示Benchmark
	number := 10

	b.ResetTimer() //重置计时器，保证测试代码执行之前的初始化代码，不会干扰计时器的结果

	for i := 0; i < b.N; i++ { // 必须把要基准测试的代码都放到循环里，要使用 b.N
		fmt.Sprintf("%d", number)
	} // 测完发现这个最慢
}

// BenchmarkFormat provides performance numbers for the
// strconv.FormatInt function.
func BenchmarkFormat(b *testing.B) {
	number := int64(10)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		strconv.FormatInt(number, 10)
	}
}

// BenchmarkItoa provides performance numbers for the
// strconv.Itoa function.
func BenchmarkItoa(b *testing.B) {
	number := 10

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		strconv.Itoa(number)
	}
}
// go test -v -run="none" -bench="BenchmarkSprintf"
// go test -v -run="none" -bench=. -benchtime="3s"
```

commands:

```sh
# run all package's test
go test ./...
# code coverage
go test -coverprofile=c.out
go tool cover -func=c.out
go tool cover -html=c.out
```

[How to write testable code](https://www.youtube.com/watch?v=BaMcmZRrG-c)
