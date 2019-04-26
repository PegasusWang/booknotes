# 4 函数

函数参数过多可以重构为一个符合类型结构，也算是变相实现可选参数和命名实参功能。

```go
// Package main  provides ...
package main

import (
	"log"
	"time"
)

type serverOption struct {
	address string
	port    int
	path    string
	timeout time.Duration
	log     *log.Logger
}

func newOption() *serverOption {
	return &serverOption{
		address: "0.0.0.0",
		port:    8080,
		path:    "/var/test",
		timeout: time.Second * 5,
		log:     nil,
	}
}

func server(option *serverOption) {}
func main() {
	opt := newOption()
	opt.port = 8085
	server(opt)
}
```

闭包(closure)：是在其此法上下文中引用了自由变量的函数。或者说是函数和其引用的环境的组合体。
本质上返回的是一个funcval 结构，可在 runtime/runtime2.go 找到定义。

```go
package main

func test(x int) func() {
	return func() {
		println(x) //匿名函数引用了x
	}
}

func main() {
	f := test(123)
	f()
}
```

闭包通过指针引用环境变量，可能导致其生命周期延长，甚至被分配到堆内存。
延迟求值特性。

```go
package main

func test() []func() {
	var s []func()
	for i := 0; i < 2; i++ {
		s = append(s, func() {
			println(&i, i)
		})
	}
	return s
}

func main() {
	for _, f := range test() {
		f()
	}
}
```

输出格式:for循环复用局部变量i，每次添加的匿名函数引用的自然是同一变量。添加操作仅仅是把匿名函数放入列表，并未执行。
当 main 执行这些函数时，读取的是环境变量 i 最后一次循环的值。

    0xc00001a088 2
    0xc00001a088 2

解决方式是每次用不同的环境变量或者传参复制，让各自的闭包环境各部相同。


```go
func test() []func() {
	var s []func()
	for i := 0; i < 2; i++ {
		x := i
		s = append(s, func() {
			println(&x, x)
		})
	}
	return s
}
```

go 复古地使用了返回错误值（通常是最后一个值），大量函数和方法返回 error，会使得调用方比较难看，
处理方式有：

-  使用专门的检查函数处理错误逻辑（比如记录日志），简化检查代码
- 在不影响逻辑的情况下，使用 defer 延后处理错误状态(err退化赋值)
- 在不中断逻辑的情况下，将错误作为内部状态保存，等最终提交时再处理。

panic, recover: 内置函数而非语句。panic 会立即中断当前函数流程，执行延迟调用。
而在延迟调用函数中，recover 可捕获并返回 panic 提交的错误对象。
连续调用 panic，仅最后一个会被 recover 捕获。
在defer 函数中 panic，不会影响后续延迟调用的执行。而recover之后 panic，可被再次捕获。
recover必须在延迟调用函数中执行才能正常工作。

建议：除非是不可恢复性、导致系统无法正常工作的错误，否则不建议使用
panic。比如文件系统无权限、端口被占用、数据库未启动等。


```go
// func panic(v interface{})
// func recover() interface{}
package main

import "log"

func main() {
	defer func() {
		if err := recover(); err != nil { //捕获错误
			log.Fatalln(err)
		}
	}()
	panic("I am dead")
	println("exit.") // 永不会执行
}



// Package main  provides ...
package main

func test(x, y int) {
	z := 0
	func() { //如果要保护代码片段，只能将其重构为函数调用。利用匿名函数保护"z=x/y"
		defer func() {
			if recover() != nil {
				z = 0
			}
		}()
		z = x / y
	}()
	println("x/y=", z)
}
func main() {
	test(5, 0)
}



//调试阶段可以用 runtime.debug.PrintStack 输出堆栈信息
package main

import "runtime/debug"

func test() {
	panic("i am dead")
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
		}
	}()
	test()
}
```


# 5 数据

### 字符串

字符串操作通常在堆上分配内存，会对 web 等高并发场景造成交大影响，会有大量字符串对象要做垃圾回收。
建议使用[]byte 缓存池活在栈上自行拼装实现 zero-garbage.

rune 用来存储unicode 码点(code point)，int32别名。单引号字面量默认类型就是rune.


### 数组

数组类型包含长度，元素类型相同但是长度不同也不是同一类型，无法比较。
如果元素类型支持比较，数组也支持。
数组是值类型，赋值或者传参都会复制整个数组数据。可以用指针或者切片。

### 切片 slice

    type slice struct {
	    array unsafe.Pointer
	    len   int
	    cap   int
    }


slice 无法比较，只能判断是否为nil。可以用 slice 很容易实现栈


    package main

    import (
	    "errors"
	    "fmt"
    )

    func main() {
	    // cap is 5
	    stack := make([]int, 0, 5)

	    // 入栈
	    push := func(x int) error {
		    n := len(stack)
		    if n == cap(stack) {
			    return errors.New("stack full")
		    }
		    stack = stack[:n+1]
		    stack[n] = x
		    return nil
	    }

	    // 出栈
	    pop := func() (int, error) {
		    n := len(stack)
		    if n == 0 {
			    return 0, errors.New("stack is empty")
		    }
		    x := stack[n-1]
		    stack = stack[:n-1]
		    return x, nil
	    }

	    for i := 0; i < 7; i++ {
		    fmt.Printf("push %d: %v, %v\n", i, push(i), stack)
	    }
	    for i := 0; i < 7; i++ {
		    x, err := pop()
		    fmt.Printf("pop :%d, %v, %v\n", x, err, stack)
	    }

    }


