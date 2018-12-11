# 2 Go 基础


定义变量:

```go
var varname type
var name1, name2, name3 type
var name type = value
var a,b,c = v1, v2, v3
a,b,c := v1, v2 ,v3
_, b = 1, 2
```

常量：

```
const name = value
const Pi float32 = 3.14
const i = 100
```

内置基础类型：

```
// Boolean
var isActive = bool  // default false
var enabled, disabled = true, false
valid := false

// 数值类型, rune(int32), int8, int16, int32, int64, byte(uint8), uint8, uint16, uint32, uint64
// float32, float64(默认)
// complex64


//字符串, go 字符串都是采用UTF8，用 双引号或者反引号括起来定义，类型是 string
var hello string
var emptyString = ""
s := "hello"  // 字符串不可变
c := []byte(s)  // 把字符串转换为 []byte 类型
c[0] = 'c'
s2 := string(c)  // 转为 string
fmt.Printf("%s\n", s2)
m := `hello
    world`   // ` 括号来的是Raw string
```

错误类型：

```
// go有一个error 类型，package.errors
err:=errors.New("emit macho swarf: elf header corrupted")
if err != nil {
   fmt.Print(err)
}
```

分组声明：

```
import (
   "fmt"
   "os"
)
const (
    i = 100
    pi = 3.14
)

var(
    i int
    pi float32
)
```


iota 枚举：

```
const (
    x = iota  //x==0
    y = iota  //y==1
)
```

设计规则：

- 大写开头的变量可以导出（共有变量）
- 大写字母开头的函数是共有函数


array, slice, map

```
// array 数组，数组长度也是类型的一部分，[3]int, [4]int 是不同类型
var arr [n]type
var arr [10]int
arr[0] = 42
arr[1] = 13
fmt.Printf("first %d\n", arr[0])

a := [3]int{1, 2, 3}
b := [10]int{1, 2, 3}  //其余元素0
c := [...]int{1,2,3}  //自动计算个数
doubleArray := [2][4]int{[4]int{1,2,3,4}, [4]int{5,6,7,8}}


// slice 动态『数组』，引用类型，指向一个底层 array。
var fslice []int  //声明同数组，只是没有长度
slice := []byte {'a','b'}
a := [10]int{1,2,3,4,5,6,7,8,9,10}
var slice []int
slice = a[:]
/*  slice 内置函数
len 获取长度
cap 获取最大容量
append 追加一个或多个元素，返回一个 和 slice 类型一样的slice
copy 从 源 slice src 中复制元素到目标 dst，返回复制的元素个数
*/


// map，字典的概念。 map[keyType]valueType
// map 无序，长度不固定，**引用类型**，len(map)返回 key 数量，map 值可以修改。
// map和其他基本类型不同，不是 thread-safe，多个 go-routine 存取用 mutex lock
var numers map[string]int
numbers := make(map[string]int)
numbers["one"] = 1
numbers["ten"] = 10
fmt.Println("one is ", numbers["three"])

rating := map[string]float32{"c": 5, "Go": 4.5}
goRaking, ok := rating["Go"]
if ok {
fmt.Println("go rating is ", goRaking)
} else {
fmt.Println("no  go rating")
}
delete(rating, "C")
```


make, new 操作：

make 只能用于 内建类型(map/slice/chanell)的内存分配, new 用于各种类型，new 返回指针。
make 返回初始化后的（非零）值。


零值: 变量未填充之前的默认值，通常是 0

```
int,int8,int32,int64 0
uint 0x0
rune 0
byte 0x0
float32, float64 0
bool false
string ""
```


流程控制:
```
// if(不用括号)

if x > 10 {
   //...
} else  {
   //...
}

if x  := computeValue(); x > 10 {  //允许声明一个变量，但是只在该条件内作用域
//...
} else {
//...
}

if x == 3 {
//...
} else if x < 3 {
//
} else {

}



// goto，还是少用不用吧


// for
sum := 0
for idx := 0; index < 10; index++ {
    sum += idx
}
sum := 1
for ; sum < 1000; {
    sum += sum
}
// 实现 while 功能
sum := 1
for sum < 1000 {
    sum += sum
}

// break/continue


// for 配合 range 读取 slice 和 map 数据
for k,v := range map {
  fmt.Println("key, value", k, v)
}


// switch
i := 6
switch i {
case 4:
	//
	fallthrough // 强制执行后边的代码
}
```


函数: func 声明

```
func funcName(input1 type1, input2 type2) (output1 type1, output2 type2) {
  //
  return value1, value2
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}


func SumAndProduct(A, B int) (add int, Multiplied int) {
	add = A + B
	Multiplied = A * B
	return
}

