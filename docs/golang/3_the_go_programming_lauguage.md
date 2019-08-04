# 10 Packages and the Go Tool

## 10.1 Introduction

包系统都是通过把相关特性组织成单元，使得大型程序的设计和维护更易理解和修改，和程序的其他包隔离。
模块化使得包更同意被分享、重用。go 里每个 package
都定义了一组包含实体的命名空间，防止和其他包冲突。并且通过命名来区分可见性。

## 10.2 Import Paths

每个包都被一个独一无二的字符串标识，叫做导入路径(import path)。对于你想分享或者发布的包，导入路径必须是全局唯一的。
为了防止冲突，非官方的包都应该包含域名或者组织的前缀。

    import (
    	"fmt"
    	"math"

    	"golang.org/x/net/html"
    	"github.com/go-sql-driver/mysql"
    )

## 10.3 The Pacakge Decalration

包声明需要在每个 go 源文件头声明，主要是为了当包被导入的时候决定默认标识符。

    import (
    	"fmt"
    	"math/rand"
    )

比如这里我们可以通过 rand.Int 访问。常规上包名是导入路径的最后一段（后边会讨论重名冲突问题），但是有三种例外：

-   一种是像 main 这种直接被 buildo 成可执行文件的。
-   一些目录里的文件可能包含 `_test` 后缀（一般是 `_test.go`测试文件）
-   对于有版本号的 `gopkg.in/yaml.v2` 直接用 yaml 就行

## 10.4 Import Decalration

导入可以每行导入一个或者统一放在括号里（笔者用的 vim-go 设置了自动导入，保存文件的时候会自动处理导入
pacakge，其他编辑器或者 IDE 都有类似的设置）
如果有包名冲突了，可以用 rename 的方式防止冲突。

    import (
    	"crypto/rand"
    	mrand "math/rand"   // 别名防止冲突，也可以用来简化一些比较长的 package name
    )

## 10.5 Blank Imports

如果导入包但是不用是会报错的，但是有些场景比如使用是在 init 里的，这个时候就需要使用 empty import

    import _ "image/png"   // 其实就是 rename 成了 _

## 10.6 Pacakges and Naming

讲下包的命名:

-   尽量短小，但不要难以理解
-   具有描述性但不要有歧义
-   一般用单数形式，但是内置的 bytes, errors, strings 是防止和关键字冲突
-   避免使用已经具有其他含义的名称，比如 temp（一般理解为临时文件相关），我们的转换温度的包改成 tempconv

讲下包成员的命名：

-   考虑到和包结合使用时候的含义。 bytes.Equal, flag.Int, http.Get, json.Marshal
-   保持精简，不要和包名重复开头。比如 strings.Index not strings.StringIndex

## 10.7 The Go Tool

go tool 提供了下载、查询、格式化、构建、测试、安装包等功能。直接输入 go 回车就可以看到列出的工具，使用 go help command
可以查看具体使用方法。

### 10.7.1 Workspace Organization

大部分用户唯一需要设置的就是 GOPATH 环境变量，指定了工作区的根目录，切换工作区只需要改变 GOPATH。
GOPATH 里有仨子目录:

-   src: 保存源文件。相对于 `$GOPATH/src`就是它的导入路径
-   pkg：build tools 存储编译的包
-   bin: 保存可执行文件

GOROOT: 指定 go 发布的根目录，提供标注库的包的路径。可以用 go env 查看

### 10.7.2 Downloading Packages

使用go tool的时候，一个包的导入路径不仅指定了本地如何查找，还指定了网络上哪里可以下载和更新包。
go get 会创建远程仓库的本地客户端，不仅仅是拷贝文件，所以可以用版本控制工具看历史等

### 10.7.3 Building Packages

go build 会编译指定包，如果是一个库，结果会丢弃，仅仅检查包是否有编译错误。如果是 main 就会 linker
生成可执行程序，并且用导入路径的最后一段来命名。
默认 go build 会 build 指定的包和所有它的依赖，然后丢弃所有编译错误，除非是最终可执行文件。
go install(或者 go build -i) 和 go build 很像，但是会保存所有每个编译完的包，如果包没更新能大大加快编译速度。

### 10.7.4 Documenting Packages

Go 风格强烈建议给包 API 加上良好的文档，每个导出的 package member 和包自己都应该在前面加上『意图』和『用法』注释。
大部分声明都可以用一句话描述，如果行为非常明显，可以省略注释。

`go doc time` or `go doc time.Since`

godoc 提供了 html 文档

`godoc -http :8000`

### 10.7.5 Internal Packages

go 的可见性设置很简单，未导入标志符仅对同一个包可见，导出标识符对『全世界』都可见。
go 提供了 internal packages，只能被同级根的父目录导入

    net/http
    net/http/internal/chunked
    net/http/httputil
    net/url<Paste>

`net/httpinternal/chunked`只能被 `net/http/httputil net/http`导入