append: append操作向切片尾部追加元素，返回新的切片对象。如果超出 cap限制， 为新切片对象重新分配数组。
向 nil 切片追加数据时，会为其分配底层数组。某些场合建议预留空间，防止重新分配。

    package main

    import "fmt"

    func main() {
	    s := make([]int, 0, 5)
	    s1 := append(s, 10)
	    s2 := append(s1, 20, 30)
	    fmt.Println(s, len(s), cap(s)) //不会修改原slice属性
	    fmt.Println(s1, len(s1), cap(s1))
	    fmt.Println(s2, len(s2), cap(s2))

    }


copy: 在两个切片对象之间复制数据，允许指向同一个底层数组，允许目标区间重叠。最终复制长度一较短的 slice 长度为准。


### 字典

key必须是支持相等运算符(==/!=) 的数据类型，比如数字、字符串、指针、数组、结构体以及对应的接口类型。
字典是引用类型，使用 make函数或者初始化表达语句创建。
对字典进行迭代，每次返回的键值对次序都不同。
字典被设计成 "not addressable"，不能直接修改 value 成员（结构或数组）。
正确的方式是返回 整个 value，待修改后再设置字典键值或直接用指针类型。


    package main

    func main() {
	    type user struct {
		    name string
		    age  byte
	    }
	    m := map[int]user{
		    1: {"Tom", 19},
	    }
	    //m[1].age++ // wrong!!!

	    u := m[1]
	    u.age++
	    m[1] = u

	    m2 := map[int]*user{
		    1: &user{"jack", 20},
	    }

	    m2[1].age++
    }

在迭代期间删除或者新增键值是安全的。
运行时会对字典并发操作做出检测，如果某个任务在对字典进行写操作，其他任务不能对其执行并发操作(读/写/删除)，
否则会导致进程崩溃。可以用 sync.RWMutex 实现同步，避免读写同时进行。


## 结构 struct

struct 将多个不同类型的命名字段(field)序列打包成一个符合类型。
字段名必须唯一，可以用 _ 补齐，支持使用自身指针类型的成员。

空结构可作为通道元素类型，用于事件通知。


    package main

    func main() {

	    exit := make(chan struct{})

	    go func() {
		    println("hello")
		    exit <- struct{}{}
	    }()

	    <-exit
	    println("end.")
    }


字段标签：对字段进行描述的元数据。不属于数据成员，但却是类型的组成部分。运行期间可以用反射获取类型信息。


    package main

    import (
	    "fmt"
	    "reflect"
    )

    type user struct {
	    name string `昵称`
	    sex  byte   `性别`
    }

    func main() {
	    u := user{"Tom", 1}
	    v := reflect.ValueOf(u)
	    t := v.Type()

	    for i, n := 0, t.NumField(); i < n; i++ {
		    fmt.Printf("%s:%v\n", t.Field(i).Tag, v.Field(i))
	    }
    }


结构体分配内存时字段须做对齐处理。通常以所有字段中最长的基础类型宽度为准。


# 方法

方法是与对象实例绑定的特殊函数。可以为当前包，以及除了接口以外的任何类型定义方法。
方法不支持重载。receiver 参数名没有限制，按惯例会选用简短有意义的名称，但不推荐用this/self。
如果方法内部并不引用实例，可省略只保留类型。

    package main

    import "fmt"

    type N int

    func (n N) toString() string {
	    return fmt.Sprintf("%#x", n)
    }

    func main() {
	    var a N = 25
	    println(a.toString())
    }



类型有一个与之相关的方法集(method set)，决定了它是否实现某个接口。


方法和函数一样出了直接调用还可以赋值给变量、或者作为参数传递。按照具体引用方式不同，
分为 expression/value 两种状态。

- method expression: 通过类型引用的 method expression会被还原为普通函数样式，receiver是第一参数，调用时需要显示传参。

- method value:基于实例或者指针引用的method value，参数签名不会改变，按照正常方式调用。
但当method value被赋值给变量或者作为参数传递，会立即计算并复制该方法执行所需的receiver对象，与其绑定，
以便稍后执行时，能隐式传入 receiver 参数。


# 7 接口

接口代表一种调用契约，是多个方法声明的集合。
在某些动态语言里，接口（interface）也被称作协议（protocol）。准备交互的双方，共同遵守事先约定的规则，使得在无须知道对方身份的情况下进行协作。接口要实现的是做什么，而不关心怎么做，谁来做。
接口最常见的使用场景，是对包外提供访问，或预留扩展空间。
Go接口实现机制很简洁，只要目标类型方法集内包含接口声明的全部方法，就被视为实现了该接口，无须做显示声明。当然，目标类型可实现多个接口。
从内部实现来看，接口自身也是一种结构类型，只是编译器会对其做出很多限制。
不能有字段;不能定义自己的方法;只能声明方法不能实现；可嵌入其他接口类型。


    type iface struct {
	    tab  *itab // 类型信息
	    data unsafe.Pointer //实际对象指针
    }


空接口类似于面向对象里的根类型Object，可被赋值为任何类型的对象。接口变量默认值是 nil。如果实现接口的类型支持，可做相等运算。
超集接口变量可以隐式转换为子集，反过来不行。
将对象赋值给接口变量时，会复制该对象
注意只有当接口变量内部两个指针（itab，data） 都为 nil时，接口才等于 nil。这个地方容易出 bug。

    package main

    import "log"

    type TestError struct{}

    func (*TestError) Error() string {
	    return "error"
    }

    func test(x int) (int, error) {
	    var err *TestError
	    if x < 0 {
		    err = new(TestError)
		    x = 0
	    } else {
		    x += 100
	    }
	    return x, err // NOTE：这个 err 有类型的  ，应该明确返回 nil
    }

    func main() {
	    x, err := test(100)
	    if err != nil {
		    log.Fatalln("err!=nil")
	    }
	    println(x)
    }


