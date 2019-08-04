最近开始学习 golang，先挑一本入门书《The Go Programming Language》。以下是我的读书笔记，会有一些和 python
的对(tu)比(cao)，动态语言和编译型静态语言写法和思路上还是不小的差别的，不过如果熟悉 C 的话， 还是比较容易上手的。

# 1.Tutorial

第一章通过一些例子介绍了 go 语言，给出了几个直观的例子展示了 golang，还是先从 helloworld 开始:

    # go run hello.go 运行，或者 go build 编译后执行二进制文件
    package main

    import "fmt"

    func main() {
     	fmt.Println("hello world")
    }

go 通过 package（包）组织模块，一个包含多个 go 源文件，每个源文件开头用 package 声明，紧跟一堆 import 语句。
package main 比较特殊，是程序的入口，定义了单独的可执行程序，而不是 library。main 函数也比较特殊，程序从 main 开始执行
go 有个 gofmt 工具可以用来格式化代码（类似 autopep8，笔者用的 vim-go 写完保存会自动执行），消除不同人对代码格式的撕逼大战。

下一个例子我们多重方式实现一个打印命令行参数的代码：

    package main

    import (
    	"fmt"
    	"os"
    	"strings"
    )

    func main() {
    	var s, sep string
    	for i := 1; i < len(os.Args); i++ {
    		s += sep + os.Args[i]
    		sep = " "
    	}
    	fmt.Println(s)
    }

    func main2() {
    	s, sep := "", ""
    	for _, arg := range os.Args[1:] {
    		s += sep + arg
    		sep = ""
    	}
    }

    func main3() {
    	fmt.Println(strings.Join(os.Args[1:], " "))
    }

接着是消除从标准输入读取的重复行的代码(like uniq command)：

    package main

    import (
    	"bufio"
    	"fmt"
    	"os"
    )

    func main() {
    	counts := make(map[string]int) // like python dict {string: int}
    	input := bufio.NewScanner(os.Stdin)
    	for input.Scan() { //until EOF (ctrl + d)
    		counts[input.Text()]++
    	}
    	for line, n := range counts {
    		if n > 1 {
    			fmt.Printf("%d\t%s\n", n, line) // 格式化函数一般以 f 结尾
    		}
    	}

    }


    // 同样的功能，同时可以处理从文件读取的行
    package main

    import (
    	"bufio"
    	"fmt"
    	"os"
    )

    func main() {
    	counts := make(map[string]int)
    	files := os.Args[1:]
    	if len(files) == 0 {
    		countLines(os.Stdin, counts)
    	} else {
    		for _, arg := range files {
    			f, err := os.Open(arg)
    			if err != nil {
    				fmt.Fprintf(os.Stderr, "dup2: %v\n", err)
    				continue
    			}
    			countLines(f, counts)
    			f.Close()
    		}
    	}
    	for line, n := range counts {
    		if n > 1 {
    			fmt.Printf("%d\t%s\n", n, line)
    		}
    	}
    }

    func countLines(f *os.File, counts map[string]int) {
    	input := bufio.NewScanner(f)
    	for input.Scan() {
    		counts[input.Text()]++
    	}
    }

    // 还是一样的功能，我们让程序可以处理整个文件，而不是一行一行读取，引入 io/ioutil 包
    package main

    import (
    	"fmt"
    	"io/ioutil"
    	"os"
    	"strings"
    )

    func main() {
    	counts := make(map[string]int)
    	for _, filename := range os.Args[1:] {
    		data, err := ioutil.ReadFile(filename) // return byte slice
    		if err != nil {
    			fmt.Fprintf(os.Stderr, "dup3: %v\n", err)
    			continue
    		}
    		for _, line := range strings.Split(string(data), "\n") {
    			// 把文件拆成一行一行计数
    			counts[line]++
    		}
    	}
    	for line, n := range counts {
    		if n > 1 { // 打印重复行
    			fmt.Printf("%d\t%s\n", n, line)
    		}
    	}
    }

接下来是一个爬虫的例子，当然看起来没有用 python requests 那么优雅，有个类似的 grequests 更精简：

    package main

    import (
    	"fmt"
    	"io/ioutil"
    	"net/http"
    	"os"
    )

    // go run main.go http://baidu.com
    func main() {
    	for _, url := range os.Args[1:] {
    		resp, err := http.Get(url)
    		if err != nil {
    			fmt.Fprintf(os.Stderr, "fetch :%v\n", err)
    		}

    		b, err := ioutil.ReadAll(resp.Body)
    		resp.Body.Close()
    		if err != nil {
    			fmt.Fprintf(os.Stderr, "fetch: reading %s: %v\n", url, err)
    			os.Exit(1)
    		}
    		fmt.Printf("%s", b)
    	}
    }

然后我们并发请求数据，并发可是 golang 的卖点之一:

    package main

    import (
    	"fmt"
    	"io"
    	"io/ioutil"
    	"net/http"
    	"os"
    	"time"
    )

    // go run main.go  https://baidu.com https://zhihu.com https://douban.com
    func main() {
    	start := time.Now()
    	ch := make(chan string)
    	for _, url := range os.Args[1:] {
    		go fetch(url, ch) // 启动一个 goroutine
    	}
    	for range os.Args[1:] {
    		//当一个 goroutine 从 channel 发送或者接收值的时候，会被 block 直到另一个 goroutine 响应
    		fmt.Println(<-ch) // receive from channel ch
    	}
    	fmt.Printf("%.2fs elapsed\n", time.Since(start).Seconds())
    }

    func fetch(url string, ch chan<- string) {
    	start := time.Now()
    	resp, err := http.Get(url)
    	if err != nil {
    		ch <- fmt.Sprint(err) // send to channel ch
    		return
    	}
    	nbytes, err := io.Copy(ioutil.Discard, resp.Body)
    	resp.Body.Close()
    	if err != nil { // 不得不吐槽下 go 的错误处理，写异常习惯了，到处都是 err 真烦人
    		ch <- fmt.Sprintf("while reanding %s:%v", url, err) // 向channel 发送数据(ch <- expression)
    		return
    	}
    	secs := time.Since(start).Seconds()
    	ch <- fmt.Sprintf("%.2fs %7d %s", secs, nbytes, url)
    }

这一个例子是写一个 web echo server，看起来比较简单：

    package main

    import (
    	"fmt"
    	"log"
    	"net/http"
    )

    // echo server
    func main() {
    	http.HandleFunc("/", handler) // each request calls handler
    	log.Fatal(http.ListenAndServe("localhost:8000", nil))
    }

    func handler(w http.ResponseWriter, r *http.Request) {
    	fmt.Fprintf(w, "URL.Path = %q\n", r.URL.Path)
    }

然后我们添点料，比如实现个请求计数的功能：

    package main

    import (
    	"fmt"
    	"log"
    	"net/http"
    	"sync"
    )

    var mu sync.Mutex
    var count int

    func main() {
    	http.HandleFunc("/", handler)
    	http.HandleFunc("/count", counter)
    	log.Fatal(http.ListenAndServe("localhost:8000", nil))
    }

    func handler(w http.ResponseWriter, r *http.Request) {
    	mu.Lock() // 限定最多一个 goroutine 访问 count (每个请求开一个 goroutine 处理)
    	count++
    	mu.Unlock()
    	fmt.Fprintf(w, "URL.Path = %q\n", r.URL.Path)
    }

    func counter(w http.ResponseWriter, r *http.Request) {
    	mu.Lock()
    	fmt.Fprintf(w, "count %d\n", count)
    	mu.Unlock()
    }

最后再几个 golang 的语法特性：

-   控制流 switch

```go
switch coinflip() {
case "heads":    // case 还支持简单的语句 ()。case可以支持 string
    heads++
case "tails":
    tails++
default:
    fmt.Println("landed on edge!")
}
```

-   Named Types:

```go
//定义一个 Point 类型
type Point struct {
    X, Y int
}
var p Point
```

-   Pointers(指针)：和 C 类似，go 中也实现了指针
-   Methods and interfaces（方法和接口）: 方法是关联到一个命名类型的函数。接口是一种把不同类型同等对待的抽象类型
-   Packages(包): 通过包组织程序
-   Comments(注释)：和 C 一样的注释,  `// or /* XXXX */`

# 2. Program Structure

表达式+控制流 -> 语句 -> 函数 -> 源文件 -> 包

## 2.1 Names

go 定义了几十个关键字，不能用来给变量命名，go 使用一般使用骆驼命名法(HTTP等缩略词除外)。需要注意的是只有大骆驼命名
"fmt.Fprintf" 这种是可以被其他包引入使用的。go 通过大小写表示是否能够引入，没有访问控制声明符。

