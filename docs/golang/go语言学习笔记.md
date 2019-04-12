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