类型转换：类型推断（type assertiong） 可以把接口变量转换为 原始类型，或者用来判断是否实现了某个更具体的接口类型。

    package main

    import (
	    "fmt"
    )

    type data int

    func (d data) String() string {
	    return fmt.Sprintf("data:%d", d)
    }

    func main() {
	    var d data = 15
	    var x interface{} = d

	    if n, ok := x.(fmt.Stringer); ok { // 转换为更具体的接口类型
		    fmt.Println(n)
	    }

	    if d2, ok := x.(data); ok { // 转换回原始类型
		    fmt.Println(d2)
	    }

	    e := x.(error) // 错误:main.data is not error
	    fmt.Println(e)
    }


# 8 并发

- 并发：逻辑上具备同时处理多个任务的能力
- 并行：物理上在同一时刻执行多个 并发任务

使用 go 关键字就可以创建并发任务。go 并发执行并发操作，而是创建一个并发任务单元。
新建任务被放置在系统队列中，等待调度器安排合适系统线程去获取执行权。
goroutine自定义栈仅需要2KB，自定义栈采用按需分配策略，需要时进行扩容。

和 defer 一样，goroutine也会因『延迟执行』而立即计算并复制执行参数。

    package main

    import "time"

    var c int

    func counter() int {
	    c++
	    return c
    }

    func main() {
	    a := 100

	    go func(x, y int) {
		    time.Sleep(time.Second) //让gortouine在 main 逻辑之后执行
		    println("go:", x, y)
	    }(a, counter()) //立即计算并且复制参数
	    a += 100
	    println("main:", a, counter())

	    time.Sleep(time.Second * 3) //等待goroutine 结束
    }
    // 注意输出：func()参数里的couter()会先计算
    // main: 200 2
    // go: 100 1


Wait:进程退出时不会等待并发任务结束，可用通道(channel)阻塞，然后发出退出信号。

    package main

    import "time"

    func main() {
	    exit := make(chan struct{})

	    go func() {
		    time.Sleep(time.Second)
		    println("goroutine done.")
		    close(exit) //关闭通道，发出信号
	    }()

	    println("main...")
	    <-exit //如果通道关闭，立即解除阻塞
	    println("main exit")
    }

如果等待多个任务结束，推荐使用 sync.WaitGroup。通过设定计数器让每个goroutine 在退出之前递减，直到归零解除阻塞。

    package main

    import (
	    "sync"
	    "time"
    )

    func main() {
	    var wg sync.WaitGroup

	    for i := 0; i < 10; i++ {
		    wg.Add(1)

		    go func(id int) {
			    //wg.Add(1)这里设置可能会来不及执行
			    defer wg.Done()
			    time.Sleep(time.Second)
			    println("goroutine", id, "done.")
		    }(i)
	    }

	    println("main...")
	    wg.Wait()
	    println("main exit.")
    }


GOMAXPROCS: 运行时可能会创建很多线程，但是任何时候仅有限的几个线程参与并发任务执行。
该数量默认和处理器核数相等，可用 runtime.GOMAXPROCS 函数或者环境变量修改。
参数小于1仅返回当前设置值，不做任何调整。


LocalStorage: 与线程不同，goroutine任务无法设置优先级，无法获取编号，没有局部存储（TLS），
甚至连返回值都会被抛弃。但除优先级外，其他功能都很容易实现。


    package main

    import (
	    "fmt"
	    "sync"
    )

    func main() {
	    var wg sync.WaitGroup
	    var gs [5]struct {
		    id     int
		    result int
	    }
	    for i := 0; i < len(gs); i++ {
		    wg.Add(1)
		    go func(id int) { //使用参数避免闭包延迟求值
			    defer wg.Done()
			    gs[id].id = id
			    gs[id].result = (id + 1) * 100
		    }(i)
	    }
	    wg.Wait()
	    fmt.Printf("%+v\n", gs)
    }



Goshed: 暂停，释放线程去执行其他任务，当前任务被放回队列，等待下次调度时恢复执行。
该函数很少被使用，因为运行时会主动向长时间运行（10 ms）的任务发出抢占调度。只是当前版本实现的算法稍显粗糙，不能保证调度总能成功，所以主动切换还有适用场合。


    package main

    import "runtime"

    func main() {
	    runtime.GOMAXPROCS(1)
	    exit := make(chan struct{})

	    go func() { //任务 a
		    defer close(exit)

		    go func() { //任务 b，放在此处为了确保 a 优先制行
			    println("b")
		    }()

		    for i := 0; i < 4; i++ {
			    println("a:", i)
			    if i == 1 { //让出当前线程，调度执行b
				    runtime.Gosched()
			    }
		    }
	    }()

	    <-exit
    }


Goexit: Goexit立即终止当前任务，运行时确保所有已注册延迟调用被执行。该函数不会影响其他并发任务，不会引发panic，自然也就无法捕获。
如果在main.main里调用Goexit，它会等待其他任务结束，然后让进程直接崩溃
标准库函数os.Exit可终止进程，但不会执行延迟调用。

    package main

    import "runtime"

    func main() {
	    exit := make(chan struct{})

	    go func() {
		    defer close(exit)
		    defer println("a")
		    func() {
			    defer func() {
				    println("b", recover() == nil)
			    }()

			    func() {
				    println("c")
				    runtime.Goexit() //立即终止整个调用堆栈
				    println("c done")
			    }()
			    println("b done")

		    }()
		    println("a done")
	    }()
	    <-exit
	    println("main exit")
    }
    /*
    c
    b true
    a
    main exit
    */