## 2.2 Declarations

var, const ,type, func 声明

    func fToC(f float64) float64 {
    	return (f - 32) * 5 / 9
    }

## 2.3 Variables

定义变量：如果省略了 type 将会用每个类型的初始值初始化赋值（类似 java）

    var name type = expression
    // 可以一次性定义多个
    var i, j, k int   //i,j,k == 0
    var b, f, s  = true, 2.3. "four"    // 同样有类似解包的操作

    // 接下来是 short variable declarations， 短赋值，`:=` 是 声明，而 `=` 是赋值
    freq := rand.Float64() * 3.0
    t := 0.0

    // multiple variables declared
    i, j = 0, 1
    i, j = j, i    // swap, 这个和python一样支持

    f, err := os.Open()
    f, err := os.Close()     // wrong,  a short variable declaration must declare at least one new variable

指针: 如果学过 c，这里的指针很类似，表示一个变量的地址

    x := 1
    p := &x    // p, of type *int, points to x
    fm.Println(*p) // "1"
    *p = 2
    fmt.Println(x) // "2"

    var x, y int
    fmt.Println(&x == &x, &x == &y, &x == nil)
    // "true false false"，指针可以比较，当指向相同的值的时候相等。但是为了简化，go指针不能做运算

    // 函数返回一个局部变量的指针也是安全的
    var p = f()
    func f() *int {
        v:=1
        return &v
    }

    // 可以传入指针给函数改变其所指向的值
    func incr(p *int) int {
    	*p++
    	return *p
    }
    v :=1
    incr(&v)
    fmt.Println(incr(&v))

new 函数：创建一个变量的另一种方式是使用内置函数 new，new(T)创建一个未命名T
类型的变量，用初始值初始化，然后返回其地址(\*T)

    func test() {
    	p := new(int)
    	fmt.Println(*p)
    	*p = 2
    	fmt.Println(*p)    //可以不通过变量名就访问它（指针的好处之一）
    }

    //两种等价写法
    func newInt() *int {
    	return new(int)
    }
    func newInt2() *int {
    	var dummy int
    	return &dummy
    }

变量生命周期(lifetime):
package-level变量在整个程序执行过程中都存在。局部变量生存周期是动态的，每次一个实例在生命语句执行的时候被创建，直到不可访问的时候被回收。
虽然 go 有自己的垃圾回收机制，但是代码里尽量让变量的生存周期更短

## 2.4 赋值

相较于 python，go 支持自增操作符。和 py相同的是同样支持 tuple 的解包赋值，演示一些例子：

    	a, b, c = 1, 2, 3
    	x, y = y, x    // swap x and y like python

    	v, ok = m[key] // map lookup
    	v, ok = x.(T)  // type assertion
    	v, ok = <-ch   // channel receive
    	_, ok = x.(T)  // check type but discard result

    	medals := []string{"gold", "silver", "bronze"}
    	// 隐式赋值等价于
    	medals[0] = "gold"
    	medals[1] = "silver"
    	medals[2] = "bronze"

可赋值性：For the types we’ve discussed so far, the rules are simple: the types must exactly match, and nil may be assigned to any variable of interface or reference type.

## 2.5 Type Declarations

一个变量的类型决定了它的属性和能做哪些操作，比如它的大小（占多少 bit），内部表示，能做哪些操作，关联了哪些方法。
`type name underlying-type`

    //看一个摄氏温度和华氏温度转换的例子
    type Celsius float64 //  自定义类型
    type Fahrenheit float64

    const (
    	AbsoluteZeroC Celsius = -273.15
    	FreezingC     Celsius = 0
    	BoilingC      Celsius = 100
    )

    func CToF(c Celsius) Fahrenheit { return Fahrenheit(c*95 + 32) }
    func FToC(f Fahrenheit) Celsius { return Celsius((f - 32) * 5 / 9) }

如果 underlying type 一样可以类型强转, T(x)

## 2.6 Packages and Files

每个包都声明了名字空间，当有名字冲突的时候，需要显式通过包名调用。大写开头的标识符才可以被导出。把上面的例子改成包，

    // gopl.io/ch2/tempconv
    // tempconv.go
    package tempconv

    import (
    	"fmt"
    )

    //看一个摄氏温度和华氏温度转换的例子
    type Celsius float64 //  自定义类型
    type Fahrenheit float64

    const (
    	AbsoluteZeroC Celsius = -273.15
    	FreezingC     Celsius = 0
    	BoilingC      Celsius = 100
    )

    // 可以给自定义的类型来编写方法
    func (c Celsius) String() string    { return fmt.Sprintf("%g°C", c) }
    func (f Fahrenheit) String() string { return fmt.Sprintf("%g°F", c) }

    // conv.go
    package tempconv
    func CToF(c Celsius) Fahrenheit { return Fahrenheit(c*95 + 32) }
    func FToC(f Fahrenheit) Celsius { return Celsius((f - 32) * 5 / 9) }

go程序里每个包都通过一个唯一的 import path 标识，比如 "gopl.io/ch2/tempconv"
包初始化的顺序：以包级别的变量开始按照声明顺序初始化，但是被依赖的值先初始化

    var a=b+c // 3
    var b=f() // 2
    var c=1//1
    func f() int {return c+1}

如果一个包有多个 go 源文件，会被给编译器的顺序执行初始化，go tools 通过把文件名排序后给交给编译器
可以在文件里定义 init() 函数，他们不能被引用，程序执行的时候会按照被声明顺序自动运行(比如初始化一个查找表)

    //gopl.io/ch2/popcount
    package popcount

    var pc [256]byte

    func init() {
    	for i := range pc {
    		pc[i] = pc[i/2] + byte(i&1)
    	}
    }

    func PopCount(x uint64) int {
    	return int(pc[byte(x>>(0*8))] +
    		pc[byte(x>>(1*8))] +
    		pc[byte(x>>(2*8))] +
    		pc[byte(x>>(3*8))] +
    		pc[byte(x>>(4*8))] +
    		pc[byte(x>>(5*8))] +
    		pc[byte(x>>(6*8))] +
    		pc[byte(x>>(7*8))])
    }

## 2.7 Scope

声明绑定了名字和程序实体，比如函数或者变量，一个声明的作用域表示这个声明的名字在哪些代码块起作用。不要和生命周期（lifetime）
混淆，声明的作用域是一段程序片段，编译时属性，生命周期指的是运行期间可以被程序其他部分所引用的持续时间，运行时属性。
golang 中通过块（block）来圈定作用域。名字查找遵循『就近』原则，内部声明会屏蔽外部声明（最好不要重名，理解各种重名覆盖问题比较费劲）

# 3. Basic Data Types

go的数据类型分成4类：

-   基础类型: numbers, strings, booleans
-   聚合类型: arrays, structs,
-   引用类型: pointers, slices, maps, functions, channels
-   接口类型: interface types

## 3.1 Integers

int8, int16, int32, int64, uint8, uint16, uint32, uint64
rune &lt;=> int32 , byte &lt;=> uint8

- 有符号数： `-2**(n-1) to 2**(n-1)-1`
- 无符号数: `0 to 2**n-1`

尤其注意不同类型之间数字强转可能会有精度损失

## 3.2 Floating-Point Numbers

-   float32: 1.4e-45 to 3.4e38
-   float64: 1.8e308 to 4.9e-324
-   math package

## 3.3 Complex Numbers

-   complex64
-   complex128
-   math/cmplx package

    `var x complex128 = complex(1,2)`

## 3.4 booleans

-   true
-   false
-   短路求值特性，比如 a() || b()，如果 a() 返回true，b()不会执行

## 3.5 Strings

不可变字节序列，可以包含任何数据，通常是人类可读的。文本 string 可以方便地解释成 utf8编码的 unicode 码(runes)

-   len: len(string) return number of bytes(not runes), `0 <= i < len(s)`
-   切片: s[i:j] yield a new substring
-   字符串字面量: "Hello，世界" (双引号) 。go 源文件总是用 utf8编码，go 文本字符串被解释成 utf8。
-   raw string literal: 反引号包含的字符串，转义符不会被处理，经常用来写正则
-   unicode: go 术语里叫做 rune(int32)
-   utf8: 一种把 unicode 字节码转成字节序列的一种变长编码方案(utf8由 Ken Thompson and Rob Pike发明，同时还是 go 的创造者)
-   Go’s range loop, when applied to a string, performs UTF-8 decoding implicitly
-   A \[]rune conversion applied to a UTF-8-encoded string returns the sequence of Unicode code points that the string encodes

