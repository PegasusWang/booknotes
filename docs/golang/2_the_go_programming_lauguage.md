# 8 Goroutines and Channels

 go 支持两种风格的并发编程。本章讲 goroutines and channels，支持 CSP(communicating sequential processes) 模型的并发。
 CSP模型里值通过独立的 activities(goroutines) 传递但是大部分变量被限定在一个 activity
 里。下一章讲传统的基于多线程共享内存的并发。

## 8.1 Goroutines

在 go 里每个并发执行的活动叫做 goroutine。考虑一个场景，两个函数一个计算一个写输出，但是互不调用，串行方式是第一个调用
完成了再去调用另一个。但是在两个及多个 goroutine 的并发场景下，两个函数可以同时执行。
程序执行的开始，只有一个调用 main 的goroutine，我们叫它 main goroutine，新的 goroutine 可以通过 go
语句创建，语法上就是普通的函数或方法加上 go 关键字，go 声明使得函数在被创建的新的 goroutine 里执行，**go
语句则会立刻返回**，不会阻塞住。

    f()    // call f(); wait for it to return
    go f() // create a new goroutine that calls f(); don't wait

我们看一个好玩的例子，这里输出斐波那契数，在计算的同时，屏幕上会显示一个转动的指针：

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
我们先来写个 tcp serer发送时间：

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
    		go handleConn(conn) // NOTE: 并发处理连接，就是这么简单
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
channel 有两个基本操作，send 和 receive。一个 send 把值从一个 goroutine 发送到另一个对应使用 receive 接收的 goroutine。

    ch <- x // send 语句
    x = <- ch  // 赋值语句中的 receive 表达式
    <- ch  // receive 语句，丢弃结果
    close(ch) // 关闭之后send 将 panic,但是却可以接受

channel 还支持第三个操作 close(ch)，设置 flag 指示没有值将会发送到 channel，后续尝试 send 将会 panic。
在一个关闭的 channel 接收值将会一直接收到 channel 没有剩余的值，之后任何 receive 操作会立刻完成并且接收到 channel
的元素类型零值。
make 创建 channel 还可以指定容量：

    ch = make(chan int)    // unbuffered channel
    ch = make(chan int, 0) // unbuffered channel
    ch = make(chan int, 3) // buffered channel with capacity 3

### 8.4.1 Unbuffered Channels

在一个unbufferd channel 执行 send后 操作会 block 发送goroutine，直到另一个 goroutine 在相同 channel 执行对应的
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
    	<-done // wait for background goroutine to finish，即使main先执行了，也会block到goroutine 执行send
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

    func counter(out chan int)
    func squarer(out, in chan int)
    func printer(in chan int)

go 里提供了单向 channel 允许只接收或者发送值。

-   `chan<- int`, send-only channel of int ，只允许 send
-   `<-chan int`, receive-only channel of int,只允许 receive，close 一个只接收的 channel 会导致编译时错误

好，重构下上边的代码：

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

### 8.4.4 Buffered Channels

有容量的 channel: `ch = make(chan string, 3)`。当 channel 满的时候会 block send 一直到其他goroutine receive 释放了空间。
当 channel 为空的时候接收者被 block 一直到其他 goroutine send 值。
可以用内置 cap 函数获取 channel 容量 `cap(ch)`，而 `len(ch)`返回元素个数。
通常 bufferd channel 在多个 goroutine 使用，go 新手经常在一个 goroutine 里当队列用，这个是错误的，channel 和 goroutine
调度是深度关联的，如果没有另一个 goroutine 从 channel 接收，sender 或者整个程序有可能被永久阻塞。简单的队列应该用
slice。

    func mirroredQuery() string {
    	responses := make(chan string, 3)
    	go func() { responses <- request("asia.gopl.io") }()
    	go func() { responses <- request("europe.gopl.io") }()
    	go func() { responses <- request("americas.gopl.io") }()
    	return <-responses // return the quickest response  ,慢的 goroutine 会泄露
    }
    func request(hostname string) (response string) { /* ... */ }

goroutine leak: goroutine 泄露（视为bug）。NOTE: 泄露的 goroutine 不会被自动回收，必须确定不需要的时候自行终结。

## 8.5 Looping in Parallel

