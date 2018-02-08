# 10 Packages and the Go Tool

## 10.1 Introduction
包系统都是通过把相关特性组织成单元，使得大型程序的设计和维护更易理解和修改，和程序的其他包隔离。
模块化使得包更同意被分享、重用。go 里每个 package
都定义了一组包含实体的命名空间，防止和其他包冲突。并且通过命名来区分可见性。

## 10.2 Import Paths
每个包都被一个独一无二的字符串标识，叫做导入路径(import path)。对于你想分享或者发布的包，导入路径必须是全局唯一的。
为了防止冲突，凡事非官方的包都应该包含域名或者组织的前缀。
```
import (
	"fmt"
	"math"

	"golang.org/x/net/html"
	"github.com/go-sql-driver/mysql"
)
```

## 10.3 The Pacakge Decalration

包声明需要在每个 go 源文件头声明，主要是为了当包被导入的时候决定默认标识符。
```
import (
	"fmt"
	"math/rand"
)
```
比如这里我们可以通过 rand.Int 访问。常规上包名是导入路径的最后一段（后边会讨论重名冲突问题），但是有三种例外：
- 一种是像 main 这种直接被 buildo 成可执行文件的。
- 一些目录里的文件可能包含 `_test` 后缀（一般是 `_test.go`测试文件）
- 对于有版本号的 `gopkg.in/yaml.v2` 直接用 yaml 就行


## 10.4 Import Decalration
导入可以每行导入一个或者统一放在括号里（笔者用的 vim-go 设置了自动导入，保存文件的时候会自动处理导入
pacakge，其他编辑器或者 IDE 都有类似的设置）
如果有包名冲突了，可以用 rename 的方式防止冲突。

```
import (
	"crypto/rand"
	mrand "math/rand"   // 别名防止冲突，也可以用来简化一些比较长的 package name
)
```

## 10.5 Blank Imports
如果导入包但是不用是会报错的，但是有些场景比如使用是在 init 里的，这个时候就需要使用 empty import 

```
import _ "image/png"   // 其实就是 rename 成了 _
```

## 10.6 Pacakges and Naming
讲下包的命名:
- 尽量短小，但不要难以理解
- 具有描述性但不要有歧义
- 一般用单数形式，但是内置的 bytes, errors, strings 是防止和关键字冲突
- 避免使用已经具有其他含义的名称，比如 temp（一般理解为临时文件相关），我们的转换温度的包改成 tempconv

讲下包成员的命名：
- 考虑到和包结合使用时候的含义。 bytes.Equal, flag.Int, http.Get, json.Marshal
- 保持精简，不要和包名重复开头。比如 strings.Index not strings.StringIndex

## 10.7 The Go Tool

go tool 提供了下载、查询、格式化、构建、测试、安装包等功能。直接输入 go 回车就可以看到列出的工具，使用 go help command
可以查看具体使用方法。

### 10.7.1 Workspace Organization
大部分用户唯一需要设置的就是 GOPATH 环境变量，指定了工作区的根目录，切换工作区只需要改变 GOPATH。
GOPATH 里有仨子目录:
- src: 保存源文件。相对于 `$GOPATH/src`就是它的导入路径
- pkg：build tools 存储编译的包
- bin: 保存可执行文件

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
```
net/http
net/http/internal/chunked
net/http/httputil
net/url<Paste>
```
`net/httpinternal/chunked`只能被 `net/http/httputil net/http`导入

### 10.7.6 Querying Pacakges
go list 可以用来查找包