```
package main

import (
    "fmt"
    "unicode/utf8"
)

func main() {
    s := "hello, 世界"
    fmt.Println(len(s))                    // 13
    fmt.Println(utf8.RuneCountInString(s)) // 9
}
```

几个用来处理字符串的包：

-   bytes:  操作 slices of bytes, type \[]byte
-   strings: searching, replacing, comparing, trimming, splitting, joining
-   strconv: boolean, integer, floating-point values 和 他们的 string 表示形式来回转；
-   unicode: IsDigit, IsLetter, IsUpper, IsLower识别 runes

## 3.6 Constants

const 表达式的值在编译器确定，无法改变。const 值可以是boolean, string or number
The const generator

iota: 定义从0 开始的递增枚举

    const (
    	Sunday Weekday = iota
    	Monday
    	Tuesday
    	Wednesday
    	Thrusday
    	Friday
    	Saturday
    )

# 4. Composite Types

本章讨论聚合类型: arrays, slices, maps, structs

## 4.1 Arrays

定长的包含0个或多个相同类型元素的序列，因为不可变长所以一般 slices 更常用。
注意数组的长度也是类型的一部分， [3]int 和 [4]int 是不同类型，相同长度的可以比较。

    	var a [3]int // array of 3 integers，第一次看这种写法略奇怪，理解为 [3]个int 就好多啦
    	fmt.Println(a[0])
    	for i, v := range a {
    		fmt.Printf("%d %d\n", i, v)
    	}

    	var q [3]int = [3]int{1, 2, 3} //初始化
    	p := [...]int{1, 2, 3}         // 省略号代表长度由右边的长度决定, []int{1,2,3} slice

    	// 数组还支持用 下标和值 初始化
    	type Currency int
    	const (
    		USD currency = iota
    		EUR
    		GBP
    		RMB
    	)
    	symbol:= [...]string{USD: "$", EUR: "€", GBP: "₤", RMB: "￥"}
    	fmt.Println(RMB, symbol[RMB]) // "3 ￥"

数组传值，拷贝值，如果需要传递引用可以通过传递指针

    func zero(ptr *[32]byte) {  // 注意这里的数字长度也是类型的一部分
    	for i := range ptr {
    		ptr[i] = 0
    	}
    }
    func zero(ptr *[32]byte) {
    	*ptr = [32]byte{}
    }

## 4.2 Slices

包含相同类型元素的可变长序列，写法 \[]T，类似于没有 size 的数组，由三部分组成：

-   a pointer: 指向数组第一个元素
-   a length: number of slice elements  (len 函数)
-   a capacity: the number of elements between the start of the slice and the end of the underlying array (cap函数)

slice operator s[i:j], where 0 ≤ i ≤ j ≤ cap(s), 创建一个新的 slice

注意，slice 包含了一个数组的指针，作为函数参数传递的时候允许函数修改数组。(引用类型)

    package main

    import "fmt"

    func reverse(s []int) {
    	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
    		s[i], s[j] = s[j], s[i]
    	}
    }

    func main() {
    	a := [...]int{0, 1, 2, 3, 4}
    	reverse(a[:])
    	fmt.Println(a)

    	s := []int{0, 1, 2, 3, 4, 5} //注意 slice 初始化和数组的区别，没有 size
    	reverse(s[:2])
    	reverse(s[2:])
    	reverse(s)
    	fmt.Println(s)
    }

**并且 slice 不像数组一样能直接比较**，基于两个原因：

-   slice 可以包含自己，深度比较无论是复杂度还是实现效率都不好
-   不适合作为 map 的 key。对于引用类型（pointers and channels）， == 一般比较引用实体是否是同一个。但是这回导致 slice
    和数组的 == 一个是 shallow equality，一个是 deep equality，为了防止歧义最安全的方式就是干脆禁止 slice 比较。唯一合法的
    slice 比较就是判断 `if slice == nil {}`

slice 的初始化值（zero value）是 nil，nil slice没有隐含指向的数组，其 length==capacity==0。但是也有包含 nil
的slice，我们看几个比较(我感觉这里 python 的 None 用 is 判断比较优雅)：

    var s []int    // len(s) == 0, s == nil，空切片
    s = make ([]int, 0)  // or slice := []int{}   // 都是空切片
    s = nil        // len(s) == 0, s == nil
    s = []int(nil) // len(s) == 0, s == nil
    s = []int{}    // len(s) == 0, s != nil

所以如果只是判断一个 slice 是否为空，应该用 len(s) == 0，而不是 s == nil。内置的 make 函数用来创建 slice

    	make([]T, len)
    	make([]T, len, cap)

向 slice 追加元素用内置的 append 函数。append 会返回一个包含结果的新切片（容量有可能改变）
如果切片的底层数组没有足够的可用容量，append 函数会创建一个新的底层数组，将被引用的现有数组复制到新的数组里，
再追加新的值。

    var runes []rune
    for _, r := range "hello,世界" {
    	runes = append(runes, r)
    } // or just use :  []rune("hello,世界")

    // 我们实现一个 appendInt 看下 append 的工作原理
    func appendInt(x []int, y int) []int {
    	var z []int
    	zlen := len(x) + 1
    	if zlen <= cap(x) { // 容量够用
    		z = x[:zlen]
    	} else {
    		// 如果容量不够 append 的了，将容量扩大两倍(python 的 list 和c 艹 vector 类似扩充策略)
    		zcap := zlen
    		if zcap < 2*len(x) {
    			zcap = 2 * len(x)
    		}
    		z = make([]int, zlen, zcap)
    		copy(z, x) // build int funcion `func copy(dst, src []Type) int`
    	}
    	z[len(x)] = y
    	return z
    }

任何改变 长度或者容量的操作都会更新 slice，虽然 slice 指向的数组是间接的，但是指针、长度和容量不是，slice
不是纯引用类型，更像是如下结构：

    type IntSlice struct {
    	ptr      *int
    	len, cap int
    }

append 是个可变参数的函数，可以接收多个参数， ... 运算符可以追加一个切片的所有元素到另一个切片里。

	s1 := []int{1, 2}
	s2 := []int{3, 4}
	fmt.Printf("%v\n", append(s1, s2...))
    for index, value := range s1 {
		// value 是元素的副本，而不是直接返回该元素的引用
		fmt.Print(index, value, "\n")
	}
    for index := 2; index < len(s); index++ {
		fmt.Print(s[index])
	}

    // 多维切片
	slice := [][]int{{10}, {100, 200}}
	slice[0] = append(slice[0], 20)


可以使用省略号来让函数接受任意多个 final arguments：

    func appendInt(x []int, y ...int) []int{
    	var z[]int
    	zlen:=len(x) + len(y)
    	// ... expand z to at least zlen...
    	copy(z[len(x):], y)
    	return z
    }

切片作为函数传递的代价很小，不涉及到底层数组复制，只会复制切片自己的结构(24个字节)



## 4.3 Maps

无序键值对，go 里 map 是 哈希表的引用。要求所有 key 类型相同，所有的 value 类型相同，key
必须是可比较==的。不过用法上和 python 的 map 还是比较类似的。
切片、函数和包含切片的结构类型由于具有引用语句，不能作为映射的键(key)。
作为参数的传递的 map 如果修改了 map 会对所有的引用可见，不会复制底层的数据结构。

    var colors map[string][string] // 声明一个未初始化的映射来创建一个值为 nil 的映射， nil 映射不能存储键值对

    package main

    import (
    	"fmt"
    	"sort"
    )

    func main() {
    	ages := make(map[string]int) //创建 map
    	ages["alice"] = 31
    	ages["charlie"] = 34

    	ages2 := map[string]int{ //创建并初始化 map
    		"alice":   31,
    		"charlie": 34,
    	}
    	fmt.Println(ages2)

    	ages["bob"] = ages["bob"] + 1 // bob不存在的 key 默认是0，注意不像 py 访问不存在的 key 会 KeyError
    	delete(ages, "alice")         //删除 key
    	fmt.Println(ages)

    	// 遍历无序的 kv，注意每次返回顺序是不保证一致的
    	for name, age := range ages {
    		fmt.Printf("%s\t%d\n", name, age)
    	}

    	// 有序遍历，实现有序遍历需要我们把名字先有序排列
    	var names []string
    	for name := range ages { // 省略 age 会只遍历 key
    		names = append(names, name)
    	}
    	sort.Strings(names)
    	for _, name := range names {
    		fmt.Printf("%s\t%d\n", name, ages[name])
    	}

    	// map 对于不存在的 key 返回类型默认值，要想判断是否存在 key 得这么写
    	if age, ok := ages["noexists"]; !ok { /* */
    		fmt.Printf("not a key in this map, age is %s", age)
    	}
    }

