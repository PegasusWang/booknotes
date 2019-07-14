七牛-许式伟


# 2 顺序编程

字符类型包括俩：rune/byte(uint8别名)，rune 代表单个 unicode 字符

```go
// go 如何比较 float
import "math"
// p为用户自定义的比较精度，比如0.00001
func IsEqual(f1, f2, p float64) bool {
    return math.Fdim(f1, f2) < p
}
```


append 元素： append(slice1, slice2...)


map: map delete操作如果 key 不存在不会发生作用，也没有任何副作用


不定参数：

```go
import "fmt"

// ...type 本质上一个数组切片，`
func f(args ...int) {
	for _, arg := range args {
		fmt.Println(arg)
		myfunc(args...) // 透传可变参数
	}
}
```

### 闭包

闭包：包含是可以包含自由(未绑定到特定对象)变量的代码块，这些变量不在这个代码块内或者任何全局上下文中定义，
而是在定义代码块的环境中定义。要执行的代码块（由于自由变量包含在代码块中，所以这些自由变量和它们引用的对象都没有被释放）
为自由变量提供绑定的计算环境（作用域）

闭包的价值：可以作为函数对象或者匿名函数。支持闭包的多数语言都把函数作为一级对象，这些函数可以存储到变量中作为参数传递给
其他函数，还能被函数动态创建和返回。


# 3.面向对象编程


go有4个类型比较特别，看起来像是引用类型：切片；map; channel; interface

go没有构造函数概念，对象的创建通常交给一个全局创建函数完成，用 NewXXX命名，表示"构造函数"

```go

func NewRect(x, y width, height float64) *Rect {
    return &Rect{x,y,width,height}
}

```

# 4.并发编程

> 不要通过共享内存来通信，而应该通过通信来共享内存。

main函数不会等待非主 goroutine 结束。

两种常见的并发通信模型：共享数据和消息

```go
var chanName chan ElementType
ch <- value //向 channle 写入数据会导致程序阻塞，直到有其他 goroutine 从这个channel读入厨具
value := <-ch //如果 channel 之前没有数据，从 channel 中读取数据也会导致程序阻塞，直到channel中被写入数据为止

// 随机输出0 or 1
package main

import "fmt"

func main() {
	ch := make(chan int, 1)
	for {
		select {
		case ch <- 0:
		case ch <- 1:
		}
		i := <-ch
		fmt.Println(i)
	}
}
```

用 select 实现超时机制：

```go
func main() {
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(1e9)
		timeout <- true
	}()

	select {
	case <-ch:
		// read from ch
	case <-timeout:
		// 一直没有从 ch 中读取到数据，从timeout读取到了
	}

}
```

如何判断channel是否已经关闭：


```
close(ch) //close channel
x, ok :=<=ch //如果返回的 ok false，表示 ch已经被关闭了
// 使用 runtime.Gosched() 出让时间片
```

go 提供了 sync.Mutex和sync.RWmutex 锁类型

sync.Once 全局唯一操作。

sync.atomic 提供了对于一些基础数据类型的原子操作。


# 5.网络编程


## 5.1 socket编程

```go
package main
// conn, err  net.Dioal("tcp", "127.0.0.1")

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s host:port", os.Args[0])
		os.Exit(1)
	}
	service := os.Args[1]

	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)

	_, err = conn.Write([]byte("HEAD / HTTP/1.1\r\n\r\n"))
	checkError(err)

	result, err := ioutil.ReadAll(conn)
	checkError(err)

	fmt.Println(string(result))
	os.Exit(0)
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal Error:%s", err.Error())
		os.Exit(1)
	}
}
```


