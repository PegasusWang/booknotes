# 8 Goroutines and Channels

 go 支持两种风格的并发编程。本章讲 goroutines and channels，支持 CSP(communicating sequential processes) 模型的并发。
 CSP模型里值通过独立的 activities(goroutines) 传递但是大部分变量被限定在一个 activity
 里。下一章讲传统的基于多线程共享内存的并发。

## 8.1 Goroutines

在 go 里每个并发执行的活动叫做 goroutine。考虑一个场景，两个函数一个计算一个写输出，但是互不调用，串行方式是第一个调用
完成了再去调用另一个。但是在两个及多个 goroutine 的并发场景下，两个函数可以同时执行。
程序执行的开始，只有一个调用 main 的goroutine，我们叫它 main goroutine，新的 goroutine 可以通过 go
语句创建，语法上就是普通的函数或方法加上 go 关键字，go 声明使得函数在被创建的新的 goroutine 里执行，go
语句则会立刻返回。

    f()    // call f(); wait for it to return
    go f() // create a new goroutine that calls f(); don't wait

我们看一个好玩的例子，这里输出斐波那契数，在计算的同事，屏幕上会显示一个转动的指针：

    package main

    import (
    	"fmt"
    	"time"
    )

    func main() {
    	go spinner(100 * time.Millisecond)
    	const n = 45
    	fibN := fib(n) // slow
    	fmt.Printf("\rFib(%d) = %d\n", n, fibN)
    }

    func spinner(delay time.Duration) {
    	for {
    		for _, r := range `-\|/` { //实现指针旋转的等待效果
    			fmt.Printf("\r%c", r)
    			time.Sleep(delay)
    		}
    	}
    }
    func fib(x int) int {
    	if x < 2 {
    		return x
    	}
    	return fib(x-1) + fib(x-2)
    }

运行的话就能看到一个指针在转，然后过会就输出了 fib(10)（这个递归计算很耗时）。并没有一种直接的编程方式让一个 goroutine
去结束掉另一个 goroutine,但是有方式可以结束自己。 

## 8.2 Example: Concurrent Clock Server

web server 是最常用的使用并发的地方，要处理来自客户端的大量独立的连接。
我们先来系诶个 tcp serer发送时间：

    package main

    import (
    	"io"
    	"log"
    	"net"
    	"time"
    )

    func main() {
    	listener, err := net.Listen("tcp", "localhost:8000")
    	if err != nil {
    		log.Fatal(err)
    	}
    	for {
    		conn, err := listener.Accept() //blocks until an incoming connection request is made
    		if err != nil {
    			log.Print(err) // e.g., coonnection aborted
    			continue
    		}
    		handleConn(conn) // handle one connection at a time
    	}
    }

    func handleConn(c net.Conn) {
    	defer c.Close()
    	for {
    		_, err := io.WriteString(c, time.Now().Format("15:04:05\n"))
    		if err != nil {
    			return // e.g., client disconnected
    		}
    		time.Sleep(1 * time.Second)

    	}
    }

接下来写一个 只读的 tcp client:

    package main

    import (
    	"io"
    	"log"
    	"net"
    	"os"
    )

    func main() {
    	conn, err := net.Dial("tcp", "localhost:8000")
    	if err != nil {
    		log.Fatal(err)
    	}
    	defer conn.Close()
    	mustCopy(os.Stdout, conn)
    }

    func mustCopy(dst io.Writer, src io.Reader) {
    	if _, err := io.Copy(dst, src); err != nil {
    		log.Fatal(err)
    	}
    }

如果你在两个终端里运行 client，你会发现另一个一直没有输出，server 一次只能处理一个 cilent（server
里是一直循环）。但是重点来了，我们只需要加上一个 go 关键字，就能并发处理多个 client 啦，so easy

    	for {
    		conn, err := listener.Accept() //blocks until an incoming connection request is made
    		if err != nil {
    			log.Print(err) // e.g., coonnection aborted
    			continue
    		}
    		go handleConn(conn) // 并发处理连接，就是这么简单。艹，这一点确实比 python 强多了
    	}