### 10.7.6 Querying Pacakges

go list 可以用来查找包

# 11. Testing

当今程序越来越复杂，控制复杂度需要付出越来越多努力。有两个有效的方式被广泛使用，一个是代码审查，还有一个就是本章要讲的测试。

## 11.1 The go test Tool

go test 是go 包的测试驱动，在一个包目录下以 `_test.go` 结尾的文件会被 go test 而不是 go build 构建。这些 `*_test.go`
文件有三种函数被特殊对待：

-   tests: 以 Test 开头的测试函数，用来测试程序逻辑，go test 调用测试函数报告是 PASS 或者 FAIL
-   benchmarks Benchmark开头用来测试一些操作的执行性能
-   examples:  Example 开头，提供机器检测文档

go test 扫描 `*_test.go` 文件里这些特殊函数，生成一个临时的 main 包调用它们并且报告结果。

## 11.2 Test Functions

每个测试文件都必须导入 testing package，测试函数签名如下：

    func TestName(t *testing.T) {
    // ...
    }

让我们来写个判断回文数的 pacakge 和测试：

    package word

    // IsPalindrome 判断是否是回文字符串
    func IsPalindrome(s string) bool {
    	for i := range s {
    		if s[i] != s[len(s)-1-i] {
    			return false
    		}
    	}
    	return true
    }

这里写一个 word package ，然后在该 package 目录下写个测试文件：word_test.go

    package word

    import "testing"

    func TestPalindrome(t *testing.T) {
    	if !IsPalindrome("detartrated") {
    		t.Error(`IsPalindrome("detartrated") = false`)
    	}
    	if !IsPalindrome("kayak") {
    		t.Error(`IsPalindrome("kayak") = false`)
    	}
    }

    func TestNonPalindrome(t *testing.T) {
    	if IsPalindrome("palindrome") {
    		t.Error(`IsPalindrome("palindrome") = true`)
    	}
    }

运行 go test 我们会看到没有 FAIL 的输出，测试用例都通过。这个时候测试工程师提 bug 啦，说有用例用你的函数不能过，并且提供了测试用例。
我们在测试代码里加上如下测试：

    func TestFrenchPalindrome(t *testing.T) {
    	if !IsPalindrome("été") {
    		t.Error(`IsPalindrome("été") = false`)
    	}
    }
    func TestCanalPalindrome(t *testing.T) {
    	input := "A man, a plan, a canal: Panama"
    	if !IsPalindrome(input) {
    		t.Errorf(`IsPalindrome(%q) = false`, input)
    	}
    }

这次再运行下 go test，发现有俩 FAIL 的测试用例(使用 `go test -v`会输出所有成功和失败用例) 。嗯，确实有 bug。
还可以加上 `-run` 参数和一个正则表达式，只跑我们想跑的测试函数 `go test -v -run="French|Canal"`
这样我们发现了俩 bug，无法判断 rune 序列(就是 python 里的unicode)，并且没有忽略掉空格、标点和大小写。我们改下代码：

    package word

    import "unicode"

    // IsPalindrome 判断是否是回文字符串
    func IsPalindrome(s string) bool {
    	var letters []rune
    	for _, r := range s {
    		if unicode.IsLetter(r) {
    			letters = append(letters, unicode.ToLower(r))
    		}
    	}
    	for i := range letters {
    		if letters[i] != letters[len(letters)-1-i] {
    			return false
    		}
    	}
    	return true
    }

这个时候再次运行 go test 会发现测试用例都通过了。
但是为了保险，我们再加一些测试用例(单测的好处就是你不怕修改代码会影响以前正确的逻辑，方便回归测试和重构，所以不能只 print
出来就觉得正确了，而要落地到测试文件里，笔者之前呆过的小团队就有人不会写测试，全靠输出，这样是很不好的习惯)。
这次用到了表驱动法方便增删用例：

    func TestIsPalindrome(t *testing.T) {
    	var tests = []struct {
    		input string
    		want  bool
    	}{
    		{"", true},
    		{"a", true},
    		{"aa", true},
    		{"ab", false},
    		{"kayak", true},
    		{"detartrated", true},
    		{"A man, a plan, a canal: Panama", true},
    		{"Evil I did dwell; lewd did I live.", true},
    		{"Able was I ere I saw Elba", true},
    		{"été", true},
    		{"Et se resservir, ivresse reste.", true},
    		{"palindrome", false}, // non-palindrome
    		{"desserts", false},   // semi-palindrome
    	}
    	for _, test := range tests {
    		if got := IsPalindrome(test.input); got != test.want {
    			t.Errorf("IsPalindrome(%q) = %v", test.input, got)
    		}
    	}
    }

### 11.2.1 Randomized Testing

如果是随机输入怎么测试呢？两种方式：

-   1.使用一种更加简单明了的算法实现，对比两个输出（对拍）
-   2.写函数构造正确输出看看要测试的函数输出是否正确