通道：Go鼓励使用CSP通道，以通信来代替内存共享，实现并发安全。

Don't communicate by sharing memory，share memory by communicating.
CSP：Communicating Sequential Process.

通过消息来避免竞态的模型除了CSP，还有Actor。但两者有较大区别。

作为CSP核心，通道（channel）是显式的，要求操作双方必须知道数据类型和具体通道，并不关心另一端操作者身份和数量。可如果另一端未准备妥当，或消息未能及时处理时，会阻塞当前端。

除传递消息（数据）外，通道还常被用作事件通知。

相比起来，Actor是透明的，它不在乎数据类型及通道，只要知道接收者信箱即可。默认就是异步方式，发送方对消息是否被接收和处理并不关心。


    package main

    func main() {
	    done := make(chan struct{}) //结束事件
	    c := make(chan string)      // 数据传输通道

	    go func() {
		    s := <-c
		    println(s)
		    close(done) //关闭通道作为结束通知
	    }()

	    c <- "hi" //发送消息
	    <-done    //阻塞直到有数据或者管道关闭
    }


同步模式必须有配对操作的goroutine出现，否则会一直阻塞。而异步模式在缓冲区未满或数据未读完前，不会阻塞。

    package main

    import "fmt"

    func main() {
	    c := make(chan int, 3) //带有3个缓冲槽的异步通道
	    c <- 1
	    c <- 2

	    fmt.Println(<-c)
	    fmt.Println(<-c)
    }

内置函数cap和len返回缓冲区大小和当前已缓冲数量；而对于同步通道则都返回0，据此可判断通道是同步还是异步。

    package main

    func main() {
	    a, b := make(chan int), make(chan int, 3)
	    b <- 1
	    b <- 2
	    println("a:", len(a), cap(a))
	    println("b:", len(b), cap(b))
    }

    // a: 0 0
    // b: 2 3


首发消息：除使用简单的发送和接收操作符外，还可用ok-idom或range模式处理数据。

    package main

    func main() {
	    done := make(chan struct{})
	    c := make(chan int)

	    go func() {
		    defer close(done)
		    for {
			    x, ok := <-c //判断通道是否关闭
			    if !ok {
				    return
			    }
			    println(x)
		    }
		    /*
		    for x:= range c {  //  使用range 更方便点
			    println(x)
		    }
		    */
	    }()

	    c <- 1
	    c <- 2
	    c <- 3
	    close(c)
	    <-done
    }

    // a: 0 0
    // b: 2 3


对于closed或nil通道，发送和接收操作都有相应规则：

- 向已关闭通道发送数据，引发panic。
- 从已关闭接收数据，返回已缓冲数据或零值。
- 无论收发，nil通道都会阻塞。


    package main

    func main() {
	    c := make(chan int, 3)
	    c <- 10
	    c <- 20
	    close(c)

	    for i := 0; i < cap(c)+1; i++ {
		    x, ok := <-c
		    println(i, ":", ok, x)
	    }
    }
    // 0 : true 10
    // 1 : true 20
    // 2 : false 0
    // 3 : false 0


单向通道：通道默认是双向的，并不区分发送和接收端。但某些时候，我们可限制收发操作的方向来获得更严谨的操作逻辑。
尽管可用make创建单向通道，但那没有任何意义。通常使用类型转换来获取单向通道，并分别赋予操作双方。


    package main

    import "sync"

    func main() {
	    var wg sync.WaitGroup
	    wg.Add(2)

	    c := make(chan int)
	    var send chan<- int = c // 注意这里，send 通道
	    var recv <-chan int = c

	    go func() {
		    defer wg.Done()
		    for x := range recv {
			    println(x)
		    }
	    }()

	    go func() {
		    defer wg.Done()
		    defer close(c)

		    for i := 0; i < 3; i++ {
			    send <- i
		    }
	    }()
	    wg.Wait()
    }


选择：如要同时处理多个通道，可选用select语句。它会随机选择一个可用通道做收发操作。

    package main

    import "sync"

    func main() {
	    var wg sync.WaitGroup
	    wg.Add(2)

	    a, b := make(chan int), make(chan int)

	    go func() {
		    defer wg.Done()
		    for {
			    var (
				    name string
				    x    int
				    ok   bool
			    )
			    select { //随机选择可用 channel 接收数据
			    case x, ok = <-a:
				    name = "a"
			    case x, ok = <-b:
				    name = "b"
			    }
			    if !ok { //如果任意一个通道关闭，终止接收
				    return
			    }
			    println(name, x)
		    }
	    }()

	    go func() { //发送端
		    defer wg.Done()
		    defer close(a)
		    defer close(b)
		    for i := 0; i < 10; i++ {
			    select { //随机选择发送channel
			    case a <- i:
			    case b <- i * 10:
			    }
		    }
	    }()
	    wg.Wait()
    }