这时候再运行 server，然后打开俩终端同时运行 client，你会发现俩 client 都有输出啦。

## 8.3 Example: Concurrent Echo Server

之前向客户端端输出时间的 clock server对每个连接使用了一个 goroutine，这一节我们对每个连接使用多个 goroutine

    // netcat2
    package main

    import (
    	"bufio"
    	"fmt"
    	"io"
    	"log"
    	"net"
    	"os"
    	"strings"
    	"time"
    )

    func mustCopy(dst io.Writer, src io.Reader) {
    	if _, err := io.Copy(dst, src); err != nil {
    		log.Fatal(err)
    	}
    }

    func echo(c net.Conn, shout string, delay time.Duration) {
    	fmt.Fprintln(c, "\t", strings.ToUpper(shout))
    	time.Sleep(delay)
    	fmt.Fprintln(c, "\t", shout)
    	time.Sleep(delay)
    	fmt.Fprintln(c, "\t", strings.ToLower(shout))
    }

    func handleConn(c net.Conn) {
    	input := bufio.NewScanner(c)
    	for input.Scan() {
    		echo(c, input.Text(), 1*time.Second)
    	}
    	c.Close()
    }

    func main() {
    	conn, err := net.Dial("tcp", "localhost:8000")
    	if err != nil {
    		log.Fatal(err)
    	}
    	defer conn.Close()
    	go mustCopy(os.Stdout, conn)
    	mustCopy(conn, os.Stdin)
    }

    // reverb2
    package main

    import (
    	"bufio"
    	"fmt"
    	"net"
    	"strings"
    	"time"
    )

    func echo(c net.Conn, shout string, delay time.Duration) {
    	fmt.Fprintln(c, "\t", strings.ToUpper(shout))
    	time.Sleep(delay)
    	fmt.FPrintln(c, "\t", shout)
    	time.Sleep(delay)
    	fmt.Fprintln(c, "\t", strings.ToLower(shout))
    }

    func handleConn(c net.Conn) {
    	input := bufio.NewScanner(c)
    	for input.Scan() {
    		go echo(c, input.Text(), 1*time.Second) //input.Text() is evaluated in the main goroutine.
    	}
    	// NOTE: ignoring potential errors from input.Err()
    	c.Close()
    }

## 8.4 Channels

如果 goroutine 是并行 go 程序的活动单元，channels 就是它们之间的连接通道。
channel是一种通信机制允许一个 goroutine 向另一个 goroutine 发送值。每个 channel 都是特定类型的值的管道，称为channel
的元素类型（element type）。比如 `chan int`，使用内置的 make 函数创建管道:
`ch := make(chan int) // ch has type 'chan int'`
和 map 一样，channel 也是引用类型，作为参数传递的时候会拷贝该引用本身，零值是 nil。
channel 有两个基本操作，send 和 receive。一个 send 把值从一个 gortouine 发送到另一个对应使用 receive 接收的 goroutine。

    ch <- x // send 语句
    x = <- ch  // 赋值语句中的 receive 表达式
    <- ch  // receive 语句，丢弃结果

channel 还支持第三个操作 close(ch)，设置 flag 指示没有值将会发送到 channel，后续尝试 send 将会 panic。
在一个关闭的 channel 接收值将会一直接收到 channel 没有剩余的值，之后任何 receive 操作会立刻完成并且接收到 channel
的元素类型零值。
make 创建 channel 还可以指定容量：

    ch = make(chan int)    // unbuffered channel
    ch = make(chan int, 0) // unbuffered channel
    ch = make(chan int, 3) // buffered channel with capacity 3

### 8.4.1 Unbuffered Channels

