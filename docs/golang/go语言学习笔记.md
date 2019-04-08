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