这里采用方式2，我们写函数来构造回文数:

    package word
    import (
    	"math/rand"
    	"testing"
    	"time"
    )

    // randomPalindrome returns a palindrome whose length and contents
    // are derived from the pseudo-random number generator rng.
    func randomPalindrome(rng *rand.Rand) string {
    	n := rng.Intn(25) // random length up to 24
    	runes := make([]rune, n)
    	for i := 0; i < (n+1)/2; i++ {
    		r := rune(rng.Intn(0x1000)) // random rune up to '\u0999'
    		runes[i] = r
    		runes[n-1-i] = r
    	}
    	return string(runes)
    }
    func TestRandomPalindromes(t *testing.T) {
    	// Initialize a pseudo-random number generator.
    	seed := time.Now().UTC().UnixNano()    // seed 值可以用来复现测试
    	t.Logf("Random seed: %d", seed)
    	rng := rand.New(rand.NewSource(seed))
    	for i := 0; i < 1000; i++ {
    		p := randomPalindrome(rng)
    		if !IsPalindrome(p) {
    			t.Errorf("IsPalindrome(%q) = false", p)
    		}
    	}
    }

然后运行 `go test -v`，可以看到 PASS 了。注意对于随机测试是未决的(nondeterministic)，
所以如果有错误要提供足够的输出结果用来排查问题。注意这里的随机种子，如果测试未通过我们可以用一样的种子复现上一次测试。
可以用 t.t.Fatalf 在遇到失败测试的时候停止，必须和 Test 在同一个goroutine 调用。

### 11.2.2 Testing a Command

本节讲如何测试命令行参数相关的程序。写个打印命令行参数的程序：

    package main

    import (
    	"flag"
    	"fmt"
    	"io"
    	"os"
    	"strings"
    )

    var (
    	n = flag.Bool("n", false, "imit trailing newline")
    	s = flag.String("s", " ", "separator")
    )

    var out io.Writer = os.Stdout //输出重定向，这样就能用来测试

    func echo(newline bool, sep string, args []string) error {
    	fmt.Fprintf(out, strings.Join(args, sep))
    	if newline {
    		fmt.Fprintln(out)
    	}
    	return nil
    }

    func main() {
    	flag.Parse()
    	if err := echo(!*n, *s, flag.Args()); err != nil {
    		fmt.Fprintf(os.Stderr, "echo : %v\n", err)
    		os.Exit(1)
    	}
    }

然后是对应的测试代码

    package main

    import (
    	"bytes"
    	"fmt"
    	"testing"
    )

    func TestEcho(t *testing.T) {
    	var tests = []struct {
    		newline bool
    		sep     string
    		args    []string
    		want    string
    	}{
    		{true, "", []string{}, "\n"},
    		{false, "", []string{}, ""},
    		{true, "\t", []string{"one", "two", "three"}, "one\ttwo\tthree\n"},
    		{true, ",", []string{"a", "b", "c"}, "a,b,c\n"},
    		{false, ":", []string{"1", "2", "3"}, "1:2:3"},
    	}

    	for _, test := range tests {
    		descr := fmt.Sprintf("echo(%v, %q, %q)", test.newline, test.sep, test.args)
    		out = new(bytes.Buffer) // captured output
    		if err := echo(test.newline, test.sep, test.args); err != nil {
    			t.Errorf("%s failed: %v", descr, err)
    			continue
    		}
    		got := out.(*bytes.Buffer).String()
    		if got != test.want {
    			t.Errorf("%s = %q, want %q", descr, got, test.want)
    		}
    	}
    }

注意这里的 main 在测试的时候被当做了 library，go test 会忽略 main 函数。
注意其实被测试的代码不应该调用 log.Fatal or os.Exit，会导致进程停止，应该放在 main 里。

### 11.2.3 White-Box Testing

测试一般有白盒和黑盒测试，黑盒测试要测试的包其内部实现对测试代码来说是不可见的，我们只能测试导出api是否符合逻辑，
黑盒测试的代码一般更健壮，不会被测试代码的内部修改影响。
白盒测试则可以深入到被测试代码的具体内部实现，提供更加精细的测试，甚至可以修改其内部代码实现 fake implementation.
考虑一种很常见的场景，实现代码里会真实发送邮件，但是我们不想测试的时候也发送邮件：

    package storage

    import (
    	"fmt"
    	"log"
    	"net/smtp"
    )

    func bytesInUse(username string) int64 { return 0 /* ... */ }

    // Email sender configuration.
    // NOTE: never put passwords in source code!
    const sender = "notifications@example.com"
    const password = "correcthorsebatterystaple"
    const hostname = "smtp.example.com"
    const template = `Warning: you are using %d bytes of storage, %d%% of your quota.`

    func CheckQuota(username string) {
    	used := bytesInUse(username)
    	const quota = 1000000000 // 1GB
    	percent := 100 * used / quota
    	if percent < 90 {
    		return // OK
    	}
    	msg := fmt.Sprintf(template, used, percent)
    	auth := smtp.PlainAuth("", sender, password, hostname)
    	err := smtp.SendMail(hostname+":587", auth, sender,
    		[]string{username}, []byte(msg))
    	if err != nil {
    		log.Printf("smtp.SendMail(%s) failed: %s", username, err)
    	}
    }