map 同样不能直接比较，需要自己写比较函数：

    func equal(x, y map[string]int) bool {
    	if len(x) != len(y) {
    		return false
    	}
    	for k, xv := range x {
    		if yv, ok := y[k]; !ok || xv != yv {
    			return false
    		}
    	}
    	return true
    }

go 里没提供 set 类型，我们可以通过 map[string]bool 之类的来模拟。key 不支持 slice 类型，但是如果想用的话可以用转换函数转成 string 后来作为
map 的 key。

    package main

    import (
    	"bufio"
    	"fmt"
    	"io"
    	"os"
    	"unicode"
    	"unicode/utf8"
    )

    func main() {
    	counts := make(map[rune]int) // counts of unicode chars
    	var utflen [utf8.UTFMax + 1]int
    	invalid := 0

    	in := bufio.NewReader(os.Stdin)
    	for {
    		r, n, err := in.ReadRune() // rune, bytes, error
    		if err == io.EOF {
    			break
    		}
    		if err != nil {
    			fmt.Fprintf(os.Stderr, "charcount: %v\n", err)
    		}
    		if r == unicode.ReplacementChar && n == 1 {
    			invalid++
    			continue
    		}
    		counts[r]++
    		utflen[n]++
    	}
    	fmt.Printf("rune\tcount\n")
    	for c, n := range counts {
    		fmt.Printf("%q\t%d\n", c, n)
    	}
    	fmt.Print("\nlen\tcount\n")
    	for i, n := range utflen {
    		if i > 0 {
    			fmt.Printf("%d\t%d\n", i, n)
    		}
    	}
    	if invalid > 0 {
    		fmt.Printf("\n%d invalid UTF-8 characters\n", invalid)
    	}

    }

## 4.4 Structs

结构体是聚合数据类型：结合多个任意类型的数据组成单个实体，每个值都叫做一个
field，可以被当做一个单元执行拷贝，传递给函数或者返回，存到数组等。

    package main

    import "time"

    func main() {
    }

    type Employee struct {
    	ID        int
    	Name      string    // 首字母大写 is exported
    	Address   string
    	DoB       time.Time
    	Position  string
    	Salary    int
    	ManagerID int
    }

    func main() {
    	var dilbert Employee
    	dilbert.Salary -= 5000 // demoted, for writing too few lines of code

    	position := &dilbert.Position
    	*position = "Senior " + *position // promoted, for outsourcing to Elbonia

    	var employeeOfTheMonth *Employee = &dilbert
    	employeeOfTheMonth.Position += " (proactive team player)"

    }

    // NOTE: 一个命名的结构体 S 无法自己包含自己，但是却可以包含指针类型 type *S。方便实现链表、树等结构
    type tree struct {
    	value       int
    	left, right *tree
    }

    // sort values in place
    func Sort(values []int) {
    	var root *tree
    	for _, v := range values {
    		root = add(root, v)
    	}
    	appendValues(values[0:], root)
    }

    func appendValues(values []int, t *tree) []int {
    	if t != nil {
    		values = appendValues(values, t.left)
    		values = append(values, t.value)
    		values = appendValues(values, t.right)
    	}
    	return values
    }

    func add(t *tree, value int) *tree {
    	if t == nil {
    		t = new(tree)
    		t.value = value
    		return t
    	}
    	if value < t.value {
    		t.left = add(t.left, value)
    	} else {
    		t.right = add(t.right, value)
    	}
    	return t
    }

结构体字面量：

    package main

    import "fmt"

    type Point struct{ X, Y int }

    func main() {

    	p := Point{1, 2} //不指定 field 初始化需要遵守顺序
    	p = Point{X: 1, Y: 2}
    	fmt.Printf("%v", Scale(p, 2))
    	fmt.Printf("%v", ScaleByPointer(&p, 2))
    }

    // 结构体作为函数参数
    func Scale(p Point, factor int) Point {
    	return Point{p.X * factor, p.Y * factor}
    }

    // 结构体作为函数参数
    func ScaleByPointer(p *Point, factor int) Point {
    	return Point{p.X * factor, p.Y * factor}
    }
    /*
    ppp := new(Point)
    *pp = Point{1, 2}
    //等价写法
    pp := &Point{1, 2}
    */

可比较性：如果结构体的所有 field 都是可以比较的，struct 本身也可以比较。

golang的 struct 允许嵌套：

    package main

    import "fmt"

    type Point struct {
    	X, Y int
    }
    type Circle struct {
    	Center Point
    	Radius int
    }
    type Wheel struct {
    	Circle Circle
    	Spokes int
    }

    func main() {

    	var w Wheel
    	w.Circle.Center.X = 8
    	w.Circle.Center.Y = 8
    	w.Circle.Radius = 5
    	w.Spokes = 20

    }

之前这种连续赋值写起来很繁琐，可以用匿名 field 简化(其实不喜欢这种方式，虽然多写一些麻烦了点，但是很直观)：

    package main

    func main() {

    	var w Wheel
    	w.X = 8 // equivalent to w.Circle.Point.X = 8
    	w.Y = 8 // equivalent to w.Circle.Point.Y = 8
    }

    //go允许匿名 field，但是它的类型必须是命名类型或者其指针
    type Point struct {
    	X, Y int
    }
    type Circle struct {
    	Point
    	Radius int
    }

    type Wheel struct {
    	Point
    	Spokes int
    }

## 4.5 JSON

JavaScript Object Notation (JSON) is a standard notation for sending and receiving structured information.
JSON is an encoding of JavaScript values—strings, numbers, booleans, arrays, and objects—as Unicode text
Go data structure like movies to JSON is called marshaling. Marshaling is done by json.Marshal:

    package main

    import (
    	"encoding/json"
    	"fmt"
    	"log"
    )
    // 大写开头的 field 才会被 marshaled
    type Movie struct {
    	Title  string
    	Year   int  `json:"released"`
    	Color  bool `json:"color,omitempty"`     // json tag 可以用来重命名字段
    	Actors []string
    }

    func main() {
    	var movies = []Movie{
    		{Title: "Casablanca", Year: 1942, Color: false, Actors: []string{"Humphrey Bogart", "Ingrid Bergman"}},
    		{Title: "Cool Hand Luke", Year: 1967, Color: true, Actors: []string{"Paul Newman"}},
    		{Title: "Bullitt", Year: 1968, Color: true, Actors: []string{"Steve McQueen", "Jacqueline Bisset"}},
    	}
    	// data, err := json.Marshal(movies)
    	data, err := json.MarshalIndent(movies, "", "    ")
    	if err != nil {
    		log.Fatalf("JSON marshaling failed:%s", err)
    	}
    	fmt.Printf("%s\n", data)

    	var titles []struct{ Title string }
    	if err := json.Unmarshal(data, &titles); err != nil {    // json.Unmarshal 反序列化
    		log.Fatalf("JSON unmarshaling failed: %s", err)
    	}
    	fmt.Println(titles)
    }

注意到 field 后边的 `json:"released"` 叫做 field tags，是在编译器就绑定到 field 上的元信息，可以控制序列化/反序列化的行为。
field tag 第一部分指定了 field 的 json 名称（比如从骆驼命名改成下划线命名），第二个可选的选项（omitempty），指定了当如果该 field 是零值的时候不会输出该字段.

## 4.6 Text and HTML Templates

text/template and html/template packages

# 5. Functions

## 5.1 Function Declarations

    func name(parameter-list) (result-list) {
    	body
    }

定义演示：

    func f(i, j, k int, s, t string)                { /* ... */ }
    func f(i int, j int, k int, s string, t string) { /* ... */ }
    func add(x int, y int) int                      { return x + y }
    func sub(x, y int) (z int)                      { z = x - y; return }
    func first(x int, _ int) int                    { return x }
    func zero(int, int) int                         { return 0 }

    // 函数类型通常叫做函数签名(signature)
    fmt.Printf("%T\n", add)
    fmt.Printf("%T\n", sub)
    fmt.Printf("%T\n", first) // "func(int, int) int"
    fmt.Printf("%T\n", zero)  // "func(int, int) int"

注意 go 里参数传的是值的拷贝，如果是不可变类型不会影响原参数，但是如果是引用类型比如 pointer, slice, map, function or
channel ，函数是可以修改参数指向的内容的。

## 5.2 Recursion

函数可以直接或者间接调用自己，称为递归

    func fib(n int) int {
    	if n <= 1 {
    		return 1
    	}
    	return fib(n-1) + fib(n-2)
    }

