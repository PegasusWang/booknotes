# go 内置 net/http 代码阅读

看一下内置的 http server 如何实现的

```go
package main

import (
    "net/http"
    "log"
)

func myHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello there!\n")
}

func main(){
    http.HandleFunc("/", myHandler)		//	设置访问路由
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```