如要等全部通道消息处理结束（closed），可将已完成通道设置为nil。这样它就会被阻塞，不再被select选中。

    package main

    import "sync"

    func main() {
	    var wg sync.WaitGroup
	    wg.Add(2)

	    a, b := make(chan int), make(chan int)

	    go func() {
		    defer wg.Done()
		    for {
			    select {
			    case x, ok := <-a:
				    if !ok {
					    a = nil //如果通道关闭设置为 nil，阻塞
					    break
				    }
				    println("a", x)
			    case x, ok := <-b:
				    if !ok {
					    b = nil
					    break
				    }
				    println("b", x)
			    }
			    if a == nil && b == nil { //全部结束退出循环
				    return
			    }
		    }
	    }()

	    go func() { //发送端
		    defer wg.Done()
		    defer close(a)
		    for i := 0; i < 3; i++ {
			    a <- i
		    }
	    }()

	    go func() { //发送端
		    defer wg.Done()
		    defer close(b)
		    for i := 0; i < 5; i++ {
			    b <- i * 10
		    }
	    }()

	    wg.Wait()
    }


当所有通道都不可用时，select会执行default语句。如此可避开select阻塞，但须注意处理外层循环，以免陷入空耗。

    package main

    import (
	    "fmt"
	    "time"
    )

    func main() {
	    done := make(chan struct{})
	    c := make(chan int)

	    go func() {
		    defer close(done)

		    for {
			    select {
			    case x, ok := <-c:
				    if !ok {
					    return
				    }
				    fmt.Println("data:", x)
			    default: // 避免 select 阻塞
			    }
			    fmt.Println(time.Now())
			    time.Sleep(time.Second) //避免空耗
		    }
	    }()
	    time.Sleep(time.Second * 5)
	    c <- 100
	    close(c)
	    <-done
    }


也可用default处理一些默认逻辑。


    package main

    func main() {
	    done := make(chan struct{})

	    data := []chan int{ //缓冲数据区
		    make(chan int, 3),
	    }
	    go func() {
		    defer close(done)
		    for i := 0; i < 10; i++ {
			    select {
			    case data[len(data)-1] <- i: //生产数据
			    default: //当前通道已经满，生成新的缓存通道
				    data = append(data, make(chan int, 3))
			    }
		    }
	    }()
	    <-done

	    for i := 0; i < len(data); i++ {
		    c := data[i]
		    close(c)
		    for x := range c {
			    println(x)
		    }
	    }
    }


模式: 通常使用工厂方法将goroutine和通道绑定。

    package main

    import "sync"

    type receiver struct {
	    sync.WaitGroup
	    data chan int
    }

    func newReceiver() *receiver {
	    r := &receiver{
		    data: make(chan int),
	    }

	    r.Add(1)
	    go func() {
		    defer r.Done()
		    for x := range r.data { // 接收消息，直到通道被关闭
			    println("recv:", x)
		    }
	    }()

	    return r
    }

    func main() {
	    r := newReceiver()
	    r.data <- 1
	    r.data <- 2

	    close(r.data) // 关闭通道，发出结束通知
	    r.Wait()      // 等待接收者处理结束
    }



用通道实现信号量（semaphore）。


    package main

    import (
	    "fmt"
	    "runtime"
	    "sync"
	    "time"
    )

    func main() {
	    runtime.GOMAXPROCS(4)
	    var wg sync.WaitGroup

	    sem := make(chan struct{}, 2) // 最多允许2个并发同时执行

	    for i := 0; i < 5; i++ {
		    wg.Add(1)

		    go func(id int) {
			    defer wg.Done()

			    sem <- struct{}{}        //acquire: 获取信号
			    defer func() { <-sem }() //release: 释放信号

			    time.Sleep(time.Second * 2)
			    fmt.Println(id, time.Now())
		    }(i)
	    }

	    wg.Wait()
    }


性能:将发往通道的数据打包，减少传输次数，可有效提升性能。从实现上来说，通道队列依旧使用锁同步机制，单次获取更多数据（批处理），可改善因频繁加锁造成的性能问题。

资源泄漏:通道可能会引发goroutine leak，确切地说，是指goroutine处于发送或接收阻塞状态，但一直未被唤醒。垃圾回收器并不收集此类资源，导致它们会在等待队列里长久休眠，形成资源泄漏。


    package main

    import (
	    "runtime"
	    "time"
    )

    func test() {
	    c := make(chan int)

	    for i := 0; i < 10; i++ {
		    go func() {
			    <-c
		    }()
	    }
    }

    func main() {
	    test()

	    for {
		    time.Sleep(time.Second)
		    runtime.GC() // 强制垃圾回收
	    }
    }

    // $go build-o test
    // $GODEBUG="gctrace=1,schedtrace=1000,scheddetail=1" ./test


### 同步

通道并非用来取代锁的，它们有各自不同的使用场景。通道倾向于解决逻辑层次的并发处理架构，而锁则用来保护局部范围内的数据安全。
标准库sync提供了互斥和读写锁，另有原子操作等，可基本满足日常开发需要。Mutex、RWMutex的使用并不复杂，只有几个地方需要注意。
将Mutex作为匿名字段时，相关方法必须实现为pointer-receiver，否则会因复制导致锁机制失效。

应将Mutex锁粒度控制在最小范围内，及早释放。


    // 错误用法
    func doSomething() {
	    m.Lock()
	    url := cache["key"]
	    http.Get(url) // 该操作并不需要锁保护
	    m.Unlock()
    }

    // 正确用法
    func doSomething() {
	    m.Lock()
	    url := cache["key"]
	    m.Unlock() // 如使用defer，则依旧将Get保护在内
	    http.Get(url)
    }


