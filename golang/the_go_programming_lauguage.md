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

接下来是一个爬虫的例子，当然看起来没有用 python requests 那么优雅：

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


    switch coinflip() {
    case "heads":    // case 还支持简单的语句 ()
    	heads++
    case "tails":
    	tails++
    default:
    	fmt.Println("landed on edge!")
    }

-   Named Types:


    //定义一个 Point 类型
    type Point struct {
    	X, Y int
    }
    var p Point

-   Pointers(指针)：和 C 类似，go 中也实现了指针
-   Methods and interfaces（方法和接口）: 方法是关联到一个命名类型的函数。接口是一种把不同类型同等对待的抽象类型
-   Packages(包): 通过包组织程序
-   Comments(注释)：和 C 一样的注释,  `// or /* XXXX */`

# 2. Program Structure

表达式+控制流 -> 语句 -> 函数 -> 源文件 -> 包

## 2.1 Names

go 定义了几十个关键字，不能用来给变量命名，go 使用一般使用骆驼命名法(HTTP等缩略词除外)。需要注意的是只有大骆驼命名
"fmt.Fprintf" 这种是可以被其他包引入使用的。

## 2.2 Declarations

var, const ,type, func

    func fToC(f float64) float64 {
    	return (f - 32) * 5 / 9
    }

## 2.3 Variables

定义变量：如果省略了 type 将会用每个类型的初始值初始化赋值（类似 java）

    var name type = expression
    // 可以一次性定义多个
    var i, j, k int
    var b, f, s  = true, 2.3. "four"    // 同样有类似解包的操作

    // 接下来是 short variable declarations， 短赋值，`:=` 是 声明，而 `=` 是赋值
    freq := rand.Float64() * 3.0
    t := 0.0

    // multiple variables declared
    i, j = 0, 1
    i, j = j, i    // swap

    f, err := os.Open()
    f, err := os.Close()     // wrong,  a short variable declaration must declare at least one new variable

指针: 如果学过 c，这里的指针很类似，表示一个变量的地址

    x := 1
    p := &x    // p, of type *int, points to x
    fm.Println(*p) // "1"
    *p = 2
    fmt.Println(x) // "2"

    var x, y int
    fmt.Println(&x == &x, &x == &y, &x == nil) // "true false false"，指针可以比较，当指向相同的值的时候相等

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
    	x,y = y,x   // swap x and y like python
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

每个包都声明了名字空间，当有名字冲突的时候，需要显示通过包名调用。大写开头的标识符才可以被导出。把上面的例子改成包，

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
    有符号数： -2**(n-1) to 2**(n-1)-1
    有符号数: 0 to 2\*\*n-1

    注意不同类型之间数字强转可能会有精度损失

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
-   短路求值特性

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

几个用来处理字符串的包：

-   bytes:  操作 slices of bytes, type \[]byte
-   strings: searching, replacing, comparing, trimming, splitting, joining
-   srconv: boolean, integer, floating-point values 和 他们的 string 表示形式来回转；
-   unicode: IsDigit, IsLetter, IsUpper, IsLower识别 runes

## 3.6 Constants

const 表达式的值在编译器确定，无法改变。const 值可以是boolean, string or number
The const generator iota: 定义从0 开始的递增枚举

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

本章讨论聚合类型: arrays, slices, maps, structsa

## 4.1 Arrays

定长的包含0个或多个相同类型元素的序列，因为不可变长所以一般 slices 更常用。
注意数组的长度也是类型的一部分， [3]int 和 [4]int 是不同类型，相同长度的可以比较。

    	var a [3]int // array of 3 integers，第一次看这种写法略奇怪
    	fmt.Println(a[0])
    	for i, v := range a {
    		fmt.Printf("%d %d\n", i, v)
    	}

    	var q [3]int = [3]int{1, 2, 3} //初始化
    	p := [...]int{1, 2, 3}         // 省略号代表长度由右边的长度决定

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

数组传值传过去拷贝，可以通过传递指针

    func zero(ptr *[32]byte) {
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

注意，slice 包含了一个数组的指针，作为函数参数传递的时候允许函数修改数组。

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

并且 slice 不像数组一样能直接比较，基于两个原因：

-   slice 可以包含自己，深度比较无论是复杂度还是实现效率都不好
-   不适合作为 map 的 key。对于引用类型（pointers and channels）， == 一般比较引用实体是否是同一个。但是这回导致 slice
    和数组的 == 一个是 shallow equality，一个是 deep equality，为了防止歧义最安全的方式就是干脆禁止 slice 比较。唯一合法的
    slice 比较就是判断 `if slice == nil {}`

slice 的初始化值（zero value）是 nil，nil slice没有隐含指向的数组，其 length==capacity==0。但是也有包含 nil
的slice，我们看几个比较(我感觉这里 python 的 None 用 is 判断比较优雅)：

    var s []int    // len(s) == 0, s == nil
    s = nil        // len(s) == 0, s == nil
    s = []int(nil) // len(s) == 0, s == nil
    s = []int{}    // len(s) == 0, s != nil

所以如果只是判断一个 slice 是否为空，应该用 len(s) == 0，而不是 s == nil。内置的 make 函数用来创建 slice

    	make([]T, len)
    	make([]T, len, cap)

向 slice 追加元素用内置的 append 函数

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

可以使用省略号来让函数接受任意多个 final arguments：

    func appendInt(x []int, y ...int) []int{
    	var z[]int
    	zlen:=len(x) + len(y)
    	// ... expand z to at least zlen...
    	copy(z[len(x):], y)
    	return z
    }

## 4.3 Maps

无序键值对，go 里 map 是 哈希表的引用。要求所有 key 类型相同，所有的 value 类型相同，key
必须是可比较的。不过用法上和 python 的 map 还是比较类似的。

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

    	ages["bob"] = ages["bob"] + 1 // bob不存在的 key 默认是0，不像 py 访问不存在的 key 会 KeyError
    	delete(ages, "alice")         //删除 key
    	fmt.Println(ages)

    	// 遍历无序的 kv
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

go 里没提供 set 类型，我们可以通过 map 来模拟。key 不支持 slice 类型，但是如果想用的话可以用转换函数转成 string 后来作为
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

```
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
```

结构体字面量：

```
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
```

可比较性：如果结构体的所有 field 都是可以比较的，struct 本身也可以比较。

golang的 struct 允许嵌套：


```
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
```

之前这种连续赋值写起来很繁琐，可以用匿名 field 简化(其实不喜欢这种方式，虽然多写一些麻烦了点，但是很直观)：
```
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
```

## 4.5 JSON

JavaScript Object Notation (JSON) is a standard notation for sending and receiving structured information.
JSON is an encoding of JavaScript values—strings, numbers, booleans, arrays, and objects—as Unicode text
Go data structure like movies to JSON is called marshaling. Marshaling is done by json.Marshal:

```
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
	Color  bool `json:"color,omitempty"`
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

```

注意到 field 后边的 `json:"released"` 叫做 field tags，是在编译器旧绑定到 field 上的元信息，可以控制序列化/反序列化的行为。
field tag 第一部分指定了 field 的 json 名称（比如从骆驼命名改成下划线命名），第二个可选的选项（imitempty），指定了当如果该 field 是零值的时候不会输出该字段.