本节介绍几个把循环改成并发执行的几种模式。

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

注意这里的循环操作每个都是独立的，互不依赖，叫做 embarrassingly
parallel，这种方式是最容易实现并发的。你能会立马写出如下代码：

    // NOTE: incorrect!
    func makeThumbnails2(filenames []string) {
    	for _, f := range filenames {
    		go thumbnail.ImageFile(f) // NOTE: ignoring errors
    	}
    }

但是这是不对的，这段代码会启动所有 goroutine，然后没等他们完成直接退出了(运行它你会发现执行很快，但是没卵用，并非是并发的效果)。并没有一种直接的方法来等待 goroutine
完成，但是我们可以让内部的 goroutine 通过向一个共享 channel 发送事件来通知外部 goroutine 它完成了。

    // makeThumbnails3 makes thumbnails of the specified files in parallel.
    func makeThumbnails3(filenames []string) {
    	ch := make(chan struct{})
    	for _, f := range filenames {
    		go func(f string) {
    			thumbnail.ImageFile(f) // NOTE: ignoring errors
    			ch <- struct{}{}
    		}(f) // 注意这里使用参数防止只使用最后一个循环
    	}
    	// wait for goroutine complete
    	for range filenames {
    		<-ch
    	}
    }

下面加上错误处理：

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

注意这里有个隐含的bug，当遇到第一个 non-nil error 时， 返回error给调用者，导致没有 goroutine 消费 errors
channel。每个还在工作的 worker goroutine 想要往 errors channel send 值的时候会被永久阻塞，无法终止，出现了 goroutine
泄露，程序被阻塞或者出现内存被用光。

两种解决方式：简单的方式是用一个有足够空间的 bufferd channel，当它发送消息的时候没有工作的 goruotine 被
block。另一种是在main goroutine 返回第一个 error 的时候创建一个新的 goroutine 消费 channel。

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

当我们不知道会循环多少次的时候，可以使用 sync.WaitGroup 记录 goroutine 数(Add and Done)：

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

## 8.6 Expmple: Conrurrent Web Crawler

    package main

    import (
    	"fmt"
    	"log"
    	"os"

    	"gopl.io/ch5/links"
    )

    func crawl(url string) []string {
    	fmt.Println(url)
    	list, err := links.Extract(url)
    	if err != nil {
    		log.Print(err)
    	}
    	return list
    }

    func main() {
    	worklist := make(chan []string)
    	go func() { worklist <- os.Args[1:] }()
    	// crawl the web concurrently

    	seen := make(map[string]bool)
    	for list := range worklist {
    		for _, link := range list {
    			if !seen[link] {
    				seen[link] = true
    				go func(link string) {
    					worklist <- crawl(link)
    				}(link)
    			}
    		}
    	}
    }

但是这个程序太 "parallel"，我们想限制下它的并发。可以通过有 n 个容量的bufferd channel来限制并发数(counting semaphore)

    // tokens is a counting semaphore used to
    // enforce a limit of 20 concurrent requests.
    var tokens = make(chan struct{}, 20)

    func crawl(url string) []string {
    	fmt.Println(url)
    	tokens <- struct{}{} // acquire a token ，初始化一个空 struct
    	list, err := links.Extract(url)
    	<-tokens // release the token

    	if er != nil {
    		log.Print(err)
    	}
    	return list
    }

但是还有个问题，这个程序不会结束。我们需要当 worklist 为空并且没有爬虫 goroutine 活动的时候结束 main 里的循环。

## 8.7  Multiplexing with select

写一个火箭发射的倒计时程序：

    package main

    import (
    	"fmt"
    	"time"
    )

    //!+
    func main() {
    	fmt.Println("Commencing countdown.")
    	tick := time.Tick(1 * time.Second)
    	for countdown := 10; countdown > 0; countdown-- {
    		fmt.Println(countdown)
    		<-tick
    	}
    	launch()
    }

    func launch() {
    	fmt.Println("Lift off!")
    }

然后增加一个 abort 功能：

    func main() {
    	//!+abort
    	abort := make(chan struct{})
    	go func() {
    		os.Stdin.Read(make([]byte, 1)) // read a single byte
    		abort <- struct{}{}
    	}()
    	//!-abort

    	//!+
    	fmt.Println("Commencing countdown.  Press return to abort.")
    	select {
    	case <-time.After(10 * time.Second):
    		// Do nothing.
    	case <-abort:
    		fmt.Println("Launch aborted!")
    		return
    	}
    	launch()
    }