## 5.3 Multiple Return values

go和 python 一样可以返回多个值：

    package main

    import (
    	"fmt"
    	"net/http"
    	"os"

    	"golang.org/x/net/html"
    )

    func main() {
    	for _, url := range os.Args[1:] {
    		links, err := findLinks(url)
    		if err != nil {
    			fmt.Fprintf(os.Stderr, "findlinks2: %v\n", err)
    			continue
    		}
    		for _, link := range links {
    			fmt.Println(link)
    		}
    	}
    }

    // findLinks performs an HTTP GET request for url, parses the
    // response as HTML, and extracts and returns the links.

    func findLinks(url string) ([]string, error) {
    	resp, err := http.Get(url)
    	if err != nil {
    		return nil, err
    	}
    	if resp.StatusCode != http.StatusOK {
    		resp.Body.Close() //必须显示 close，垃圾回收不会回收系统资源比如打开的文件、网络连接等
    		return nil, fmt.Errorf("getting %s: %s", url, resp.Status)
    	}
    	doc, err := html.Parse(resp.Body)
    	resp.Body.Close()
    	if err != nil {
    		return nil, fmt.Errorf("parsing %s as HTML: %v", url, err)
    	}
    	return visit(nil, doc), nil
    }

命名返回值：当函数返回多个同类型的值的时候，挑选好名字可以使返回结构更有意义

    func Size(rect image.Rectangle) (width, height int)
    func Split(path string) (dir, file string)
    func HourMinSec(t time.Time) (hour, minute, second int)

bare return: 如果函数命名了返回值，return 后的返回结果可以省略

    func CountWordsAndImages(url string) (words, images int, err error) {
    	resp, err := http.Get(url)
    	if err != nil {
    		return // 等价于 return words, images, err
    	}
    	doc, err := html.Parse(resp.Body)
    	resp.Body.Close()
    	if err != nil {
    		err = fmt.Errorf("parsing HTML: %s", err)
    		return // 等价于 return words, images, err
    	}
    	words, images = countWordsAndImages(doc)
    	return // 等价于 return words, images, err
    }

## 5.4 Errors

有些函数总会执行成功，比如 strings.Contains and strconv.FormatBool，但是有些无法预料的场景比如内存用光了就很难恢复了。
其他函数只要先验条件是满足的总会执行成功，比如 time.Date，只要参数正确总会构造出 time.Time，除非最后一个 timezone 参数是
nil，这是就会产生 panic，panic 是调用代码的时候产生 bug 的明确标志，在良好书写的程序中是不该出现的。
对于其他一些编写良好的函数，依然无法保证程序成功执行，因为依赖程序员的控制。任何涉及到 IO
的程序都需要处理出错的可能。错误对于重要的包 API 或应用程序接口是非常重要的一部分，而且应该是预期的几个行为之一。

如果错误是期望行为通常我们多返回一个参数，比如如果只有一种错误，我们多返回一个 ok 标志，例如一个缓存查找的方法

    value, ok := cache.Lookup(key)
    if !ok {
    	// key not exists
    }

如果是像处理 IO 这种会有很多类型错误，通常多返回一个error 用来解释出错原因。内置类型  error 是接口类型，是 nil 或者
non-nil，nil 表示成功执行否则失败。non-nil会有出错信息通过调用Error方法或者 fmt.Println(err) or
fmt.Printf("%v",err)打印出来。在处理错误之前，如果需要调用者明确处理一些未完成资源，需要明确用文档表示出来。
go 没有想其他很多语言一样用异常处理错误，因为异常经常样以一种用户无法理解的调用栈形式暴露给一脸懵逼的用户，
缺少明智的上下文信息告诉用户到底什么出错了。而 go 用通常的控制流  if, return 等来响应
errors，这种方式需要付出更多精力来应对错误处理逻辑。

**go 的5种常见错误处理策略**：

-   1.propagate the error：传播错误使得被调用者的错误成为调用者错误。通常直接返回或者构造新的错误信息后返回

    	resp, err := http.Get(url)
    	if err != nil {
    		return nil, err    // just return
    	}
    	// construct new error message includes both: 填充有用信息之后再 return ，需要注意出错信息的结构和内容
    	doc, err := html.Parse(resp.Body)
    	resp.Body.Close()
    	if err != nil {
    		return nil, fmt.Errorf("parsing %s as HTML: %v", url, err)  // 附加上了 url 信息
    	}

-   2.对于短暂或者非预期的错误，可以尝试采用一定策略重试。最常见的就是爬虫里的重试策略。

```
func WaitForServer(url string) error {
    const timeout = 1 * time.Minute
    deadline := time.Now().Add(timeout)
    for tries := 0; time.Now().Before(deadline); tries++ {
	    _, err := http.Head(url)
	    if err == nil {
		    return nil // success
	    }
	    log.Printf("server not responding (%s); retrying...", err)
	    time.Sleep(time.Second << uint(tries)) // exponential back-off
    }
    return fmt.Errorf("server %s failed to respond after %s", url, timeout)
}
```

-   3.如果无法继续处理，调用者可以打印 error 然后停止执行，通常是 main
    包的保留行为。库函数通常应该传递错误给调用者，除非遇到了内部不一致性，也就是有 bug。

```
// (In function main.)
if err := WaitForServer(url); err != nil {
    fmt.Fprintf(os.Stderr, "Site is down: %v\n", err)
    os.Exit(1)
}
// 或者更方便完成一样的功能，调用 log.Fatalf
if err := WaitForServer(url); err != nil {
    log.Fatalf("Site is down: %v\n", err)
}
```

-   4.有些情况直接打印 error 然后继续执行，比如无关紧要的问题
-   5.极少数情况下可以完全忽略整个错误，但是需要做好代码注释。比如创建临时文件失败了可以不用管，操作系统定期会回收。

EOF(End of File): io package 保证任何由 end-of-file 造成的读取失败都会被报告成为另一种不同的 error:io.EOF

```
package io
import "errors"
// EOF is the error returned by Read when no more input is available.
var EOF = errors.New("EOF")

func main() {
    in := bufio.NewReader(os.Stdin)
    for {
	    r, _, err := in.ReadRune()
	    if err == io.EOF {
		    break // finished reading
	    }
	    if err != nil {
		    return fmt.Errorf("read failed: %v", err)
	    }
	    // ...use r...
    }
}
```

## 5.5 Function Values

函数在 go 里是第一公民，和其他值一样，有自己的类型，可以被赋值给变量，可以当做函数参数或者返回值（function 零值是
nil，调用会直接 panic）。函数可以和 nil 比较但是无法互相比较，不能作为map 的 key

         func add1(r rune) rune { return r + 1 }
         fmt.Println(strings.Map(add1, "HAL-9000")) // "IBM.:111"

## 5.6 Anomymous Functions

function literal: 写起来和函数声明一样，但是func 后没有名称。它是表达式，值叫做匿名函数

         fmt.Println(strings.Map(func(r, rune) rune {return r + 1}, "HAL-9000")) // "IBM.:111"

go里同样支持闭包:

    package main

    import "fmt"

    func squares() func() int {
    	var x int
    	return func() int {
    		x++
    		return x * x
    	}
    }
    func main() {
    	f := squares()
    	fmt.Println(f()) // "1"
    	fmt.Println(f()) // "4"
    	fmt.Println(f()) // "9"
    	fmt.Println(f()) // "16"
    }

来看一个 go 的lexical scope 规则导致的意外结果:

    // 考虑一个创建一堆目录然后删除他们的程序
    func main() {
    	var rmdirs []func() // 函数数组
    	for _, d := range tempDirs() {
    		dir := d               // NOTE: necessary!
    		os.MkdirAll(dir, 0755) // creates parent directories too
    		rmdirs = append(rmdirs, func() {
    			os.RemoveAll(dir)
    		})
    	}
    	// ...do some work...
    	for _, rmdir := range rmdirs {
    		rmdir() // clean up
    	}
    }

loop 循环引入了一个新的词法块，d 在这里声明，所有在 loop
里的函数值捕获并且共享同一个变量，而不是特定情况下当时的值。当执行os.RemoveAll的时候 dir 已经被 "now-completed"的 for
循环更新到了最后一个。解决方式一般有两种：使用一个内部变量覆盖。比如上例 dir:=d。或者作为参数传给匿名函数

## 5.7 Variadic Functions