为了不真实发送邮件，把邮件逻辑放到一个函数里然后赋值给一个不可导出的包级别变量：

    var notifyUser = func(username, msg string) {
    	auth := smtp.PlainAuth("", sender, password, hostname)
    	err := smtp.SendMail(hostname+":587", auth, sender,
    		[]string{username}, []byte(msg))
    	if err != nil {
    		log.Printf("smtp.SendEmail(%s) failed: %s", username, err)
    	}
    }

    func CheckQuota(username string) {
    	used := bytesInUse(username)
    	const quota = 1000000000 // 1GB
    	percent := 100 * used / quota
    	if percent < 90 {
    		return // OK
    	}
    	msg := fmt.Sprintf(template, used, percent)
    	notifyUser(username, msg)
    }

之后在测试代码里我们把 notifyUser 函数给修改掉（还是 python 的 mock 装饰器爽多了啊）

    package storage

    import (
    	"strings"
    	"testing"
    )

    func TestCheckQuotaNotifiesUser(t *testing.T) {
    	var notifiedUser, notifiedMsg string
    	notifyUser = func(user, msg string) {    // 注意这里把原来的函数 替换掉了，不让它真发邮件
    		notifiedUser, notifiedMsg = user, msg
    	}
    	// ...simulate a 980MB-used condition...
    	const user = "joe@example.org"
    	CheckQuota(user)
    	if notifiedUser == "" && notifiedMsg == "" {
    		t.Fatalf("notifyUser not called")
    	}
    	if notifiedUser != user {
    		t.Errorf("wrong user (%s) notified, want %s",
    			notifiedUser, user)
    	}
    	const wantSubstring = "98% of your quota"
    	if !strings.Contains(notifiedMsg, wantSubstring) {
    		t.Errorf("unexpected notification message <<%s>>, "+
    			"want substring %q", notifiedMsg, wantSubstring)
    	}
    }

注意，修改完以后必须要改回来

    func TestCheckQuotaNotifiesUser(t *testing.T) {
    	// Save and restore original notifyUser.
    	saved := notifyUser
    	defer func() { notifyUser = saved }()    // 修改完以后，函数完成再给改回来
    	// Install the test's fake notifyUser.
    	var notifiedUser, notifiedMsg string
    	notifyUser = func(user, msg string) {
    		notifiedUser, notifiedMsg = user, msg
    	}
    	// ...rest of test...
    }

这样就实现了『mock』，不过老实说语法相对 python 麻烦多了。通过这种方式可以替换全部变量、命令行参数、调试参数、性能参数等。
go test 一般不会并发跑多个普通测试，所以全部变量的替换也是安全的。

### 11.2.4 External Test Packages

go 是明确禁止环形引用的，但是 net/url 的测试依赖 net/http，低层次的却依赖了高层次的。
通常使用一个额外的测试并在包的测试文件里把需要的函数暴露出去来解决。

### 11.2.5 Writing Effective Tests

你可能会很奇怪 go 的测试框架如此简陋，没有提供 setup、teardown，各种断言函数、格式化错误输出等。
通常这一套测试下来测试代码就像是换了一个语言写的，对维护者来说提示信息也不友好。go
比较极端，让测试作者自己来定义函数，报告错误等，让维护者不用看源代码就能理解失败的测试。
好的测试应该从实现具体的行为开始，仅使用函数来简化和消除重复代码，而不是通过测试库的抽象和通用函数。

### 11.2.6 Avoiding Brittle Tests

避免脆弱的测试，比如稍微改动点输出测试就跪了。

-   仅测试你关心的特性
-   测试简单稳定的接口而不是内部实现函数，内部实现更容易变动
-   慎重选择断言函数，比如不要直接测试字符串是否相等，而是用子串替代

## 11.3 Coverage

> 测试只能证明缺陷存在，而无法证明缺陷不存在。-迪杰斯特拉
> 生成测试覆盖报告 `go tool cover -html=c.out`

## 11.4. Benchmark Functions

benchmark 是一种在固定负载下衡量性能的方式。和测试函数类似，Benchmark 开头， testing.B 作为参数

    package word

    import "testing"

    func BenchmarkIsPalindrome(b *testing.B) {
    	for i := 0; i < b.N; i++ {
    		IsPalindrome("A man, a plan, a canal: Panama")
    	}
    }