select 允许我们轮训 channel (polling a channel):

    select {
    case <-abort:
    	fmt.Printf("Launch aborted!\n")
    	return
    default:
    	// do nothing
    }

## 8.8 Example: Concurrent Directory Traversal

这一节实现一个类似 du 的命令，显示目录的使用量。main 函数用两个 goroutine，一个用来背后遍历目录，一个用来打印最后结果。

    package main

    import (
    	"flag"
    	"fmt"
    	"io/ioutil"
    	"os"
    	"path/filepath"
    )

    // 递归访问目录树
    func walkDir(dir string, fileSizes chan<- int64) {
    	for _, entry := range dirents(dir) {
    		if entry.IsDir() {
    			subdir := filepath.Join(dir, entry.Name())
    			walkDir(subdir, fileSizes)
    		} else {
    			fileSizes <- entry.Size()
    		}
    	}
    }

    // 返回目录的入口
    func dirents(dir string) []os.FileInfo {
    	entries, err := ioutil.ReadDir(dir) // returns a slice of os.FileInfo
    	if err != nil {
    		fmt.Fprintf(os.Stderr, "du1: %v\n", err)
    		return nil
    	}
    	return entries
    }

    func main() {
    	// 获取目录
    	flag.Parse()
    	roots := flag.Args()
    	if len(roots) == 0 {
    		roots = []string{"."} //没有命令行参数默认是当前目录
    	}

    	// 遍历文件树
    	fileSizes := make(chan int64)
    	go func() {
    		for _, root := range roots {
    			walkDir(root, fileSizes)
    		}
    		close(fileSizes)
    	}()

    	// 打印结果
    	var nfiles, nbytes int64
    	for size := range fileSizes {
    		nfiles++
    		nbytes += size
    	}
    	printDiskUsage(nfiles, nbytes)
    }
    func printDiskUsage(nfiles, nbytes int64) {
    	fmt.Printf("%d files %.1f GB\n", nfiles, float64(nbytes)/1e9)
    }

如果我们加上个进度输出会更好，用户给了 `-v` 参数，就定期打印出来结果

    func main() {
    	// ...start background goroutine...

    	//!-
    	// Determine the initial directories.
    	flag.Parse()
    	roots := flag.Args()
    	if len(roots) == 0 {
    		roots = []string{"."}
    	}

    	// Traverse the file tree.
    	fileSizes := make(chan int64)
    	go func() {
    		for _, root := range roots {
    			walkDir(root, fileSizes)
    		}
    		close(fileSizes)
    	}()

    	//!+
    	// Print the results periodically.
    	var tick <-chan time.Time
    	if *verbose {
    		tick = time.Tick(500 * time.Millisecond)
    	}
    	var nfiles, nbytes int64
    loop:
    	for {
    		select {
    		case size, ok := <-fileSizes:
    			if !ok {
    				break loop // fileSizes was closed, labeled break statement breaks out of both the select and the for loop;
    			}
    			nfiles++
    			nbytes += size
    		case <-tick:
    			printDiskUsage(nfiles, nbytes)
    		}
    	}
    	printDiskUsage(nfiles, nbytes) // final totals
    }