```

变参：

```
func myfunc(arg ...int) {
	for _, n := range arg {
		fmt.Printf("Add the number is :%d\n", n)
	}
} // arg 是一个 int 的 slice
```


传值和传指针：默认传值，无法修改传入的值。除非使用指针
```
func add(a int) int {
	a = a+1
	return a
}
// 如果需要修改值，我们可以传递指针
func add_p(a *int) int {
	*a = *a + 1
	return *a
}   // x :=

/*
传递指针使得多个函数能够操作同一个对象
传指针比较轻量级（8bytes），只是传内存地址，可以用指针传递大结构体。
GO中 channel, slice，map 实现机制类似指针，可以直接传递，而不用取地址。（如果需要修改 slice 长度让需要取地址）
*/
```


defer: 延迟语句，函数执行到最后时，这些 defer 语句会逆序执行，最后函数返回。用来处理错误，关闭资源
```
func ReadWrite() bool {
	file.Open("file")
	defer file.Close()
	if failureX {
		return false
	}
	if failureY {
		return false
	}
	return true
}
```


函数作为值、类型：go 中函数也是一种变量，可以用 type 定义，类型就是所有拥有相同的参数，相同返回值的一种类型

```
// type typeName func(input1 inputType1, intpu2 inputType2 [, ...]) (result1 resultType1 [, ...])
// 可以把这个类型的函数当做值来传递

package main

import "fmt"

type testInt func(int) bool // 声明一个函数类型
func isOdd(i int) bool {
	if i%2 == 0 {
		return false
	}
	return true
}

func filter(slice []int, f testInt) []int {
	var result []int
	for _, value := range slice {
		if f(value) {
			result = append(result, value)
		}
	}
	return result
}
func main() {
	slice := []int{1, 2, 3, 4}
	odd := filter(slice, isOdd)
	fmt.Println("Odd:", odd)

}
```


Panic/Recover:
```
Panic 是一个内建函数，可以中断原有的控制流程，进入一个令人恐慌的流程中。
当 F 函数调用 panic，F 的执行逻辑被中断，但是 F 中的延迟函数会正常执行，然后 F 返回调用它的地方。
在调用的地方，F 的行为就像调用了 panic，继续向上直到发生 panic 的 goroutine 中所有调用的函数返回，此时程序退出。
panic 可以手动或者运行时错误产生（越界数组）

 recover：内建函数，可以让进入恐慌的流程中的 goroutine 回复过来。recover 仅在延迟函数中有效。
 正常逻辑调用 recover 返回 nil。当前的 goroutine 先入恐慌可以用 recover 捕获 panic 的输入值，并且回复正常的执行。
```

main init 函数：

![](./main_init.png)


import: 导入包文件

```
// 点操作，表示你调用的时候可以省略前缀，比如 fmt.Println("hello") 可以写 Println("Hello")
import (
 . "fmt"
)

// 别名操作
import (
	f "fmt"  //别名，简化调用 f.Println
	"os"
)

// _ 操作
import (
	_ "github.com/ziutek/mymysql/godrv" //_引入该包，而不直接使用包里的函数，而是调用包里的 init 函数
)
```


Struct 类型：属性或者字段容器
```
package main

import (
	"fmt"
)

type Person struct {
	name string
	age  int
}

func Older(p1, p2 Person) (Person, int) {
	if p1.age > p2.age {
		return p1, p1.age - p2.age
	}
	return p2, p2.age - p1.age
}

func main() {
	// 初始化方式
	var P Person
	P.name = "Astaxie"
	P.age = 23
	p1 := Person{"Tom", 12}
	p2 := Person{age: 10, name: "Tom"}
	p := new(Person) // p 是个指针
	fmt.Printf("name is ", P.name)
}
```




Select: 多个 channel。select 可以监听 channel 上的数据流动。select 默认阻塞，只有当监听的 channel 中有发送或接收可以
进行时才会运行，当多个 channel 都准备好的时候随机选择一个执行。

```
package main

import "fmt"

func fibonacci(c, quit chan int) {
	x, y := 1, 1
	for {
		select {
		case c <- x:
			x, y = y, x+y
		case <-quit:
			fmt.Println("quit")
			return
    default:
        // 当监听的 channel 没有准备好的时候默认执行(select不再阻塞等待channel)

		}
	}
}

func main() {
	c := make(chan int)
	quit := make(chan int)
	go func() {
		for i := 0; i < 10; i++ {
			fmt.Println(<-c)
		}
		quit <- 0
	}()
	fibonacci(c, quit)
}
```

有时候会出现 goroutine 阻塞的情况。还可以设置 select 设置超时。


# 3 Web基础

用 go 实现一个 http server 真滴很简单，支持高并发，内部实现是一个单独的 goroutine 处理请求。


```
package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

func sayhelloName(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println(r.Form)
	fmt.Println("path", r.URL.Path)
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
	fmt.Fprintf(w, "hello") //输出到客户端

}

func main() {
	http.HandleFunc("/", sayhelloName)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
```

# 4 表单