使用 `go test -bench=.` 运行，类似如下结果

    PASS
    BenchmarkIsPalindrome-8 1000000              1035 ns/op
    ok      gopl.io/ch11/word2      2.179s

## 11.5 Profiling

benchmark方便地对特定操作度量，但是如果想要优化程序，我们不知道从何处下手。Profiling
通过在程序执行期间进行抽样采集信息，然后分析统计结果得到 profile

    $ go test -cpuprofile=cpu.out
    $ go test -blockprofile=block.out
    $ go test -memprofile=mem.out

更多内容参考官方博客：Profiling Go Programs

## 11.6. Example Functions

go test 特殊对待的第三种函数是 Example 开头的函数，没有参数和返回值

    func ExampleIsPalindrome() {
    	fmt.Println(IsPalindrome("A man, a plan, a canal: Panama"))
    	fmt.Println(IsPalindrome("palindrome"))
    	// Output:
    	// true
    	// false
    }

Example 函数有三种用途：

-   文档示例
-   可以被 go test执行，注意上边例子的 `//Output`
-   及时测试，godoc server，让用户在网页上直接修改代码测试(Go Playground)

# 12. Reflection

go 提供了反射，允许把类型当做一等公民。(在动态语言里很容易实现)
<https://en.wikipedia.org/wiki/Reflection_(computer_programming>)

## 12.1 Why Reflection?

有时候我们想写一个函数实现如下需求：处理统一类型的值但是却不满足一个公共的接口；没有一个已知的表示；甚至在设计函数的时候不存在的值。
比如

    func Sprint(x interface{}) string {
    	type stringer interface {
    		String() string
    	}
    	switch x := x.(type) {
    	case stringer:
    		return x.String()
    	case string:
    		return x
    	case int:
    		return strconv.Itoa(x)
    	// ...similar cases for int16, uint32, and so on...
    	case bool:
    		if x {
    			return "true"
    		}
    		return "false"
    	default:
    		// array, chan, func, map, pointer, slice, struct
    		return "???"
    	}
    }

如果我们没法检测到未知类型的表示，就无法实现这种通用的函数。

## 12.2 reflect.Type and reflect.Value

反射通过 reflect package 实现，它定义了两种重要类型： Type and Value

-   Type: Type 是一个 Go 个接口，提供了很多方法用来区分不同类型检查它们的组件

```
t := reflect.TypeOf(3)  // a reflect.Type, TypeOf 接收 interface{} 并且返回它的动态类型
fmt.Println(t.String()) // "int"
fmt.Println(t)          // "int"
fmt.Printf("%T\n", 3)   // 内部也会使用 reflect.TypeOf
```

- Value: reflect.Value 能保存任何类型的值，reflect.ValueOf function accepts any interface{} and returns a reflect.Value containing the interface’s dynamic value.

```
v := reflect.ValueOf(3) // a reflect.Value
fmt.Println(v)          // "3"
fmt.Printf("%v\n", v)   // "3"
fmt.Println(v.String()) // NOTE: "<int Value>"
```

reflect.ValueOf 和 reflect.Value.Interface 互为反操作

    v := reflect.ValueOf(3) // a reflect.Value
    x := v.Interface() // an interface{}
    i := x.(int) // an int
    fmt.Printf("%d\n", i) // "3"

用 reflect 实现上面的函数。 Although there are infinitely many types, there are only a finite number of kinds of type: the basic types Bool, String, and all the numbers; the aggregate types Array and Struct; the ref- erence types Chan, Func, Ptr, Slice, and Map; Interface types; and finally Invalid, meaning no value at all. (The zero value of a reflect.Value has kind Invalid.)

    package format

    import (
    	"reflect"
    	"strconv"
    )

    // Any formats any value as a string.
    func Any(value interface{}) string {
    	return formatAtom(reflect.ValueOf(value))
    }

    // formatAtom formats a value without inspecting its internal structure.
    func formatAtom(v reflect.Value) string {
    	switch v.Kind() {
    	case reflect.Invalid:
    		return "invalid"
    	case reflect.Int, reflect.Int8, reflect.Int16,
    		reflect.Int32, reflect.Int64:
    		return strconv.FormatInt(v.Int(), 10)
    	case reflect.Uint, reflect.Uint8, reflect.Uint16,
    		reflect.Uint32, reflect.Uint64, reflect.Uintptr:
    		return strconv.FormatUint(v.Uint(), 10)
    	// ...floating-point and complex cases omitted for brevity...
    	case reflect.Bool:
    		return strconv.FormatBool(v.Bool())
    	case reflect.String:
    		return strconv.Quote(v.String())
    	case reflect.Chan, reflect.Func, reflect.Ptr, reflect.Slice, reflect.Map:
    		return v.Type().String() + " 0x" +
    			strconv.FormatUint(uint64(v.Pointer()), 16)
    	default: // reflect.Array, reflect.Struct, reflect.Interface
    		return v.Type().String() + " value"
    	}
    }