在一个unbufferd channel 执行 send 操作 block send gortouine 直到另一个 goroutine 在相同 channel 执行对应的
receive，这样值才会被传输然后两个 goroutine 才有可能继续执行。如果 receive 先执行了，会被 block 直到对应的另一个
goroutine 执行了send 操作。在 unbuffered channel 上通信使得接收和发送者 goroutine 同步，所以也叫 synchronous channels。
在并发的讨论中，当我们说 x 在 y 之前发生，并不意味着时间上提前发生，而是保证它之前的操作（更新变量等），已经完成并且可以依赖它们了。
如果说 x 不在 y 前也不在 y 后发生，我们就说 x 和 y 是并发的。

    func main() {
    	conn, err := net.Dial("tcp", "localhost:8000")
    	if err != nil {
    		log.Fatal(err)
    	}
    	done := make(chan struct{})    // 这里的 done 只是为了同步用，不需要值，防止main 结束了有些 coroutine 还没结束
    	go func() {
    		io.Copy(os.Stdout, conn) // NOTE: ignoring errors
    		log.Println("done")
    		done <- struct{}{} // signal the main goroutine
    	}()
    	mustCopy(conn, os.Stdin)
    	conn.Close()
    	<-done // wait for background goroutine to finish
    }

### 8.4.2 Pipelines

channel 可以通过把一个 goroutine 的输出当成另一个的输入，成为 pipeline。

    package main

    import "fmt"

    // counter -> squarer -> printer
    func main() {
    	naturals := make(chan int)
    	squares := make(chan int)

    	// Counter
    	go func() {
    		for x := 0; ; x++ {
    			naturals <- x
    		}
    	}()

    	//Squarer
    	go func() {
    		for {
    			x := <-naturals
    			squares <- x * x
    		}
    	}()

    	// Printer( in main goroutine)
    	for {
    		fmt.Println(<-squares)
    	}
    }

go 提供了 range 语法用来迭代 channels，用来接收一个 channel 上被 send 的所有值，然后迭代到最后一个值后结束循环。
在一个已经 close 或者 nil 的channel执行 close 会 panic

### 8.4.3 Unidirectional Channel Types
上面例子最好能分开成几个函数，如下：
```
func counter(out chan int)
func squarer(out, in chan int)
func printer(in chan int)
```
go 里提供了单向 channel 允许只接收或者发送值。
- `chan<- int`, send-only channel of int ，只允许 send
- `<-chan int`, receive-only channel of int,只允许 receive，close 一个只接收的 channel 会导致编译时错误

好，重构下上边的代码：
```
package main

import "fmt"

// counter -> squarer -> printer
func main() {
	naturals := make(chan int)
	squares := make(chan int)

	go counter(naturals) // 允许双向 channel 隐式转成单向的 channel
	go squarer(squares, naturals)
	printer(squares)
}

func counter(out chan<- int) { // 只发送
	for x := 0; x < 100; x++ {
		out <- x
	}
	close(out)
}
func squarer(out chan<- int, in <-chan int) {
	for v := range in {
		out <- v * v
	}
	close(out)
}
func printer(in <-chan int) {
	for v := range in {
		fmt.Println(v)
	}
}
```

### 8.4.4 Buffered Channels
有容量的 channel: `ch = make(chan string, 3)`。当 channel 满的时候会 block send 一直到其他goroutine receive 释放了空间。
当 channel 为空的时候接收者被 block 一直到其他 goroutine send 值。
可以用内置 cap 函数获取 channel 容量 `cap(ch)`，而 `len(ch)`返回元素个数。
通常 bufferd channel 在多个 goroutine 使用，go 新手经常在一个 goroutine 里当队列用，这个是错误的，channel 和 goroutine
调度是深度关联的，如果没有另一个 goroutine 从 channel 接收，sender 或者整个程序有可能被永久阻塞。简单的队列应该用
slice。

```
func mirroredQuery() string {
	responses := make(chan string, 3)
	go func() { responses <- request("asia.gopl.io") }()
	go func() { responses <- request("europe.gopl.io") }()
	go func() { responses <- request("americas.gopl.io") }()
	return <-responses // return the quickest response  ,慢的 gortouine 会泄露
}
func request(hostname string) (response string) { /* ... */ }
```

goroutine leak: goroutine 泄露（视为bug）。泄露的 goroutine 不会被自动回收，必须确定不需要的时候自行终结。

## 8.5 Looping in Parallel
本节介绍几个把循环改成并发执行的几种模式。