但是这个程序还是比较慢，我们还可以对每个 WalkDir 调用都开个 goroutine，我们使用sync.WaitGroup计算有多少个活跃的 WalkDir
调用，还有计数同步原语来限制太多的并发数。（从这里开始程序就开始难懂了，😢 )

    package main
    import (
    	"flag"
    	"fmt"
    	"io/ioutil"
    	"os"
    	"path/filepath"
    	"sync"
    	"time"
    )
    var vFlag = flag.Bool("v", false, "show verbose progress messages")
    func main() {
    	flag.Parse()

    	// Determine the initial directories.
    	roots := flag.Args()
    	if len(roots) == 0 {
    		roots = []string{"."}
    	}

    	// Traverse each root of the file tree in parallel.
    	fileSizes := make(chan int64)
    	var n sync.WaitGroup
    	for _, root := range roots {
    		n.Add(1)
    		go walkDir(root, &n, fileSizes)
    	}
    	go func() {
    		n.Wait()
    		close(fileSizes)
    	}()

    	// Print the results periodically.
    	var tick <-chan time.Time
    	if *vFlag {
    		tick = time.Tick(500 * time.Millisecond)
    	}
    	var nfiles, nbytes int64
    loop:
    	for {
    		select {
    		case size, ok := <-fileSizes:
    			if !ok {
    				break loop // fileSizes was closed
    			}
    			nfiles++
    			nbytes += size
    		case <-tick:
    			printDiskUsage(nfiles, nbytes)
    		}
    	}

    	printDiskUsage(nfiles, nbytes) // final totals
    }

    func printDiskUsage(nfiles, nbytes int64) {
    	fmt.Printf("%d files  %.1f GB\n", nfiles, float64(nbytes)/1e9)
    }

    // walkDir recursively walks the file tree rooted at dir
    // and sends the size of each found file on fileSizes.
    func walkDir(dir string, n *sync.WaitGroup, fileSizes chan<- int64) {
    	defer n.Done()
    	for _, entry := range dirents(dir) {
    		if entry.IsDir() {
    			n.Add(1)
    			subdir := filepath.Join(dir, entry.Name())
    			go walkDir(subdir, n, fileSizes)
    		} else {
    			fileSizes <- entry.Size()
    		}
    	}
    }

    //!+sema
    // sema is a counting semaphore for limiting concurrency in dirents.
    var sema = make(chan struct{}, 20)    // 计数同步原语，限制并发数量

    // dirents returns the entries of directory dir.
    func dirents(dir string) []os.FileInfo {
    	sema <- struct{}{}        // acquire token
    	defer func() { <-sema }() // release token

    	entries, err := ioutil.ReadDir(dir)
    	if err != nil {
    		fmt.Fprintf(os.Stderr, "du: %v\n", err)
    		return nil
    	}
    	return entries
    }

## 8.9 Cancellation

有时候我们想让一个 goroutine 工作的时候指示它结束，并没有直接的方法让一个 goroutine 结束其他的
goroutine，因为这会导致它们共享的变量处于未定义状态。对于取消，我们需要一个可靠的机制通过一个 channel 广播事件，让很多
goroutine 能够看到它确实发生了并且之后能看到它已经发生了(For cancellation, what we need is a reliable mechanism to broadcast an event over a channel so that many goroutines can see it as it occurs and can later see that it has occurred.)

    package main
    import (
    	"fmt"
    	"os"
    	"path/filepath"
    	"sync"
    	"time"
    )

    var done = make(chan struct{})

    func cancelled() bool {
    	select {
    	case <-done:
    		return true
    	default:
    		return false
    	}
    }

    func main() {
    	// Determine the initial directories.
    	roots := os.Args[1:]
    	if len(roots) == 0 {
    		roots = []string{"."}
    	}

    	// Cancel traversal when input is detected.
    	go func() {
    		os.Stdin.Read(make([]byte, 1)) // read a single byte
    		close(done)
    	}()

    	// Traverse each root of the file tree in parallel.
    	fileSizes := make(chan int64)
    	var n sync.WaitGroup
    	for _, root := range roots {
    		n.Add(1)
    		go walkDir(root, &n, fileSizes)
    	}
    	go func() {
    		n.Wait()
    		close(fileSizes)
    	}()

    	// Print the results periodically.
    	tick := time.Tick(500 * time.Millisecond)
    	var nfiles, nbytes int64
    loop:
    	//!+3
    	for {
    		select {
    		case <-done:
    			// Drain fileSizes to allow existing goroutines to finish.
    			for range fileSizes {
    				// Do nothing.
    			}
    			return
    		case size, ok := <-fileSizes:
    			// ...
    			//!-3
    			if !ok {
    				break loop // fileSizes was closed
    			}
    			nfiles++
    			nbytes += size
    		case <-tick:
    			printDiskUsage(nfiles, nbytes)
    		}
    	}
    	printDiskUsage(nfiles, nbytes) // final totals
    }

    func printDiskUsage(nfiles, nbytes int64) {
    	fmt.Printf("%d files  %.1f GB\n", nfiles, float64(nbytes)/1e9)
    }

    // walkDir recursively walks the file tree rooted at dir
    // and sends the size of each found file on fileSizes.
    //!+4
    func walkDir(dir string, n *sync.WaitGroup, fileSizes chan<- int64) {
    	defer n.Done()
    	if cancelled() {
    		return
    	}
    	for _, entry := range dirents(dir) {
    		// ...
    		//!-4
    		if entry.IsDir() {
    			n.Add(1)
    			subdir := filepath.Join(dir, entry.Name())
    			go walkDir(subdir, n, fileSizes)
    		} else {
    			fileSizes <- entry.Size()
    		}
    		//!+4
    	}
    }

    //!-4

    var sema = make(chan struct{}, 20) // concurrency-limiting counting semaphore

    // dirents returns the entries of directory dir.
    //!+5
    func dirents(dir string) []os.FileInfo {
    	select {
    	case sema <- struct{}{}: // acquire token
    	case <-done:
    		return nil // cancelled
    	}
    	defer func() { <-sema }() // release token

    	// ...read directory...
    	//!-5

    	f, err := os.Open(dir)
    	if err != nil {
    		fmt.Fprintf(os.Stderr, "du: %v\n", err)
    		return nil
    	}
    	defer f.Close()

    	entries, err := f.Readdir(0) // 0 => no limit; read all entries
    	if err != nil {
    		fmt.Fprintf(os.Stderr, "du: %v\n", err)
    		// Don't return: Readdir may return partial results.
    	}
    	return entries
    }