## 12.3 Display, A Recursive Value Printer

这一节实现一个能递归打印任何类型值的函数（不能自包含或者有环形结构）

    package format

    import (
    	"fmt"
    	"reflect"
    )

    func Display(name string, x interface{}) {
    	fmt.Printf("Display %s (%T):\n", name, x)
    	display(name, reflect.ValueOf(x))
    }

    func display(path string, v reflect.Value) {
    	switch v.Kind() {
    	case reflect.Invalid:
    		fmt.Printf("%s = invalid\n", path)
    	case reflect.Slice, reflect.Array:
    		for i := 0; i < v.Len(); i++ {
    			display(fmt.Sprintf("%s[%d]", path, i), v.Index(i))
    		}
    	case reflect.Struct:
    		for i := 0; i < v.NumField(); i++ {
    			fieldPath := fmt.Sprintf("%s.%s", path, v.Type().Field(i).Name)
    			display(fieldPath, v.Field(i))
    		}
    	case reflect.Map:
    		for _, key := range v.MapKeys() {
    			display(fmt.Sprintf("%s[%s]", path,
    				formatAtom(key)), v.MapIndex(key))
    		}
    	case reflect.Ptr:
    		if v.IsNil() {
    			fmt.Printf("%s = nil\n", path)
    		} else {
    			display(fmt.Sprintf("(*%s)", path), v.Elem())
    		}
    	case reflect.Interface:
    		if v.IsNil() {
    			fmt.Printf("%s = nil\n", path)
    		} else {
    			fmt.Printf("%s.type = %s\n", path, v.Elem().Type())
    			display(path+".value", v.Elem())
    		}
    	default: // basic types, channels, funcs
    		fmt.Printf("%s = %s\n", path, formatAtom(v))
    	}
    }

现在用它来打印一个自定义结构体：

    type Movie struct {
    	Title, Subtitle string
    	Year            int
    	Color           bool
    	Actor           map[string]string
    	Oscars          []string
    	Sequel          *string
    }

    func main() {
    	strangelove := Movie{
    		Title:    "Dr. Strangelove",
    		Subtitle: "How I Learned to Stop Worrying and Love the Bomb",
    		Year:     1964,
    		Color:    false,
    		Actor: map[string]string{
    			"Dr. Strangelove":            "Peter Sellers",
    			"Grp. Capt. Lionel Mandrake": "Peter Sellers",
    			"Pres. Merkin Muffley":       "Peter Sellers",
    			"Gen. Buck Turgidson":        "George C. Scott",
    			"Brig. Gen. Jack D. Ripper":  "Sterling Hayden",
    			`Maj. T.J. "King" Kong`:      "Slim Pickens",
    		},
    		Oscars: []string{
    			"Best Actor (Nomin.)",
    			"Best Adapted Screenplay (Nomin.)",
    			"Best Director (Nomin.)",
    			"Best Picture (Nomin.)",
    		},
    	}
    	Display("strangelove", strangelove)
    }

输出如下：

    Display strangelove (main.Movie):
    strangelove.Title = "Dr. Strangelove"
    strangelove.Subtitle = "How I Learned to Stop Worrying and Love the Bomb"
    strangelove.Year = 1964
    strangelove.Color = false
    strangelove.Actor["Maj. T.J. \"King\" Kong"] = "Slim Pickens"
    strangelove.Actor["Dr. Strangelove"] = "Peter Sellers"
    strangelove.Actor["Grp. Capt. Lionel Mandrake"] = "Peter Sellers"
    strangelove.Actor["Pres. Merkin Muffley"] = "Peter Sellers"
    strangelove.Actor["Gen. Buck Turgidson"] = "George C. Scott"
    strangelove.Actor["Brig. Gen. Jack D. Ripper"] = "Sterling Hayden"
    strangelove.Oscars[0] = "Best Actor (Nomin.)"
    strangelove.Oscars[1] = "Best Adapted Screenplay (Nomin.)"
    strangelove.Oscars[2] = "Best Director (Nomin.)"
    strangelove.Oscars[3] = "Best Picture (Nomin.)"
    strangelove.Sequel = nil

## 12.5. Setting Variables with reflect.Value

上边提到的 reflection 仅仅有解释值(interpreted values)，本节讲如何修改它们。
一个变量定位了内存中的一块包含值的可访问地址，通过这个地址可以更新值。reflect.Values
类似，有些是可取址的(addressable)，有些不行。

    	x := 2                   // value type variable?
    	a := reflect.ValueOf(2)  // 2     int   no
    	b := reflect.ValueOf(x)  // 2     int   no
    	c := reflect.ValueOf(&x) //&x    *int   no
    	d := c.Elem()            // 2     int   yes

