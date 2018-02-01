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

```
var a=b+c // 3
var b=f() // 2
var c=1//1
func f() int {return c+1}
```

如果一个包有多个 go 源文件，会被给编译器的顺序执行初始化，go tools 通过把文件名排序后给交给编译器
可以在文件里定义 init() 函数，他们不能被引用，程序执行的时候会按照被声明顺序自动运行(比如初始化一个查找表)
```
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
```

## 2.7 Scope
声明绑定了名字和程序实体，比如函数或者变量，一个声明的作用域表示这个声明的名字在哪些代码块起作用。不要和生命周期（lifetime）
混淆，声明的作用域是一段程序片段，编译时属性，生命周期指的是运行期间可以被程序其他部分所引用的持续时间，运行时属性。
golang 中通过块（block）来圈定作用域。名字查找遵循『就近』原则，内部声明会屏蔽外部声明（最好不要重名，理解各种重名覆盖问题比较费劲）

# 3. Basic Data Types
go的数据类型分成4类：
- 基础类型: numbers, strings, booleans
- 聚合类型: arrays, structs,
- 引用类型: pointers, slices, maps, functions, channels
- 接口类型: interface types

 ## 3.1 Integers
 int8, int16, int32, int64, uint8, uint16, uint32, uint64
 rune <=> int32 , byte <=> uint8
 有符号数： -2**(n-1) to 2**(n-1)-1
 有符号数: 0 to 2**n-1

 注意不同类型之间数字强转可能会有精度损失

## 3.2 Floating-Point Numbers
- float32: 1.4e-45 to 3.4e38
- float64: 1.8e308 to 4.9e-324
- math package

## 3.3 Complex Numbers
- complex64
- complex128
- math/cmplx package

  `var x complex128 = complex(1,2)`

## 3.4 booleans
- true
- false
- 短路求值特性