```
package thumbnail

import "log"

func ImageFile(infile string) (string, error)

// makeThumbnails makes thumbnails of the specified files.
func makeThumbnails(filenames []string) {
	for _, f := range filenames {
		if _, err := thumbnail.ImageFile(f); err != nil {
			log.Println(err)
		}
	}
}
```
注意这里的循环操作每个都是独立的，互不依赖，叫做 embarrassingly
parallel，这种方式是最容易实现并发的。你能会立马写出如下代码：
```
// NOTE: incorrect!
func makeThumbnails2(filenames []string) {
	for _, f := range filenames {
		go thumbnail.ImageFile(f) // NOTE: ignoring errors
	}
}
```
但是这是不对的，这段代码会启动所有 goroutine，然后没等他们完成直接退出了(运行它你会发现执行很快，但是没卵用，并非是并发的效果)。并没有一种直接的方法来等待 goroutine
完成，但是我们可以让内部的 goroutine 通过向一个共享 channel 发送事件来通知外部 goroutine 它完成了。

```
// makeThumbnails3 makes thumbnails of the specified files in parallel.
func makeThumbnails3(filenames []string) {
	ch := make(chan struct{})
	for _, f := range filenames {
		go func(f string) {
			thumbnail.ImageFile(f) // NOTE: ignoring errors
			ch <- struct{}{}
		}(f)
	}
	// wait for goroutine complete
	for range filenames {
		<-ch
	}
}
```
下面加上错误处理：

```
// makeThumbnails4 makes thumbnails for the specified files in parallel.
// It returns an error if any step failed.
func makeThumbnails4(filenames []string) error {
	errors := make(chan error)
	for _, f := range filenames {
		go func(f string) {
			_, err := thumbnail.ImageFile(f)
			errors <- err
		}(f)
	}
	for range filenames {
		if err := <-errors; err != nil {
			return err // NOTE: incorrect: goroutine leak! 注意直接返回第一个err 会造成 goroutine 泄露
		}
	}
	return nil
}
```
注意这里有个隐含的bug，当遇到第一个 non-nil error 时， 返回error给调用者，导致没有 goroutine 消费 errors
channel。每个还在工作的 worker goroutine 想要往 errors channel send 值的时候会被永久阻塞，无法终止，出现了 goroutine
泄露，程序被阻塞或者出现内存被用光。
两种解决方式：简单的方式是用一个有足够空间的 bufferd channel，当它发送消息的时候没有工作的 goruotine 被
block。另一种是在main goroutine 返回第一个 error 的时候创建一个新的 goroutine 消费 channel。
```
// makeThumbnails5 makes thumbnails for the specified files in parallel.
// It returns the generated file names in an arbitrary order,
// or an error if any step failed.
func makeThumbnails5(filenames []string) (thumbfiles []string, err error) {
	type item struct {
		thumbfile string
		err       error
	}
	ch := make(chan item, len(filenames)) // buffered channel
	for _, f := range filenames {
		go func(f string) {
			var it item
			it.thumbfile, it.err = thumbnail.ImageFile(f)
			ch <- it
		}(f)
	}

	for range filenames {
		it := <-ch
		if it.err != nil {
			return nil, it.err
		}
		thumbfiles = append(thumbfiles, it.thumbfile)
	}
	return thumbfiles, nil
}
```
当我们不知道会循环多少次的时候，可以使用 sync.WaitGroup 记录 goroutine 数：

```
// makeThumbnails6 makes thumbnails for each file received from the channel.
// It returns the number of bytes occupied by the files it creates.
func makeThumbnails6(filenames <-chan string) int64 {
	sizes := make(chan int64)
	var wg sync.WaitGroup // number of working goroutines
	for f := range filenames {
		wg.Add(1)    // Add 必须在 goroutine 前调用
		// worker
		go func(f string) {
			defer wg.Done()    // 等价于 Add(-1)
			thumb, err := thumbnail.ImageFile(f)
			if err != nil {
				log.Println(err)
				return
			}
			info, _ := os.Stat(thumb) // OK to ignore error
			sizes <- info.Size()
		}(f)
	}
	// closer
	go func() {
		wg.Wait()
		close(sizes)
	}()
	var total int64
	for size := range sizes {
		total += size
	}
	return total
}
```
