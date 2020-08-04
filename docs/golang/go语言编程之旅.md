# 1. 命令行应用

### 1.1 子命令的使用

```go
package main

import (
	"flag"
	"log"
)

var name string

// go run main.go go -name=laowang
func main() {
	flag.Parse()

	goCmd := flag.NewFlagSet("go", flag.ExitOnError)
	goCmd.StringVar(&name, "name", "Go", "go help")
	phpCmd := flag.NewFlagSet("php", flag.ExitOnError)
	phpCmd.StringVar(&name, "n", "php", "php help")

	args := flag.Args()
	switch args[0] {
	case "go":
		_ = goCmd.Parse(args[1:])
	case "php":
		_ = phpCmd.Parse(args[1:])
	}
	log.Printf("name: %s", name)
}
```

### 1.2

使用 cobra 构建命令行 app： https://github.com/spf13/cobra

```go
// 使用 cobra 完成单词转换程序
package main

import (
	"log"
	"strings"

	"github.com/spf13/cobra"
)

const (
	MODE_UPPER = iota + 1
	MODE_LOWER
)

var wordCmd = &cobra.Command{
	Use:   "word",
	Short: "单词格式转换",
	Long:  "支持多种格式转换",
	Run: func(cmd *cobra.Command, args []string) {
		var content string
		switch mode {
		case MODE_UPPER:
			content = ToUpper(str)
		case MODE_LOWER:
			content = ToLower(str)
		default:
			log.Fatalf("暂不支持格式")
		}
		log.Printf("输出结果: %s", content)
	},
}

func ToUpper(s string) string {
	return strings.ToUpper(s)
}

func ToLower(s string) string {
	return strings.ToLower(s)
}

var desc = strings.Join([]string{
	"该命令支持单词转换，格式如下:",
	"1: 大写",
	"2: 小写",
}, "\n")

var str string // 输入
var mode int8

func init() {
	wordCmd.Flags().StringVarP(&str, "str", "s", "", "请输入单词内容")
	wordCmd.Flags().Int8VarP(&mode, "mode", "m", 0, "请输入单词转换模式")
}

// go run main.go word -s=laowang -m=1
func main() {
	err := wordCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
```
