# 1. Getting started with REST API Development

Live reloading the application with supervisord and gulp

# 2. Handling Routing for Rest Srevices

### ServeMux a basic router in Go

`go get github.com/julienschmidt/httprouter`


```
// execService.go

// Package main  provides ...
package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os/exec"

	"github.com/julienschmidt/httprouter"
)

func getCommandOutput(command string, arguments ...string) string {
	cmd := exec.Command(command, arguments...) //unpack
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Start()
	if err != nil {
		log.Fatal(fmt.Sprint(err) + ": " + stderr.String())
	}
	err = cmd.Wait()
	if err != nil {
		log.Fatal(fmt.Sprint(err) + ": " + stderr.String())
	}
	return out.String()
}

func goVersion(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	fmt.Fprintf(w, getCommandOutput("/usr/local/bin/go", "version"))
}

func getFileContent(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	fmt.Fprintf(w, getCommandOutput("/bin/cat", params.ByName("name")))
}

func main() {
	router := httprouter.New()
	router.GET("/api/v1/go-version", goVersion)
	router.GET("/api/v1/show-file/:name", getFileContent)
	log.Fatal(http.ListenAndServe(":8000", router))
}
```

### Building the simple static file server

```
package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func main() {
	router := httprouter.New()
	router.ServeFiles("/static/*filepath", http.Dir("/users/naren/static"))
	log.Fatal(http.ListenAndServe(":8000", router))
}
```

### Gorilla Mux, a powerful HTTP router

- Path-based matching
- Query-based matching
- Domain-based matching
- Sub-domain based matching
- Reverse URL generation