## 8.10. Example: Chat Server

这一节实现一个聊天室功能结束本章。

    package main

    import (
    	"bufio"
    	"fmt"
    	"log"
    	"net"
    )

    func main() {
    	listener, err := net.Listen("tcp", "localhost:8000")
    	if err != nil {
    		log.Fatal(err)
    	}
    	go broadcaster()

    	for {
    		conn, err := listener.Accept()
    		if err != nil {
    			log.Print(err)
    			continue
    		}
    		go handleConn(conn)
    	}
    }

    type client chan<- string // an outgoing message channel
    var (
    	entering = make(chan client)
    	leaving  = make(chan client)
    	messages = make(chan string) // all incoming client messages
    )

    func broadcaster() {
    	clients := make(map[client]bool) // all connected clients
    	for {
    		select {
    		case msg := <-messages:
    			for cli := range clients {
    				cli <- msg
    			}
    		case cli := <-entering:
    			clients[cli] = true
    		case cli := <-leaving:
    			delete(clients, cli)
    			close(cli)
    		}
    	}
    }

    func handleConn(conn net.Conn) {
    	ch := make(chan string) // outgoing client messages
    	go clientWriter(conn, ch)

    	who := conn.RemoteAddr().String()
    	ch <- "You are " + who
    	messages <- who + "has arrived"
    	entering <- ch

    	input := bufio.NewScanner(conn)
    	for input.Scan() {
    		messages <- who + ": " + input.Text()
    	}

    	leaving <- ch
    	messages <- who + " has left"
    	conn.Close()
    }
    func clientWriter(conn net.Conn, ch <-chan string) {
    	for msg := range ch {
    		fmt.Fprintln(conn, msg)
    	}
    }

# 9. Concurrency with Shared Variables

## 9.1 Race Conditions

在有多个 goroutine 执行的程序中，我们无法知道一个 goroutine 中的事件 x 在另一个 goroutine 中的事件
y之前发生，还是之后发生，或者同时发生。当我们无法确定事件 x 是在 y 之前发生，事件 x 和 y 就是并发的。
考虑下面这段代码：

    var balance int

    func Deposit(amount int) { balance = balance +  amount }    //注意这里不是原子操作，是个先读后写操作
    func Balance() int       { return balance }

我们考虑两个 goroutine 里并发执行的事务：

    	// Alice:
    	go func() {
    		bank.Deposit(200)                // A1，分成读取和更新操作，我们记作 A1r, A1w
    		fmt.Println("=", bank.Balance()) // A2
    	}()
    	// Bob:
    	go bank.Deposit(100) // B