对任意变量，可用 reflect.ValueOf(&x).Elem() 获取一个可取址的 Value。
从一个可取址的 reflect.Value 恢复变量需要三步骤：

    	x := 2
    	d := reflect.ValueOf(&x).Elem()   // d refers to the variable x
    	px := d.Addr().Interface().(*int) // px := &x
    	*px = 3                           // x = 3
    	fmt.Println(x)                    // "3"

reflect.Value.Set 方法简化了这三步，注意类型要是可赋值的：

    	d.set(reflect.ValueOf(4))
    	fmt.Println(x) // "4"

## 12.7. Accessing Struct Field Tags

本节讲如何获取 struct 的 field tag。通过解析 http 请求参数来示例。

    package main

    import (
    	"fmt"
    	"log"
    	"net/http"
    	"reflect"
    	"strconv"
    	"strings"
    )

    // search implements the /search URL endpoint.
    func searchHandler(resp http.ResponseWriter, req *http.Request) {
    	var data struct {
    		Labels     []string `http:"l"`
    		MaxResults int      `http:"max"`
    		Exact      bool     `http:"x"`
    	}
    	data.MaxResults = 10 // set default
    	if err := Unpack(req, &data); err != nil {
    		http.Error(resp, err.Error(), http.StatusBadRequest) // 400
    		return
    	}
    	// ...rest of handler...
    	fmt.Fprintf(resp, "Search: %+v\n", data)
    }

    func main() {
    	http.HandleFunc("/search", searchHandler)
    	log.Fatal(http.ListenAndServe(":12345", nil))
    }

    func Unpack(req *http.Request, ptr interface{}) error {
    	if err := req.ParseForm(); err != nil {
    		return err
    	}
    	fields := make(map[string]reflect.Value)
    	v := reflect.ValueOf(ptr).Elem()
    	for i := 0; i < v.NumField(); i++ {
    		fieldInfo := v.Type().Field(i) // a reflect.StructField
    		tag := fieldInfo.Tag           // a reflect.StructTag
    		name := tag.Get("http")
    		if name == "" {
    			name = strings.ToLower(fieldInfo.Name)
    		}
    		fields[name] = v.Field(i)
    	}
    	fmt.Println(fields)

    	for name, values := range req.Form {
    		f := fields[name]
    		if !f.IsValid() {
    			continue // 忽略不能识别的 http 参数
    		}
    		for _, value := range values {
    			// fmt.Printf("%v\n", value)
    			if f.Kind() == reflect.Slice { // 同一个参数可能多次出现
    				elem := reflect.New(f.Type().Elem()).Elem()
    				if err := populate(elem, value); err != nil {
    					return fmt.Errorf("%s: %v", name, err)
    				}
    				f.Set(reflect.Append(f, elem))
    			} else {
    				if err := populate(f, value); err != nil {
    					return fmt.Errorf("%s: %v", name, err)
    				}
    			}
    		}
    	}
    	return nil
    }

    func populate(v reflect.Value, value string) error {
    	switch v.Kind() {
    	case reflect.String:
    		v.SetString(value)
    	case reflect.Int:
    		i, err := strconv.ParseInt(value, 10, 64)
    		if err != nil {
    			return err
    		}
    		v.SetInt(i)
    	case reflect.Bool:
    		b, err := strconv.ParseBool(value)
    		if err != nil {
    			return err
    		}
    		v.SetBool(b)
    		// ... 其他类型

    	default:
    		return fmt.Errorf("unsupported kinds %s", v.Type())
    	}
    	return nil
    }

## 12.8. Displaying the Methods of a Type

最后一个例子用 reflect.Type 任意类型值和并迭代其方法。

    package main

    import (
    	"fmt"
    	"reflect"
    	"strings"
    	"time"
    )

    func Print(x interface{}) {
    	v := reflect.ValueOf(x)
    	t := v.Type()
    	fmt.Printf("type %s\n", t)
    	for i := 0; i < v.NumMethod(); i++ {
    		methType := v.Method(i).Type()
    		fmt.Printf("func (%s) %s%s\n", t, t.Method(i).Name,
    			strings.TrimPrefix(methType.String(), "func"))
    	}
    }

    func main() {
    	Print(time.Hour)
    }

## 12.9. A Word of Caution

反射很强大，使用起来需要小心：

-   基于反射的代码比较脆弱，每个编译器报告类型错误的地方，都会有对应的反射使用不当的地方，有可能程序运行很长时间才会发现。
    最好的避免脆弱代码的方式就是使用反射的地方在包里完全封装，在包的 API 里尽量避免 reflect.Value
    而使用特定类型，把输入限定在合法的值。如果无法做到就使用动态检测。反射还会减弱静态分析工具的精确性和安全性。
-   使用反射太多的代码难以理解。应该注释期望的类型值
-   基于反射的函数执行速度更慢。最好在关键逻辑上避免反射。

# 13 Low-Level Programming