可变函数：可以接受变长参数的函数。比如 fmt.Printf。可以通过声明最后一个参数是 省略号"..."来指示

    package main

    import "fmt"

    func sum(vals ...int) int { //type of vals is an []int slice
    	total := 0
    	for _, val := range vals {
    		total += val
    	}
    	return total
    }

    func main() {
    	fmt.Println(sum())        // "0"
    	fmt.Println(sum(1, 2, 3)) // "3"
    	values := []int{1, 2, 3}
    	fmt.Println(sum(values...))    // slice 参数可以通过后缀省略号传入可变参数的函数中
    }

## 5.8 Deferred Funciton Calls

之前我们写过一个获取 html 页面然后拿到 title 的代码:

    package main

    import "fmt"

    func title(url string) error {
    	resp, err := http.Get(url)
    	if err != nil {
    		return err
    	}
    	// Check Content-Type is HTML (e.g., "text/html; charset=utf-8").
    	ct := resp.Header.Get("Content-Type")
    	if ct != "text/html" && !strings.HasPrefix(ct, "text/html;") {
    		resp.Body.Close()
    		return fmt.Errorf("%s has type %s, not text/html", url, ct)
    	}
    	doc, err := html.Parse(resp.Body)
    	resp.Body.Close()
    	if err != nil {
    		return fmt.Errorf("parsing %s as HTML: %v", url, err)
    	}
    	visitNode := func(n *html.Node) {
    		if n.Type == html.ElementNode && n.Data == "title" &&
    			n.FirstChild != nil {
    			fmt.Println(n.FirstChild.Data)
    		}
    	}
    	forEachNode(doc, visitNode, nil)
    	return nil
    }

注意到这里的 `resp.Body.Close()` 重复了几次吗，如果逻辑更复杂就需要更多处的调用，为了消除这个问题，go 提供了 defer
声明。语法上来说就是 普通的函数或者方法前面加个 defer 前缀，这样就会在包含 defer
的函数完成的时候才调用，无论是函数正常return 或者 panic 了。多个 defer 声明会按照声明顺序逆序执行。

defer 经常用在结对操作中，比如打开或关闭，连接或断开，加锁和解锁，保证资源正确释放，而不用管代码逻辑多复杂。

    //用 defer 改写上例
    func title(url string) error {
    	resp, err := http.Get(url)
    	if err != nil {
    		return err
    	}
    	defer resp.Body.Close()    // NOTE：使用 defer
    	ct := resp.Header.Get("Content-Type")
    	if ct != "text/html" && !strings.HasPrefix(ct, "text/html;") {
    		return fmt.Errorf("%s has type %s, not text/html", url, ct)
    	}
    	doc, err := html.Parse(resp.Body)
    	if err != nil {
    		return fmt.Errorf("parsing %s as HTML: %v", url, err)
    	}
    	// ...print doc's title element...
    	return nil
    }

比如打开和关闭文件：

    func ReadFile(filename string) ([]byte, error) {
    	f, err := os.Open(filename)
    	if err != ni {
    		return nil, err
    	}
    	defer f.Close()
    	return ReadAll(f)
    }

再比如加锁操作:

    var mu sync.Mutex
    var m = make(map[string]int)

    func lookup(key string) int {
    	mu.Lock()
    	defer mu.Unlock()
    	return m[key]
    }

还可以用 defer 来调试复杂函数：

    func bigSlowOperation() {
    	defer trace("bigSlowOperation")() // don't forget the extra parentheses
    	// ...lots of work...
    	time.Sleep(10 * time.Second) // simulate slow operation by sleeping
    }
    func trace(msg string) func() {
    	start := time.Now()
    	log.Printf("enter %s", msg)
    	return func() { log.Printf("exit %s (%s)", msg, time.Since(start)) }
    }

defer 是在函数 return 语句更新了函数的结果变量之后执行的。我们可以用defer 加上一个匿名函数实现函数完成后打印参数和返回值

    func double(x int) int {
    	return x + x
    }

    func double(x int) (result int) {
    	defer func() { fmt.Printf("double(%d) = %d\n", x, result) }()
    	return x + x
    }

注意由于 defer 是在函数的最后执行的，可能有一些危险问题，比如文件描述符被用光：

    	for _, filename := range filenames {
    		f, err := os.Open(filename)
    		if err != nil {
    			return err
    		}
    		defer f.Close() // NOTE: risky; could run out of file descriptors
    		// ...process f...
    	}

一种解决方式是把循环体包括 defer 单独提取成一个函数

    for _, filename := range filenames {
        if err := doFile(filename); err != nil {
    		return err
    	}
    }
    func doFile(filename string) error {
    	f, err := os.Open(filename)
    	if err != nil {
    		return err
    	}
    	defer f.Close()
    	// ...process f...
    }

## 5.9 Panic

go 的类型系统能在编译器找到大量错误，但是对于数组越界访问，nil 指针反解需要运行时确定，当 go
在运行期间遇到了这种问题就会 panic。在一个典型的 panic 期间，正常的执行流程会停止，所有在那个 goroutine 里的 defer
函数会被执行，然后程序 crash 并附上 log 信息，包括 panic 值和栈信息等。不是所有的 panic 都来自运行期间，内置的 panic
函数可以直接调用，接受任何参数。**panic 最适合做一些不可能出现的场景，比如逻辑上不会被执行到的 switch 分之。**
一个好的实践使用 assert 来验证函数的参数，但是这经常被过度使用

    import "fmt"
    switch s := suit(drawCard()); s {
         case "Clubs":    // ...
         default:
             panic(fmt.Sprintf("invalid suit %q", s)) // Joker?
    	 }
     }

对于运行期间会执行检查的参数，我们没必要自己再去 panic（除非你想附加一些信息）:

    func Reset(x *Buffer) {
    	if x == nil {
    		panic("x is nil")  // 没必要加
    	}
    	x.elements = nil
    }

panic 会造成程序 crash，通常处理严重的错误，比如逻辑不一致，勤奋的程序员认为任何 crash 都是有 bug
的证明。在一个鲁棒的系统中，期望的错误，比如输入不合法、配置不合法、或者失败的 IO，应该被温和对待，最好用错误值处理。

## 5.10 Recover

panic 后我们可以选择不 crash，比如 http 服务器应该关闭连接而不是直接crash 让客户端等待。
如果内置的 recover 函数在一个 defer 函数中使用并且 defer 正在 panicking，recover 会结束当前 panic 的状态然后返回panic
值。其他情况调用 recover 直接返回 nil 并且无任何影响。

    func Parse(input string) (s *Syntax, err error) {
    	defer func() {
    		if p := recover(); p != nil {
    			err = fmt.Errorf("internal error: %v", p)
    		}
    	}()
    }
    // ...parser...

应该有选择的使用，过度使用会导致问题被隐藏起来。

    package main

    import "fmt"

    func soleTitle(doc *html.Node) (title string, err error) {
    	type bailout struct{}
    	defer func() {
    		switch p := recover(); p {
    		case nil:
    			// no panic
    		case bailout{}:
    			// "expected" panic
    			err = fmt.Errorf("multiple title elements")
    		default:
    			panic(p) // unexpected panic; carry on panicking
    		}
    	}()
    	// Bail out of recursion if we find more than one non-empty title.
    	forEachNode(doc, func(n *html.Node) {
    		if n.Type == html.ElementNode && n.Data == "title" &&
    			n.FirstChild != nil {
    			if title != "" {
    				panic(bailout{}) // multiple title elements
    			}
    			title = n.FirstChild.Data
    		}
    	}, nil)
    	if title == "" {
    		return "", fmt.Errorf("no title element")
    	}
    	return title, nil
    }

# 6. Methods

1990年代以后，OOP 编程统治了工业和教育领域，几乎被所有所有广泛使用的编程语言支持。
关于面向对象编程实际并没有一个通用的被广泛接受的定义，我们这里简单地认为对象就是一个包含方法的值或者变量，方法是和特定类型绑定
在一起的函数。一个面向对象的程序使用方法来操作数据（客户端不用直接和对象的表示交互）。本章讨论方法和 OOP
的两大原则：封装和组合。

## 6.1 Method Declarations

方法的声明就是在func 关键字后加上绑定的类型

    package main

    import (
    	"fmt"
    	"math"
    )

    type Point struct{ X, Y float64 }

    // 传统函数
    func Distance(p, q Point) float64 {
    	return math.Hypot(q.X-p.X, q.Y-p.Y)
    }

    // 作为 Point 类型的方法，func 之后加了一个 (p Point) 作为 receiver
    func (p Point) Distance(q Point) float64 { // p 叫做方法的接受者 receiver
    	return math.Hypot(q.X-p.X, q.Y-p.Y)
    }
    func main() {
    	p := Point{1, 2}
    	q := Point{4, 6}
    	fmt.Println(Distance(p, q)) // function call
    	fmt.Println(p.Distance(q))  // method call, p.Distance 叫做 selector
    }