这里如果 Bob 的Deposit 操作在 Alice 的Deposit 操作之间进行，在存款被读取(balance +
amout)之后但是在存款被更新之前(balance = )，就会导致 Bob 的事务丢失。因为 `balance = balance +
amount`操作并非原子的，先读后写，记作 A1r, A1w，数据竞争(data race)问题就出现了：

| opeartion | balance |
| --------- | ------- |
|           | 0       |
| A1r       | 0       |
| B         | 100     |
| A1w       | 200     |
| A2        | "=200"  |

结果是 B 的操作被『丢失』了。**当两个 goroutine 并发访问同一个变量并且至少一个访问包含写操作就会出现数据竞争问题。**
如果变量是序列类型，可能就会出现访问越界等更难以追踪和调试的严重问题。(C语言中未定义行为)。python 中也有类似的问题，
即使有 GIL，但是非原子操作在多线程中也会有数据竞争问题。
根据定义我们有三种解决数据竞争的方式：

-   1.不要写变量。没有被修改或者不可变对象永远是并发安全的，无需同步。
-   2.避免变量被多个 goroutine 访问。之前很多例子都是变量被限定在只有 main goroutine 能访问。如果我们想要更新变量， 可以通过 channel。(Do not communicate by sharing memory; instead, share memory by communicating.)
    限定变量只能通过 channel 代理访问的 goroutine 叫做这个变量的 monitor goroutine。

```
package bank

var deposits = make(chan int) // send amount to deposit
var balances = make(chan int) // receive balance

func Deposit(amount int) { deposits <- amount }
func Balance() int       { return <-balances }

// 把 balance 变量限定在监控 goroutine teller 中
func teller() {
    var balance int // balance 被限定在了 teller goroutine
    for {
	    select {
	    case amount := <-deposits:
		    balance += amount
	    case balances <- balance:
	    }
    }
}
func init() {
    go teller() // start monitor goroutine
}
```

我们还可以通过在 pipeline 中的 goroutine 共享变量，如果 pipeline
中在把变量发送到下一个阶段后都限制访问，所有访问变量就变成了序列化的。In effect, the variable is confined to one stage
of the pipeline, then confined to the next, and so on.

    type Cake struct{ state string }

    func baker(cooked chan<- *Cake) {
    	for {
    		cake := new(Cake)
    		cake.state = "cooked"
    		cooked <- cake // baker never touches this cake again
    	}
    }
    func icer(iced chan<- *Cake, cooked <-chan *Cake) {
    	for cake := range cooked {
    		cake.state = "iced"
    		iced <- cake // icer never touches this cake again
    	}
    }

-   3.允许多个 gorouitne 访问变量，但是一次只允许一个，互斥访问(mutual exclusion)。下一节讨论。

## 9.2 Mutual Exclusion: sync.Mutex

我们可以使用容量为1的 channel 来保证同一时间最多只有一个 goroutine 访问共享变量。数量为一的的信号量叫做二元信号量。

    var (
    	sema    = make(chan struct{}, 1) // a binary semaphore guarding balance
    	balance int
    )

    func Deposit(amount int) {
    	sema <- struct{}{} // acquire token
    	balance = balance + amount
    	<-sema // release token，这里就没 Python 的 context manager 语法糖爽啊
    }

    func Balance() int {
    	sema <- struct{}{} // acquire token
    	b := balance
    	<-sema
    	return b
    }

这种互斥场景太常用了，sync 包提供了 Mutex 类型来支持，调用其 lock 和 unlock 来实现获取和释放 token:

    package bank

    import "sync"

    var (
    	mu      sync.Mutex
    	balance int
    )

    func Deposit(amount int) {
    	mu.Lock()
    	balance = balance + amount
    	mu.Unlock()
    }    // Lock 和 Unlock 之间的共享变量叫做 "critical section"

    func Balance() int {
    	mu.Lock()
    	b := balance
    	mu.Unlock()
    	return b
    }

这样就 ok 了。不过这个例子比较简单，有些复杂的流程里，我们可能会在某些分支忘记 Unlock，或者在 panic 的时候忘记
Unlock，这时候 defer 语句就很有用了。

    func Balance() int {
    	mu.Lock()
    	defer mu.Unlock()   // 使用 defer，连中间变量都不用啦
    	return balance
    }