Mutex不支持递归锁，即便在同一goroutine下也会导致死锁。


    package main

    import "sync"

    func main() {
	    var m sync.Mutex

	    m.Lock()
	    {
		    m.Lock()
		    m.Unlock()
	    }
	    m.Unlock()
    }


- 对性能要求较高时，应避免使用defer Unlock。
- 读写并发时，用RWMutex性能会更好一些。
- 对单个数据读写保护，可尝试用原子操作。
- 执行严格测试，尽可能打开数据竞争检查。


# 9 包结构

依照规范，工作空间（workspace）由src、bin、pkg三个目录组成。通常需要将空间路径添加到GOPATH环境变量列表中，以便相关工具能正常工作。

- GOPATH:编译器等相关工具按GOPATH设置的路径搜索目标。也就是说在导入目标库时，排在列表前面的路径比当前工作空间优先级更高。另外，go get默认将下载的第三方包保存到列表中第一个工作空间内。
- GOROOT:环境变量GOROOT用于指示工具链和标准库的存放位置。在生成工具链时，相关路径就已经嵌入到可执行文件内，故无须额外设置。 除通过设置GOROOT环境变量覆盖内部路径外，还可移动目录（改名、符号链接等），或重新编译工具链来解决。
- GOBIN: 至于GOBIN，则是强制替代工作空间的bin目录，作为go install目标保存路径。这可避免将所有工作空间的bin路径添加到PATH环境变量中。

编译器从标准库开始搜索，然后依次搜索GOPATH列表中的各个工作空间。

    import    "github.com/qyuhen/test"     默认方式:test.A
    import X "github.com/qyuhen/test"     别名方式:X.A
    import .  "github.com/qyuhen/test"     简便方式:A，不推荐使用在正式代码，一般测试中用
    import _  "github.com/qyuhen/test"     初始化方式: 无法引用，仅用来初始化目标包


除工作空间和绝对路径外，部分工具还支持相对路径。可在非工作空间目录下，直接运行、编译一些测试代码。 相对路径：当前目录，或以“./”和“../”开头的路径。
不管是否在test目录下，只要命令行路径正确，就可用go build/run/test进行编译、运行或测试。但因缺少工作空间相关目录，go install会无法工作。
在设置了GOPATH的工作空间中，相对路径会导致编译失败。go run不受此影响，可正常执行。

包（package）由一个或多个保存在同一目录下（不含子目录）的源码文件组成。包的用途类似名字空间（namespace），是成员作用域和访问权限的边界。
包名和目录名并无关系，不要求一致。包名称通常单数形式。
同一目录下所有源码文件必须使用相同包名称。因导入时使用绝对路径，所以在搜索路径下，包必须有唯一路径，但无须是唯一名字。

另有几个被保留、有特殊含义的包名称：

- main：可执行入口（入口函数main.main）。
- all：标准库以及GOPATH中能找到的所有包。
- std，cmd：标准库及工具链。
- documentation：存储文档信息，无法导入（和目录名无关）。


### 权限

所有成员在包内均可访问，无论是否在同一源码文件中。但只有名称首字母大写的为可导出成员，在包外可视。
该规则适用于全局变量、全局常量、类型、结构字段、函数、方法等。 可通过指针转换等方式绕开该限制。

### 初始化

包内每个源码文件都可定义一到多个初始化函数，但编译器不保证执行次序。
实际上，所有这些初始化函数（包括标准库和导入的第三方包）都由编译器自动生成的一个包装函数进行调用，因此可保证在单一线程上执行，且仅执行一次。o
编译器首先确保完成所有全局变量初始化，然后才开始执行初始化函数。直到这些全部结束后，运行时才正式进入main.main入口函数。

### 内部包

在进行代码重构时，我们会将一些内部模块陆续分离出来，以独立包形式维护。此时，基于首字母大小写的访问权限控制就显得过于粗旷。因为我们希望这些包导出成员仅在特定范围内访问，而不是向所有用户公开。
内部包机制相当于增加了新的访问权限控制：所有保存在internal目录下的包（包括自身）仅能被其父目录下的包（含所有层次的子目录）访问。

    src/
    |
    +--main.go
    |
    +--lib/              # 内部包internal、a、b仅能被lib、lib/x、lib/x/y访问
	|
	+--internal/           # 内部包之间可相互访问
	|      |
	|      +--a/        # 可导入外部包，比如lib/x/y
	|      |
	|      +--b/
	|
	+--x/
	    |
	    +--y/


### 依赖管理


引入名为vendor的机制，专门存放第三方包，实现将源码和依赖完整打包分发。


    src/
    |
    +--server/
	|
	+--vendor/
	|     |
	|     +--github.com/
	|           |
	|           +--qyuhen/
	|                 |
	|                 +--test/
	+--main.go


    // main.go
    package main

    import"github.com/qyuhen/test"

    func main() {
	test.Hello()
    }

注意：vendor比标准库优先级更高。

当多个vendor目录嵌套时，如何查找正确目标？要知道引入的第三方包也可能有自己的vendor依赖目录。
规则算不上复杂：从当前源文件所在目录开始，逐级向上构造vendor全路径，直到发现路径匹配的目标为止。匹配失败，则依旧搜索GOPATH。

要使用vendor机制，须开启“GO15VENDOREXPERIMENT=1”环境变量开关（Go 1.6默认开启），且必须是设置了GOPATH的工作空间。
使用go get下载第三方包时，依旧使用GOPATH第一个工作空间，而非vendor目录。
当前工具链中并没有真正意义上的包依赖管理，好在有不少第三方工具可选。