因为每种类型对于方法都有自己的名字空间，我们可以使用一样的名字，只要不是同一个类型就行。

    type Path []Point   // slice type 也可以定义 method
    func (path Path) Distance() float64 {
    	sum := 0.0
    	for i := range path {
    		if i > 0 {
    			sum += path[i-1].Distance(path[i])
    		}
    	}
    	return sum
    }

## 6.2 Methods with a Pointer Receiver

当我们想避免值拷贝或者想修改参数内容的时候，可以传递类型指针

    func (p *Point) ScaleBy(factor float64) { // 方法名是 (*Point).ScaleBy
    	p.X *= factor
    	p.Y *= factor
    }

    func main() {
    	// 三种调用方式
    	r := &Point{1, 2}
    	r.ScaleBy(2)
    	fmt.Println(*r) // "{2, 4}"

    	p := Point{1, 2}
    	pptr := &p
    	pptr.ScaleBy(2)
    	fmt.Println(p) // "{2, 4}"

    	p := Point{1, 2}
    	(&p).ScaleBy(2)
    	fmt.Println(p) // "{2, 4}"

    }

## Composing Types by Struct Embedding

    package main

    import (
    	"fmt"
    	"image/color"
    )

    type Point struct{ X, Y float64 }

    func (p Point) ScaleBy(factor float64) Point {
    	return Point{p.X * factor, p.Y * factor}
    }

    type ColoredPoint struct {
    	Point    // 匿名嵌入结构体
    	Color color.RGBA
    }

    func main() {
    	red := color.RGBA{255, 0, 0, 255}
    	var p = ColoredPoint{Point{1, 1}, red}
    	p.ScaleBy(2)   // 和访问匿名结构体的field 一样，同样可以访问匿名结构体的方法(匿名方法提升)
    	fmt.Println("%v", p.ScaleBy(2))
    }

通过使用匿名方法提升，和嵌入结构，可以实现一让匿名结构体也能拥有方法，看个例子：

    package main

    import "sync"

    var (
    	mu      sync.Mutex
    	mapping = make(map[string]string)
    )

    // 使用两个 package-level 变量实现 map
    func Lookup(key string) string {
    	mu.Lock()
    	v := mapping[key]
    	mu.Unlock()
    	return v
    }

下边示例和上边功能上等价：

    package main

    import "sync"

    var cache = struct { // 匿名的 struct
    	sync.Mutex
    	mapping map[string]string
    }{
    	mapping: make(map[string]string), //这个大括号里的是初始化语句
    }

    func Lookup(key string) string {
    	cache.Lock() //统一名字 cache。因为 sync.Mutex 是匿名 field，Lock/UnLock 方法被提升到匿名 struct 类型
    	v := cache.mapping[key]
    	cache.Unlock()
    	return v
    }

## 6.4 Method Values and Expressions

method value: p.Distance yields a method value。当一个包的 api 需要的是 function 参数的时候比较有用

    	p := Point{1, 2}
    	q := Point{4, 6}
    	distanceFromP := p.Distance
    	fmt.Println(distanceFromP(q))

method expression: T.f or (\*T).f yields a function value with a regular first parameter taking the place of the
receiver。

    	p := Point{1, 2}
    	q := Point{4, 6}
    	distance := Point.Distance
    	fmt.Println(distance(p, q))

当想针对同类型的值执行不同的 method 的时候比较有用，举个例子：

    type Point struct{ X, Y float64 }

    func (p Point) Add(q Point) Point { return Point{p.X + q.X, p.Y + q.Y} }
    func (p Point) Sub(q Point) Point { return Point{p.X - q.X, p.Y - q.Y} }

    type Path []Point

    func (path Path) TranslateBy(offset Point, add bool) {
    	var op func(p, q Point) Point
    	if add {
    		op = Point.Add
    	} else {
    		op = Point.Sub
    	}
    	for i := range path {
    		// Call either path[i].Add(offset) or path[i].Sub(offset).
    		path[i] = op(path[i], offset)
    	}
    }

## 6.5 Example: Bit Vector Type

之前都是用 map[T]bool 来模拟 set 操作，对于整数集合可以用比特数组来模拟

    package main

    type IntSet struct {
    	words []uint64
    }

    func (s *IntSet) Has(x int) bool {
    	word, bit := x/64, uint(x%64)
    	return word < len(s.words) && s.words[word]&(1<<bit) != 0
    }

    func (s *IntSet) Add(x int) {
    	word, bit := x/64, uint(x%64)    // word index and bit index
    	for word >= len(s.words) {
    		s.words = append(s.words, 0)
    	}
    	s.words[word] |= 1 << bit
    }

    func (s *IntSet) UnionWith(t *IntSet) {
    	for i, tword := range t.words {
    		if i < len(s.words) {
    			s.words[i] |= tword
    		} else {
    			s.words = append(s.words, tword)
    		}
    	}
    }

## 6.6 Encapsulation

如果一个对象的变量或者方法无法被客户端代码访问就是被封装的（信息隐藏）。go 里为了封装一个对象，需要把它变成 struct。
封装有三个好处：

- 阻止客户端直接修改对象变量；
- 防止客户端依赖将来可能会变动的内部细节；
- 阻止客户端代码随意设置对象的值

  ```
type IntSet struct {   //  隐藏 words，防止客户端直接修改
    words []uint64
}
  ```

# 7 Interfaces

接口类型表示其他类型的通用或者抽象，接口允许我们写出更灵活和适配的函数。go
里无需声明一个特定类型需要满足的所有接口，只提供必要的方法就足够了，这种方式允许我们在已有的具体类型上创建新的接口类型。

## 7.1 Interfaces as Contracts

之前介绍的都是具体类型，指定了值的特定操作(what it is and what you can do with it)。go
里有另一种接口类型，一个接口是抽象类型，它不会暴露内部表示或是支持的操作，它只展现它的一些方法。一个接口类型的值，
我们只能知道它的方法提供了什么行为。（这有点像是鸭子类型，不管你啥类型，只在意你有啥方法）

    package fmt
    type Stringer interface {
    	String() string
    }
    // Fmt.Fprintf 只要是我们实现了 String() 接口，都可以作为 Fprintf 的参数

## 7.2 InterFace Types

接口类型指定了一系列方法，可以通过实现这些具体的方法来定义具体类型。(python里可以用 filelike object
理解，只要你实现了read/write/close 等方法我就可以认为你是一个 filelike object，不管你是文件还是 socket 等)
我们探索下 io package 定义的接口：

    package io
    type Reader interface {
    	Read(p []byte) (n int, err error)
    }
    type Closer interface {
    	Close() error
    }

通过组合现有接口可以实现新的接口:

    type ReadWriter interface {
    	Reader    // embedding an interface
    	Writer
    }
    type ReadWriteCloser interface{
    	Reader
    	Writer
    	Closer
    }
    //甚至能混合用
    type ReadWriter interface {
    	Read(p []byte) (n int, err error)
    	Writer
    }

## 7.3 InterFace Satisfaction

一个类型只有在实现了**所有**接口声明的方法时才『满足』一个接口类型。go 程序员里如果说 "a concrete type is a praticular
interface type", 意味着满足接口类型。接口赋值的规则很简单：一个表达式可以被赋值给接口仅当它的类型满足接口类型（实现所有方法）
。空接口类型 interface{} 可以被任何值赋值。

## 7.4 Parsing Flags with flag.Value

flag.Value 标准接口，提供命令行标志注解

    package main

    import (
    	"flag"
    	"fmt"
    	"time"
    )

    var period = flag.Duration("period", 1*time.Second, "sleep period")

    // go build sleep.go
    func main() {
    	flag.Parse()
    	fmt.Printf("sleeping for %v...", *period)
    	time.Sleep(*period)
    	fmt.Println()
    }
    /*
    ./sleep -period 50ms
    ./sleep -period 2m30s
    ./sleep -period 1.5h
    */

    package flag
    type Value interface {
    	String() string
    	Set(string) error    // 解析 string 参数更新 flag
    }

## 7.5 Interface Values

