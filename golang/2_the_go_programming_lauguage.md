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
    	return <-responses // return the quickest response  ,慢的 gortouine 会泄露
    }
    func request(hostname string) (response string) { /* ... */ }

goroutine leak: goroutine 泄露（视为bug）。泄露的 goroutine 不会被自动回收，必须确定不需要的时候自行终结。

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
    	// to nothing
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