go的很多设计特性保证了用户不会错误使用 go，编辑期类型检查能消除很多类型错误。
无法静态检测的错误，比如数组越界访问、内存泄露等，go 的动态检测和垃圾回收终结了此类错误。go
屏蔽了很多内部细节的访问，没法探测聚合类型的内存结构，无法获取 goroutine 标识。综合这些方式减少类似低级 c
语言的很多奇怪的错误。
有时候为了提高性能，或者跟其他语言交互等，就无法仅仅用 go 代码满足需求，go 提供了由编译器实现的 unsafe
包，提供了一些操作内置语言特性的方式。

## 13.1. unsafe.Sizeof, Alignof, and Offsetof

-   Sizeof: unsafe.Sizeof function reports the size in bytes of the representation of its operand,which may be an expression of any type; the expression is not evaluated.
-   The unsafe.Alignof function reports the required alignment of its argument’s type
-   The unsafe.Offsetof function, whose operand must be a field selector x.f, computes the offset of field f relative to the start of its enclosing struct x, accounting for holes, if any.

## 13.2. unsafe.Pointer

Most pointer types are written \*T, meaning ‘‘a pointer to a variable of type T.’’ The unsafe.Pointer type is a special kind of pointer that can hold the address of any variable.

## 13.3 Example: Deep Equivalence

reflect.DeepEqual 对于基本类型使用 == 比较，对于复合类型会递归遍历它们比较对应的元素。经常用在测试里，但是是实现有差别。
比如 DeepEqual 实现认为 nil map 和空 map 是不等的：

    	var a, b []string = nil, []string{}
    	fmt.Println(reflect.DeepEqual(a, b)) // "false"
    	var c, d map[string]int = nil, make(map[string]int)
    	fmt.Println(reflect.DeepEqual(c, d)) // "false"

下边实现一个类似的，但是认为 nil map 和空 map 相等。

    package main

    import (
    	"reflect"
    	"unsafe"
    )

    type comparison struct {
    	x, y unsafe.Pointer
    	t    reflect.Type
    }

    func equal(x, y reflect.Value, seen map[comparison]bool) bool {
    	if !x.IsValid() || !y.IsValid() {
    		return x.IsValid() == y.IsValid()
    	}
    	if x.Type() != y.Type() {
    		return false
    	}
    	//循环检测
    	if x.CanAddr() && y.CanAddr() {
    		xptr := unsafe.Pointer(x.UnsafeAddr())
    		yptr := unsafe.Pointer(y.UnsafeAddr())
    		if xptr == yptr {
    			return true // 同一个引用
    		}
    		c := comparison{xptr, yptr, x.Type()}
    		if seen[c] { // 如果都是 array，x and x[0] have the same address,所以需要区分 x和y  x[0]和y[0] 是否比较过
    			return true
    		}
    		seen[c] = true
    	}

    	switch x.Kind() {
    	case reflect.Bool:
    		return x.Bool() == y.Bool()
    	case reflect.String:
    		return x.String() == y.String()
    	case reflect.Chan, reflect.UnsafePointer, reflect.Func:
    		return x.Pointer() == y.Pointer()
    	case reflect.Ptr, reflect.Interface:
    		return equal(x.Elem(), y.Elem(), seen)

    	case reflect.Array, reflect.Slice:
    		if x.Len() != y.Len() {
    			return false
    		}
    		for i := 0; i <= x.Len(); i++ {
    			if !equal(x.Index(i), y.Index(i), seen) {
    				return false
    			}
    		}
    		return true
    	}
    	panic("unreachable")

    }

    func Equal(x, y interface{}) bool {
    	seen := make(map[comparison]bool)
    	return equal(reflect.ValueOf(x), reflect.ValueOf(y), seen)
    }

    func main() {
    	fmt.Println(Equal([]int{1, 2, 3}, []int{1, 2, 3}))
    	fmt.Println(Equal([]string{"foo"}, []string{"bar"}))
    	fmt.Println(Equal([]string(nil), []string{}))
    	fmt.Println(Equal(map[string]int(nil), map[string]int{}))
    }

## 13.4. Calling C Code with cgo

cgo, a tool that creates Go bindings for C functions. Such tools are called foreign-function interfaces (FFIs), and cgo is not the only one for Go programs.
<https://golang.org/cmd/cgo>

## 13.5. Another Word of Caution

Most programmers will never need to use unsafe at all. Nevertheless, there will occasionally be situations where some critical piece of code can be best written using unsafe

# Debug

这本书居然没有讲 debug，从网上搜了下调试工具，然后试用了一把 delve，感觉还不错。目前看到两种用得比较多是 gdb 和 delve，感觉 gdb
有点原始，可以用下 delve。用起来和 Python 的 pdb 和 ipdb 差不多，都是 gdb 风格的命令，平常笔者调试 python
代码基本都是用的 ipdb。

[debugging-with-delve](https://blog.gopheracademy.com/advent-2015/debugging-with-delve/)

<https://github.com/derekparker/delve>