概念上来着，接口值有两个组成部分，一个具体的类型和该类型的值，叫做接口的动态类型和动态值。
对于静态类型的 go，类型是个编译期概念，所以一个类型并不是一个值。类型描述符提供了每个类型的信息。
接口值是可以互相比较的，当动态类型和动态值(前提是动态值可比较，slices,maps,functions无法比较，会 panic）都相等就认为相等
使用 fmt.Printf("%T") 恩可以用来报告接口值的动态值。
需要注意的是空接口值（不包含值）和包含空指针的接口值不完全相同，后者和 nil 判断并不相等。(An Interface Containing a Nil Pointer Is Non-Nil)

## 7.6 Sorting with sort.Interface

sort package 提供了任何序列的原地排序。

    type StringSlice []string

    func (p StringSlice) Len() int           { return len(p) }
    func (p StringSlice) Less(i, j int) bool { return p[i] < p[j] }
    func (p StringSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
    sort.Sort(StringSlice(names))
    // 当然排序字符串序列太常用了，sort 提供了 sort.Strings(names)

定义结构体的排序

    package main

    import (
    	"sort"
    	"time"
    )

    type Track struct {
    	Title  string
    	Artist string
    	Album  string
    	Year   int
    	Length time.Duration
    }

    var tracks = []*Track{
    	{"Go", "Delilah", "From the Roots Up", 2012, length("3m38s")},
    	{"Go", "Moby", "Moby", 1992, length("3m37s")},
    	{"Go Ahead", "Alicia Keys", "As I Am", 2007, length("4m36s")},
    	{"Ready 2 Go", "Martin Solveig", "Smash", 2011, length("4m24s")},
    }

    func length(s string) time.Duration {
    	d, err := time.ParseDuration(s)
    	if err != nil {
    		panic(s)
    	}
    	return d
    }

    type byArtist []*Track

    func (x byArtist) Len() int           { return len(x) }
    func (x byArtist) Less(i, j int) bool { return x[i].Artist < x[j].Artist }
    func (x byArtist) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }
    func main() {
        sort.Sort(byArtist(tracks))
        sort.Sort(sort.Reverse(byArtist(tracks)))
    }

## 7.7 The http.Handler Interface

    package http

    import (
    	"fmt"
    	"log"
    	"net/http"
    )

    /*
    type Handler interface {
    	ServeHTTP(w ResponseWriter, r *Request)
    }

    func ListenAndServe(address string, h Handler) error
    */
    func main() {
    	db := database{"shoes": 50, "socks": 5}
    	log.Fatal(http.ListenAndServe("localhost:8000", db))
    }

    type dollars float32

    func (d dollars) String() string { return fmt.Sprintf("$%.2f", d) }

    type database map[string]dollars

    // 这里只能处理一个路由
    func (db database) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    	for item, price := range db {
    		fmt.Fprintf(w, "%s :%s\n", item, price)
    	}
    }

    //处理多个路由
    func (db database) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    	switch req.URL.Path {
    	case "/list":
    		for item, price := range db {
    			fmt.Fprintf(w, "%s :%s\n", item, price)
    		}
    	case "price":
    		item := req.URL.Query().Get("item")
    		price, ok := db[item]
    		if !ok {
    			w.WriteHeader(http.StatusNotFound)
    			fmt.Fprintf(w, "no such item:%q\n", item)
    			return
    		}
    		fmt.Fprintf(w, "%s\n", price)
    	default:
    		w.WriteHeader(http.StatusNotFound)
    		fmt.Fprintf(w, "no such page:%s\n", req.URL)
    	}
    }

我们也可以使用 http.NewServeMux 实现一样的功能：

    package http

    import (
    	"fmt"
    	"log"
    	"net/http"
    )

    func main() {
    	db := database{"shoes": 500, "socks": 5}
    	mux := http.NewServeMux()
    	mux.Handle("/list", http.HandlerFunc(db.list))  // 注意最后不是 function call，而是类型转换
      // mux.HandleFunc("/list", db.list)
    	log.Fatal(http.ListenAndServe("localhost:8000", mux))
    }

    type dollars float64
    type database map[string]dollars

    func (db database) list(w http.ResponseWriter, req *http.Request) {
    	for item, price := range db {
    		fmt.Fprintf(w, "%s: %s\n", item, price)
    	}
    }

## 7.8 The error interface

整个 errors pacakge 只有四行

    package errors

    func New(text string) error { return &errorString{text} } //每次调用都会产生不同的实例

    type errorString struct{ text string }

    func (e *errorString) Error() string { return e.text }

## 7.10 Type Assertions

类型断言是在接口值上的一种操作，x.(T)，x 是接口类型表达式，T 是类型。类型断言检查动态类型是否和断言类型匹配。
两种可能性：

-   T is concrete type, 检查 x 的动态类型是否是 T，成功的话结构就是动态值，否则 panic

```
package main

import (
    "bytes"
    "io"
    "os"
)

func main() {
    var w io.Writer
    w = os.Stdout
    f := w.(*os.File)      // success: f == os.Stdout
    c := w.(*bytes.Buffer) // panic: interface holds *os.File, not *bytes.Buffer
}
```

-   T is interface type: 检查 x 的动态值是否满足T，检查成功动态值不会被提取出来，结果仍然是接口值，但是具有接口类型 T。
    换言之，对于接口类型的类型断言表达式改变了表达式的类型，使得更多方法可以被访问，但是保留接口值里的接口类型和值组件。

```
var w io.Writer
w = os.Stdout
rw := w.(io.ReadWriter) // success: *os.File has both Read and Write
w = new(ByteCounter)
rw = w.(io.ReadWriter) // panic: *ByteCounter has no Read method
```

如果我们期望两个返回结果，检查不成功也不会 panic，而是返回是否 ok

    var w io.Writer = os.Stdout
    f, ok := w.(*os.File)      // success:  ok, f == os.Stdout
    b, ok := w.(*bytes.Buffer) // failure: !ok, b == nil

## 7.11 Discriminating Errors with Type Assertions

os 包提供了不同的错误类型用来区分到底是哪种错误（权限问题、是否存在等）

    package os
    // PathError records an error and the operation and file path that caused it.
    func IsExist(err error) bool
    func IsNotExist(err error) bool
    func IsPermission(err error) bool

    type PathError struct {
    	Op   string
    	Path string
    	Err  error
    }

    func (e *PathError) Error() string {
    	return e.Op + " " + e.Path + ": " + e.Err.Error()
    }

## 7.12 Querying Behaviors with Inter Type Assertions

    package main

    import "io"

    // writeString writes s to w.
    // If w has a WriteString method, it is invoked instead of w.Write.
    func writeString(w io.Writer, s string) (n int, err error) {
    	type stringWriter interface {
    		WriteString(string) (n int, err error)
    	}
    	if sw, ok := w.(stringWriter); ok {
    		return sw.WriteString(s) // avoid a copy
    	}
    	return w.Write([]byte(s)) // allocate temporary copy
    }

    func writeHeader(w io.Writer, contentType string) error {
    	if _, err := writeString(w, "Content-Type: "); err != nil {
    		return err
    	}
    	if _, err := writeString(w, contentType); err != nil {
    		return err
    	}
    	// ...
    }

## 7.13 Type Switches

接口有两种典型用途：

-   一种是 io.Reader, io.Writer, fmt.Stringer, sort.Interface, http.Handler, error ，具体类型具有的相似行为，强调方法而不是具体类型。
-   另一种是接口值可以保存任何具体类型的值，可以认为是这些类型的合集。类型断言式用来根据不同的类型区分对待。
    这里强调的是满足接口的具体类型，而不是接口的方法。(discriminated unions)

如果你熟悉 OOP，这两种风格的对应子类型多态(subtype polymorphism)和临时多态(ad hoc polymorphism)。后边讨论第二种形式。
type switch:

    switch x.(type) {
    case nil:       // ...
    case int, uint: // ...
    case bool: //
    case string: //
    default: //
    }
    switch x : x.(type) { /* ... */ }

用 type switch 写一个转化 sql 变量的函数:

    package main

    import "fmt"

    func sqlQuote(x interface{}) string {    // interface{} 允许函数接受任何类型的值
    	switch x := x.(type) {
    	case nil:
    		return "NULL"
    	case int, uint:
    		return fmt.Sprintf("%d", x) // x has type interface{} here.
    	case bool:
    		if x {
    			return "TRUE"
    		}
    		return "FALSE"
    	case string:
    		return sqlQuoteString(x) // (not shown)
    	default:
    		panic(fmt.Sprintf("unexpected type %T: %v", x, x))
    	}
    }

## 7.15 A Few Words of Advice

什么时候需要接口：**当两种以上的具体类型必须以一种统一 的方式处理的时候。**

有一种例外：当接口被一个具体类型满足但是这个类型却无法和接口保持在一个 package 里（因为依赖关系），这个时候接口是解耦两个 package 的好方式。

设计接口的原则：ask only for what you need