# 10 反射
反射（reflect）让我们能在运行期探知对象的类型信息和内存结构，这从一定程度上弥补了静态语言在动态行为上的不足。同时，反射还是实现元编程的重要手段。

和C数据结构一样，Go对象头部并没有类型指针，通过其自身是无法在运行期获知任何类型相关信息的。反射操作所需的全部信息都源自接口变量。接口变量除存储自身类型外，还会保存实际对象的类型数据。

    func TypeOf(i interface{})Type
    func ValueOf(i interface{})Value

这两个反射入口函数，会将任何传入的对象转换为接口类型。
在面对类型时，需要区分Type和Kind。前者表示真实类型（静态类型），后者表示其基础结构（底层类型）类别。
有一点和想象的不同，反射能探知当前包或外包的非导出结构成员。

可用反射提取struct tag，还能自动分解。其常用于ORM映射，或数据格式验证。


    package main

    import (
	    "fmt"
	    "reflect"
    )

    type user struct {
	    name string `field:"name"type:"varchar(50)"`
	    age  int    `filed:"age"type:"int"`
    }

    func main() {
	    var u user
	    t := reflect.TypeOf(u)

	    for i := 0; i < t.NumField(); i++ {
		    f := t.Field(i)
		    fmt.Printf("%s: %s%s\n", f.Name, f.Tag.Get("field"), f.Tag.Get("type"))
	    }
    }
    // name: namevarchar(50)
    // age: int


### 值

和Type获取类型信息不同，Value专注于对象实例数据读写。

反射在带来“方便”的同时，也造成了很大的困扰。很多人对反射避之不及，因为它会造成很大的性能损失。


# 11 测试

单元测试（unit test）除用来测试逻辑算法是否符合预期外，还承担着监控代码质量的责任。任何时候都可用简单的命令来验证全部功能，找出未完成任务（验收）和任何因修改而造成的错误。它与性能测试、代码覆盖率等一起保障了代码总是在可控范围内，这远比形式化的人工检查要有用得多。

### testing

工具链和标准库自带单元测试框架，这让测试工作变得相对容易。

- 测试代码须放在当前包以“_test.go”结尾的文件中。
- 测试函数以Test为名称前缀。
- 测试命令（go test）忽略以“_”或“.”开头的测试文件。
- 正常编译操作（go build/install）会忽略测试文件。


    // main_test.go
    // go test -v
    package main

    import "testing"

    func add(x, y int) int {
	    return x + y
    }

    func TestAdd(t *testing.T) {
	    if add(1, 2) != 3 {
		    t.FailNow()
	    }
    }


对于测试是否应该和目标放在同一目录，一直有不同看法。某些人认为应该另建一个专门的包用来存放单元测试，且只测试目标公开接口。好处是，当目标内部发生变化时，无须同步维护测试代码。每个人对于测试都有不同理解，就像覆盖率是否要做到90%以上，也是见仁见智。

table driven: 我们可用一种类似数据表的模式来批量输入条件并依次比对结果。


    // main_test.go
    // go test -v
    package main

    import "testing"

    func add(x, y int) int {
	    return x + y
    }

    func TestAdd(t *testing.T) {
	    var tests = []struct {
		    x      int
		    y      int
		    expect int
	    }{
		    {1, 1, 2},
		    {2, 2, 4},
		    {2, 3, 5},
	    }

	    for _, tt := range tests {
		    actual := add(tt.x, tt.y)
		    if actual != tt.expect {
			    t.Errorf("add(%d, %d):expect%d,actual%d", tt.x, tt.y, tt.expect, actual)
		    }
	    }
    }


test main:某些时候，须为测试用例提供初始化和清理操作，但testing并没有setup/teardown机制。解决方法是自定义一个名为TestMain的函数，go test会改为执行该函数，而不再是具体的测试用例。

    func TestMain(m *testing.M) {
	    //setup
	    code := m.Run() // 调用测试用例函数
	    //teardown
	    os.Exit(code) // 注意：os.Exit不会执行defer
    }

M.Run会调用具体的测试用例，但麻烦的是不能为每个测试文件写一个TestMain。
要实现用例组合套件（suite），须借助MainStart自行构建M对象。通过与命令行参数相配合，即可实现不同测试组合。

    func TestMain(m *testing.M) {
	    match := func(pat, str string) (bool, error) { //pat: 命令行参数 -run提供的过滤条件 }}
		    return true, nil //str:InternalTest.Name
	    }

	    tests := []testing.InternalTest{ // 用例列表，可排序
		    {"b", TestB},
		    {"a", TestA},
	    }

	    benchmarks := []testing.InternalBenchmark{}
	    examples := []testing.InternalExample{}

	    m = testing.MainStart(match, tests, benchmarks, examples)
	    os.Exit(m.Run())
    }


example: 例代码最大的用途不是测试，而是导入到GoDoc等工具生成的帮助文档中。它通过比对输出（stdout）结果和内部output注释是否一致来判断是否成功。

    func ExampleAdd() {
	    fmt.Println(add(1, 2))
	    fmt.Println(add(2, 2))

	    //Output:
	    //3
	    //4
    }

如果没有output注释，该示例函数就不会被执行。另外，不能使用内置函数，print/println，因为它们输出到stderr。


### 性能测试
性能测试函数以Benchmark为名称前缀，同样保存在“*_test.go”文件里。


    // main_test.go
    // go test -bench .
    package main

    import "testing"

    func add(x, y int) int {
	    return x + y
    }

    func BenchmarkAdd(b *testing.B) {
	    for i := 0; i < b.N; i++ {
		    _ = add(1, 2)
	    }
    }