再来看个例子，Withdraw 取款执行成功减少余额并返回 true，如果余额不够了，恢复余额并且 return false

    func Withdraw(amount int) bool {
    	Deposit(-amount)
    	if Balance() < 0 {
    		Deposit(amount)
    		return false
    	}
    	return true
    }

这里虽然三个操作有 mutex，但是整体上不是序列的，没有锁来保证整个流程。当然，你上来可能会想这么写：

    // NOTE: wrong!!!
    func Withdraw(amount int) bool {
    	mu.Lock()
    	defer mu.Unlock()
    	Deposit(-amount)
    	if Balance() < 0 {
    		Deposit(amount)
    		return false
    	}
    	return true
    }

注意这样做不行滴，因为 mutex locks 是不可重入的(not re-entrant)，不可重入指的是同一个锁无法多次获取，
这段代码Deposit函数执行的时候就会产生死锁，程序被永久 block。有一种通用的解决方式是把 Deposit
函数拆分成俩，一个不可导出函数 deposit(小写)，在工作的时候假定已经持有了锁。一个导出函数 Deposit在调用 deposit
之前用来获取锁。

    package bank

    import "sync"

    var (
    	mu      sync.Mutex // 保护 balance 不会同时被多个 goroutine 访问
    	balance int
    )

    func Withdraw(amount int) bool {
    	mu.Lock()
    	defer mu.Unlock()
    	deposit(-amount) // 小写的内部函数
    	if balance < 0 {
    		deposit(amount)
    		return false
    	}
    	return true
    }

    func Deposit(amount int) {
    	mu.Lock()
    	defer mu.Unlock()
    	deposit(amount)
    }

    func Balance() int {
    	mu.Lock()
    	defer mu.Unlock()
    	return balance
    }

    // NOTE:内部函数，使用之前必须先持有锁
    func deposit(amount int) { balance += amount }

最后需要注意的就是，当你使用 mutex 的时候，确保它和它保护的变量不要被导出(不要大写首字母)，不管它是包级别的变量还是在
struct 里的 field。

## 9.3 Read/Write Mutexes: sync.RWMutex

上边的互斥锁影响到了并发读，如果有场景是想并行读取但是互斥写入，可以用读写锁。(multiple readers, single writer lock)

    // It’s only profitable to use an RWMutex when most of the goroutines that acquire the lock are readers, and the lock is under contention, that is, goroutines routinely have to wait to acquire it. 
    var mu sync.RWMutex
    var balance int

    func Balance() int {
    	mu.RLock()
    	defer mu.RUnlock()  // 防止忘记应该使用了 Lock之后立马就用 defer Unlock
    	return balance
    }

## 9.4 Memory Synchronization

现代计算机都有多个处理器，每个处理器有自己的 cache，为了提升写入效率，通常写入到内存通常是 buffer 满了以后刷到主内存。
go 的同步原语 channel mutex 操作会让处理器把数据刷到主存，这样才能让goroutine的执行结果对跑在其他处理器上的 goroutines
可见。

Where possible, confine variables to a single goroutine; for all other variables, use mutual exclusion.

## 9.5 LazyInitialization: sync.Once

在程序中经常会把一些晚会才使用的变量初始化操作延后

    var icons map[string]image.Image

    func loadIcons() {
    	icons = map[string]image.Image{
    		"spades.png":   loadIcon("spades.png"),
    		"hearts.png":   loadIcon("hearts.png"),
    		"diamonds.png": loadIcon("diamonds.png"),
    		"clubs.png":    loadIcon("clubs.png"),
    	}
    }

    // NOTE: not concurrency-safe!
    func Icon(name string) image.Image {
    	if icons == nil {
    		loadIcons() // one-time initialization
    	}
    	return icons[name]
    }

你可能想到了用 互斥锁 或者用 读写锁来处理:

    var mu sync.Mutex // guards icons
    var icons map[string]image.Image

    // Concurrency-safe.
    func Icon(name string) image.Image {
    	mu.Lock()
    	defer mu.Unlock()
    	if icons == nil {
    		loadIcons()
    	}
    	return icons[name]
    }

