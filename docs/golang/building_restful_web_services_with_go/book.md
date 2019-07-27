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


# 3. Working with Middleware and RPC


### What is middleware?

When a piece of code needs to be executed for every request or subset of HTTP requests.

### Creating a basic middleware

```
// custommiddleware.go
package main

func middleware(handler http.handler) http.handler {
	return http.handlerfunc(func(w http.responsewriter, r *http.request) {
		fmt.println("executing middleware before request phase!")
		//pass control back to handler
		handler.servehttp(w, r)
		fmt.println("executing middleware after response phase!")
	})
}

func mainlogic(w http.responsewriter, r *http.request) {
	// business logic here
	fmt.println("executing mainhandler")
	w.write([]byte("ok"))
}

func main() {
	manlogichandler := http.handlerfunc(mainlogic)
	http.handler("/", middleware(manlogichandler))
	http.listenandserve(":8000", nil)
}
```

### Multiple middleware and chaining


```
// cityAPI.go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type city struct {
	Name string
	Area uint64
}

func mainLogic(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var tempCity city
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&tempCity)
		if err != nil {
			panic(err)
		}
		defer r.Body.Close()
		// creat logic
		fmt.Printf("%s %d", tempCity.Name, tempCity.Area)
		// tell everything is fine
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("201 - Created"))
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("405 - Method Not Allowed"))
	}
}

func main() {
	http.HandleFunc("/city", mainLogic)
	http.ListenAndServe(":8000", nil)
}
```

抽象出来俩middleware:

```
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

type city struct {
	name string
	area uint64
}

// middleware to check content type as json
func filtercontenttype(handler http.handler) http.handler {
	return http.handlerfunc(func(w http.responsewriter, r *http.request) {
		log.println("currently in the check content type middleware")
		// filtering request by mime type
		if r.header.get("content-type") != "application/json" {
			w.writeheader(http.statusunsupportedmediatype)
			w.write([]byte("415 - unsupported media type. please send json"))
			return
		}
		handler.servehttp(w, r)
	})
}

// middle to add server timestamp for response cookie
func setservertimecookie(handler http.handler) http.handler {
	return http.handlerfunc(func(w http.responsewriter, r *http.request) {
		handler.servehttp(w, r)
		cookie := http.cookie{name: "server-time(utc)", value: strconv.formatint(time.now().unix(), 10)}
		http.setcookie(w, &cookie)
		log.println("currently in the set server time middleware")
	})
}

func mainlogic(w http.responsewriter, r *http.request) {
	if r.method == "post" {
		var tempcity city
		decoder := json.newdecoder(r.body)
		err := decoder.decode(&tempcity)
		if err != nil {
			panic(err)
		}
		defer r.body.close()
		// creat logic
		fmt.printf("%s %d", tempcity.name, tempcity.area)
		// tell everything is fine
		w.writeheader(http.statusok)
		w.write([]byte("201 - created"))
	} else {
		w.writeheader(http.statusmethodnotallowed)
		w.write([]byte("405 - method not allowed"))
	}
}

func main() {
	mainlogichnadler := http.handlerfunc(mainlogic)
	http.handler("/city", filtercontenttype(setservertimecookie(mainlogichnadler)))
	http.listenandserve(":8000", nil)
}
```

### Paniless middleware chaining with Alice

```
//go get github.com/justinas/alice

func main() {
    mainLogicHandler := http.HandlerFunc(mainLogic)
    chain := alice.New(filterContentType, setServerTimeCookie).Then(mainLogicHandler)
    http.Handle("/city", chain)
    http.ListenAndServe(":8000", nil)
}
```

### Using Gorilla's Handlers middleware for Logging

- LoggingHandler
- CompressingHandler: for zipping the responses
- RecoveryHandler: for recovering from unexpected panics

### What is RPC?

"net/rpc"

### JSON RPC using Gorilla RPC


# 4. Simplifying RESTful Services with Popular Go Frameworks

### go-restful, a framework for REST API creation


```
package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/emicklei/go-restful"
)

// brew install sqlite3
// go get github.com/emicklei/go-restful
// a simple ping server echoes the server time back to the client
func main() {
	webservice := new(restful.WebService)
	webservice.Route(webservice.GET("/ping").To(pingTime))
	restful.Add(webservice)
	http.ListenAndServe(":8000", nil)
}

func pingTime(req *restful.Request, resp *restful.Response) {
	io.WriteString(resp, fmt.Sprintf("%s", time.Now()))
}

// curl -X GET "http://localhost:8000/ping"
```

### CRUD and SQLite3 basics

```
// sqliteFundamentals.go
package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// go get github.com/mattn/go-sqlite3

type Book struct {
	id     int
	name   string
	author string
}

func main() {
	db, err := sql.Open("sqlite3", "./books.db")
	log.Println(db)
	if err != nil {
		log.Println(err)
	}
	//create table
	statement, err := db.Prepare("create table if not exists books(id integer primary key, isbn integer, author(varchar(64), name varchar(64) NULL)")
	if err != nil {
		log.Println("Error in creating table")
	} else {
		log.Println("success created table books")
	}
	statement.Exec()
	// Create
	statement, _ = db.Prepare("INSERT INTO books (name, author, isbn) VALUES (?, ?, ?)")
	statement.Exec("A Tale of Two Cities", "Charles Dickens", 140430547)
	log.Println("Inserted the book into database!")
	// Read
	rows, _ := db.Query("SELECT id, name, author FROM books")
	var tempBook Book
	for rows.Next() {
		rows.Scan(&tempBook.id, &tempBook.name, &tempBook.author)
		log.Printf("ID:%d, Book:%s, Author:%s\n", tempBook.id,
			tempBook.name, tempBook.author)
	}
	// Update
	statement, _ = db.Prepare("update books set name=? where id=?")
	statement.Exec("The Tale of Two Cities", 1)
	log.Println("Successfully updated the book in database!")
	//Delete
	statement, _ = db.Prepare("delete from books where id=?")
	statement.Exec(1)
	log.Println("Successfully deleted the book in database!")
}

// 略坑啊：
// This program runs on Windows and Linux without any problem. In Go versions less than 1.8.1, you may see problems on macOS X such as Signal Killed. This is because of the Xcode version; please keep this in mind.
```


### Building a Metro Rail APi with go-restful

// go get github.com/emicklei/go-restful


### Building RESTful APIs with the Gin framework

```
// go get github.com/gin-gonic/gin
// ginBasic.go
package main

import (
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/pingTime", func(c *gin.Context) {
		c.JSON(200, gin.H{"serverTime": time.Now().UTC()})
	})
	r.Run(":8000")
}
```