测试工具默认不会执行性能测试，须使用bench参数。它通过逐步调整B.N值，反复执行测试函数，直到能获得准确的测量结果。
如果希望仅执行性能测试，那么可以用run=NONE忽略所有单元测试用例。
默认就以并发方式执行测试，但可用cpu参数设定多个并发限制来观察结果。


timer: 如果在测试函数中要执行一些额外操作，那么应该临时阻止计时器工作。

    func BenchmarkAdd(b *testing.B) {
	    time.Sleep(time.Second)
	    b.ResetTimer() // 重置

	    for i := 0; i < b.N; i++ {
		    _ = add(1, 2)

		    if i == 1 {
			    b.StopTimer() // 暂停
			    time.Sleep(time.Second)
			    b.StartTimer() // 恢复
		    }
	    }
    }


memory:性能测试关心的不仅仅是执行时间，还包括在堆上的内存分配，因为内存分配和垃圾回收的相关操作也应计入消耗成本。

    // go test -bench . -benchmem -gcflags "-N -l"  # 禁用内联和优化

    package main

    import "testing"

    func heap() []byte {
	    return make([]byte, 1024*10)
    }

    func BenchmarkHeap(b *testing.B) {
	    for i := 0; i < b.N; i++ {
		    _ = heap()
	    }
    }

输出结果包括单次执行堆内存分配总量和次数。 也可将测试函数设置为总是输出内存分配信息，无论使用benchmem参数与否。

    func BenchmarkHeap(b *testing.B) {
	    b.ReportAllocs()
	    b.ResetTimer()

	    for i := 0; i < b.N; i++ {
		    _ = heap()
	    }
    }

### 代码覆盖率
如果说单元测试和性能测试关注代码质量，那么代码覆盖率（code coverage）就是度量测试自身完整和有效性的一种手段。
关键还是为改进测试提供一个可发现缺陷的机会，毕竟只有测试本身的质量得到保障，才能让它免于成为形式主义摆设。


# 12 工具链


自Go 1.5实现自举（bootstrapping）以后，我们就不得不保留两个版本的Go环境。对于初学者而言，建议先下载C版本的1.4，用GCC完成编译。
自举是指用编译的目标语言编写其编译器，简单点说就是用Go语言编写Go编译器。请提前安装gcc、gdb、binutils等工具。

### 工具

go build: 此命令默认每次都会重新编译除标准库以外的所有依赖包。

    参数          说明                 示例
    ------------------+-------------------------------------------+--------------
    -o          可执行文件名（默认与目录同名）
    -a          强制重新编译所有包（含标准库）
    -p          并行编译所使用的CPU核数量
    -v          显示待编译包名字
    -n          仅显示编译命令，但不执行
    -x          显示正在执行的编译命令
    -work       显示临时工作目录，完成后不删除
    -race       启动数据竞争检查（仅支持amd64）
    -gcflags    编译器参数
    -ldflags    链接器参数

gcflags：

 参数          说明                 示例
------------------+-------------------------------------------+----------
 -B          禁用越界检查
 -N          禁用优化
 -l          禁用内联
 -u          禁用unsafe
 -S          输出汇编代码
 -m          输出优化信息


ldflags：


 参数          说明                 示例
------------------+-------------------------------------------+----------
 -s          禁用符号表
 -w          禁用DRAWF调试信息
 -X          设置字符串全局变量值              -X ver="0.99"
 -H          设置可执行文件格式           -H windowsgui


go install:
和build参数相同，但会将编译结果安装到bin、pkg目录。最关键的是，go install支持增量编译，在没有修改的情况下，会直接链接pkg目录中的静态包。
编译器用buildid检查文件清单和导入依赖，对比现有静态库和所有源文件修改时间来判断源码是否变化，以此来决定是否需要对包进行重新编译。至于buildid算法，实现起来很简单：将包的全部文件名，运行时版本号，所有导入的第三方包信息（路径、buildid）数据合并后哈希。
算法源码请阅读src/cmd/go/pkg.go。


go get: 将第三方包下载（check out）到GOPATH列表的第一个工作空间。默认不会检查更新，须使用“-u”参数。

go env: 显示全部或指定环境参数。

go clean: 清理工作目录，删除编译和安装遗留的目标文件。


### 编译
编译并不仅仅是执行“go build”命令，还有一些须额外注意的内容。
如习惯使用GDB这类调试器，建议编译时添加-gcflags"-N-l"参数阻止优化和内联，否则调试时会有各种“找不到”的情况。

交叉编译: 所谓交叉编译（cross compile），是指在一个平台下编译出其他平台所需的可执行文件。
自Go实现自举后，交叉编译变得更方便。只须使用GOOS、GOARCH环境变量指定目标平台和架构就行。

条件编译: 除在代码中用runtime.GOOS进行判断外，编译器本身就支持文件级别的条件编译。虽说没有C预编译指令那么方便，但是基于文件的组织方式更便于维护。

- 方法一：将平台和架构信息添加到主文件名尾部。
- 方法二：使用build编译指令。
- 方法三：使用自定义tag指令。


预处理：简单点说，就是用go generate命令扫描源码文件，找出所有“go：generate”注释，提取其中的命令并执行。

- 命令必须放在.go源文件中。
- 命令必须以“//go：generate”开头（双斜线后不能有空格）。
- 每个文件可有多条generate命令。
- 命令支持环境变量。
- 必须显式执行go generate命令。
- 按文件名顺序提取命令并执行。
- 串行执行，出错后终止后续命令的执行。