互斥锁有个问题，限制了多个 goroutine 并发读，那我们用 读写锁

    var mu sync.RWMutex
    var icons map[string]image.Image

    // Concurrency-safe.
    func Icon(name string) image.Image {
    	mu.RLock()
    	if icons != nil {
    		icon := icons[name]
    		mu.RUnlock()
    		return icon
    	}
    	mu.Unlock()

    	// 获取互斥锁
    	mu.Lock()
    	if icons == nil {
    		loadIcons()
    	}
    	icon := icons[name]
    	mu.Unlock()
    	return icon
    }

但是这样写比较麻烦，sync 提供了 Once 来简化:

    var loadIconsOnce sync.Once
    var icons map[string]image.Image
    // Concurrency-safe.
    func Icon(name string) image.Image {
    	loadIconsOnce.Do(loadIcons)
    	return icons[name]
    }

## 9.6 The Race Detector

涉及到并发的程序太容易出错了，go 提供了一个好用的动态分析工具 race detector，
我们只要给 go build, go run, go test 加上 `-race` 参数就行。

## 9.7 Example: Concurrent Non-Blocking Cache

    package memo

    // Func 是需要缓存的函数类型
    type Func func(key string) (interface{}, error)

    // 调用 Func 的结果
    type result struct {
    	value interface{} // 保存任意类型返回值
    	err   error
    }

    type entry struct {
    	res   result
    	ready chan struct{} // closed when res is ready, to broadcast (§8.9) to any other goroutines that it is now safe for them to read the result from the entry.
    }

    // A request is a message requesting that the Func be applied to key.
    type request struct {
    	key      string
    	response chan<- result // the client wants a single result
    }

    type Memo struct{ requests chan request }

    // New returns a memoization of f.  Clients must subsequently call Close.
    func New(f Func) *Memo {
    	memo := &Memo{requests: make(chan request)}
    	go memo.server(f)
    	return memo
    }

    func (memo *Memo) Get(key string) (interface{}, error) {
    	response := make(chan result)
    	memo.requests <- request{key, response}
    	res := <-response
    	return res.value, res.err
    }

    func (memo *Memo) Close() { close(memo.requests) }

    func (memo *Memo) server(f Func) {
    	cache := make(map[string]*entry)
    	for req := range memo.requests {
    		e := cache[req.key]
    		if e == nil {
    			// This is the first request for this key.
    			e = &entry{ready: make(chan struct{})}
    			cache[req.key] = e
    			go e.call(f, req.key) // call f(key)
    		}
    		go e.deliver(req.response)
    	}
    }

    func (e *entry) call(f Func, key string) {
    	// Evaluate the function.
    	e.res.value, e.res.err = f(key)
    	// Broadcast the ready condition.
    	close(e.ready)
    }

    func (e *entry) deliver(response chan<- result) {
    	// Wait for the ready condition.
    	<-e.ready
    	// Send the result to the client.
    	response <- e.res
    }

## 9.8 Goroutines and Threads

-   Growable Stacks: 操作系统线程通常开辟了固定内存(一般最大
    2M)，保存正在调用函数的 局部变量。这个容量不是不够用就是开辟太多有点浪费。而 goroutine
    起初只需要很小的栈空间，通常只有2KB，并且是按需求增减的。

-   Goroutine Scheduling: OS 线程由操作系统调度，线程上下文切换比较耗时。go 运行时包含自己的调度(m:n scheduling, it
    multiplexes (or schedules) m goroutines on n OS threads)，goroutine 调度非常轻量。

-   GOMAXPROCS: Go 调度器使用一个叫做 GOMAXPROCS 的参数决定在 go 中同时执行几个 OS 线程，就是 m:n scheduling 中的
    n，默认是用的 cpu 核数。runtime.GOMAXPROCS function

-   Goroutines Have No Identity: 很多系统和编程语言提供了识别线程实体的方式，比如 python 里边的
    thread.get_ident()，这使得实现 thread-local 存储非常容易。如果你看过 python 的 flask 框架源码，你会发现就是使用了
    thread local 变量来获取当前请求的 request(这一块是很多初学 flask 的人感觉很魔幻的地方)，thread local 其实就是个全局映射，key
    就是线程标识符（通常就是个数字），值就是不同线程里存储的值。但是 go 不提供方法获取 goroutine 的标识，go
    提倡简单易懂的编程方式，让参数对函数的影响是更加直白、明显。(感觉这就是 python 哲学啊：explicity is better than
    implicity)
