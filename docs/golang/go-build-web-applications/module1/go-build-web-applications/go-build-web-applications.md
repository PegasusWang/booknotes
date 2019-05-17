# 构建 web


	package main

	import (
		"database/sql"

		_ "github.com/go-sql-driver/mysql"

		"fmt"
		"log"
		"net/http"
		"os"

		"github.com/gorilla/mux"
	)

	var database *sql.DB

	const (
		PORT    = ":9000"
		DBHost  = "localhost"
		DBPort  = ":3306"
		DBUser  = "wnn"
		DBPass  = "wnnwnn"
		DBDbase = "test"
	)

	type Page struct {
		Title   string
		Content string
		Date    string
	}

	func pageHandler(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r) //解析查询参数为 map
		fmt.Println("%v", vars)
		pageID := vars["id"]
		fileName := "files/" + pageID + ".html"
		_, err := os.Stat(fileName)
		if err != nil {
			fileName = "files/404.html"
		}
		http.ServeFile(w, r, fileName)
	}

	func ServePage(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		pageID := vars["id"]
		thisPage := Page{}
		fmt.Println(pageID)
		err := database.QueryRow("select page_title,page_content,page_date from pages where id=?", pageID).Scan(&thisPage.Title, &thisPage.Content, &thisPage.Date)
		if err != nil {
			log.Println("Couldn't get page:" + pageID)
			log.Println(err.Error)
		}
		html := `<html><head><title>` + thisPage.Title +
			`</title></head><body><h1>` + thisPage.Title + `</h1><div>` +
			thisPage.Content + `</div></body></html>`
		fmt.Fprintln(w, html)
	}

	func ServePageByGUID(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		pageGUID := vars["guid"]
		thisPage := Page{}
		fmt.Println(pageGUID)
		err := database.QueryRow("select page_title,page_content,page_date from pages where page_guid=?", pageGUID).Scan(&thisPage.Title, &thisPage.Content, &thisPage.Date)
		if err != nil {
			log.Println("Couldn't get page:" + pageGUID)
			http.Error(w, http.StatusText(404), http.StatusNotFound)
			log.Println(err.Error)
			return // 这里应该需要 return，书里没有
		}
		html := `<html><head><title>` + thisPage.Title +
			`</title></head><body><h1>` + thisPage.Title + `</h1><div>` +
			thisPage.Content + `</div></body></html>`
		fmt.Fprintln(w, html)
	}

	func main() {
		dbConn := fmt.Sprintf("%s:%s@/%s", DBUser, DBPass, DBDbase)
		fmt.Println(dbConn)
		db, err := sql.Open("mysql", dbConn)
		if err != nil {
			log.Println("Couldn't connect to: " + DBDbase)
			log.Println(err.Error)
		}
		database = db

		rtr := mux.NewRouter()
		// rtr.HandleFunc("/pages/{id:[0-9]+}", ServePage)
		rtr.HandleFunc("/pages/{guid:[a-z0-9A\\-]+}", ServePageByGUID)
		http.Handle("/", rtr)
		http.ListenAndServe(PORT, nil)
	}


# 1 An Introduction to Concurrency in Go

注意引用传递在 defer 中的坑。下边的例子输出0而不是100

```go
package main

import "fmt"

func main() {
	a := new(int)

	defer fmt.Println(*a)

	for i := 0; i < 100; i++ {
		*a++
	}
}
```

### channel 的使用。 CSP(Communicating Sequential Processes)

```go
package main

import (
	"fmt"
	"strings"
	"sync"
)

var finalString string
var initString string
var stringLen int

func addToFinalString(letterChan chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	letter := <-letterChan
	finalString += letter
}

func upper(letterChan chan string, letter string, wg *sync.WaitGroup) {
	defer wg.Done()
	upperLetter := strings.ToUpper(letter)
	letterChan <- upperLetter
}

func main() {
	var wg sync.WaitGroup
	initString = "gebilaowang"
	initBytes := []byte(initString)
	stringLen = len(initBytes)
	var letterChan chan string = make(chan string)
	for i := 0; i < stringLen; i++ {
		wg.Add(2)

		go upper(letterChan, string(initBytes[i]), &wg)
		go addToFinalString(letterChan, &wg)

		wg.Wait()
	}
	fmt.Println(finalString)
}
```

### select 的使用

语法类似 switch 。
select reacts to actions and communication across a channel.

```go
switch {
	case 'x':
	case 'y':
}
select {
	case <- channelA:
	case <- channelB:
}
```

select 会 block直到有数据发送给channel。否则会 deadlocks
如果同时 receive，go 会无法预期地执行其中一个

```go
package main

import (
	"fmt"
	"strings"
	"sync"
)

var initialString string
var initialBytes []byte
var stringLength int
var finalString string
var lettersProcessed int
var wg sync.WaitGroup
var applicationStatus bool

func getLetters(gQ chan string) {

	for i := range initialBytes {
		gQ <- string(initialBytes[i])
	}

}

func capitalizeLetters(gQ chan string, sQ chan string) {

	for {
		if lettersProcessed >= stringLength {
			applicationStatus = false
			break
		}
		select {
		case letter := <-gQ:
			capitalLetter := strings.ToUpper(letter)
			finalString += capitalLetter
			lettersProcessed++
		}
	}
}

func main() {

	applicationStatus = true

	getQueue := make(chan string)
	stackQueue := make(chan string)

	initialString = "Four score and seven years ago our fathers brought forth on this continent, a new nation, conceived in Liberty, and dedicated to the proposition that all men are created equal."
	initialBytes = []byte(initialString)
	stringLength = len(initialString)
	lettersProcessed = 0

	fmt.Println("Let's start capitalizing")

	go getLetters(getQueue)
	capitalizeLetters(getQueue, stackQueue)

	close(getQueue)
	close(stackQueue)

	for {

		if applicationStatus == false {
			fmt.Println("Done")
			fmt.Println(finalString)
			break
		}

	}
}
```


# 2 Understanding the Concurrency Model

go get github.com/ajstarks/svgo


Go/Csp vs Actor model:

```
// actor model
a = new Actor
b = new Actor
a -> b("message")

// csp
a = new Actor
b = new Actor
c = new Channel
a -> c("sending something")
b <- c("receiving something")
```

go里简单的多态实现(polymorphism):

```
package main

import "fmt"

type intInterface struct{}
type stringInterface struct{}

func (num intInterface) Add(a, b int) int {
	return a + b
}
func (s stringInterface) Add(a, b string) string {
	return a + b
}

func main() {
	number := new(intInterface)
	fmt.Println(number.Add(1, 2))

	text := new(stringInterface)
	fmt.Println(text.Add("hello", " world"))

}
```

use sync and mutexes to lock data:

```
package main

import (
	"fmt"
	"runtime"
	"sync"
)

func main() {
	runtime.GOMAXPROCS(4)
	current := 0
	iters := 100
	wg := new(sync.WaitGroup)
	wg.Add(iters)
	mutex := new(sync.Mutex)

	for i := 0; i < iters; i++ {
		go func() {
			mutex.Lock()
			fmt.Println(current)
			current++
			mutex.Unlock()
			fmt.Println(current)
			wg.Done()
		}()
	}
	wg.Wait()
}
```


# 3 Developing a Concurrent Strategy

### Use -race flag when testing, building or running.

`go run -race race-test.go`

### using mutual exclusions:

```go
package main

import (
	"fmt"
	"runtime"
	"sync"
)

func main() {
	runtime.gomaxprocs(4)
	current := 0
	iters := 100
	wg := new(sync.waitgroup)
	wg.add(iters)
	mutex := new(sync.mutex)

	for i := 0; i < iters; i++ {
		go func() {
			mutex.lock()
			fmt.println(current)
			current++
			mutex.unlock()
			fmt.println(current)
			wg.done()
		}()
	}
	wg.wait()
}
```

### Exploring timeouts:

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	ourCh := make(chan string, 1)
	go func() {
	}()

	select {
	case <-time.After(10 * time.Second):
		fmt.Println("Enough's enough")
		close(ourCh)
	}
}
```

### Synchronizing our concurrent operations


# 4 Data Integrity in an Application

### Getting depper with mutexes and sync

- RWMutex: miltireader, single-writer lock(feaquent read , infrequent write, cannot dirty read)

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

var currentTime time.Time
var rwLock sync.RWMutex

func updateTime() {
	rwLock.RLock()
	currentTime = time.Now()
	time.Sleep(5 * time.Second)
	rwLock.RUnlock()
}

func main() {

	var wg sync.WaitGroup

	currentTime = time.Now()
	timer := time.NewTicker(2 * time.Second)
	writeTimer := time.NewTicker(10 * time.Second)
	endTimer := make(chan bool)

	wg.Add(1)
	go func() {

		for {
			select {
			case <-timer.C:
				fmt.Println(currentTime.String())
			case <-writeTimer.C:
				updateTime()
			case <-endTimer:
				timer.Stop()
				return
			}

		}

	}()

	wg.Wait()
	fmt.Println(currentTime.String())
}
```

### Working with files

```go
package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"sync"
)

//

func writeFile() {
	fmt.Println("um")
	for i := 0; i < 10; i++ {
		rwLock.RLock()
		ioutil.WriteFile("test.txt", []byte(strconv.FormatInt(int64(i), 10)), 0x777)
		rwLock.RUnlock()
	}

	writer <- true

}

var writer chan bool
var rwLock sync.RWMutex

func main() {

	writer = make(chan bool)

	go writeFile()

	<-writer
	fmt.Println("Done!")
}
```

### Getting low-implementing C

```go
package main

/*
	#include <stdio.h>
	#include <stdlib.h>

	void Coutput (char* str) {
		printf("%s",str);
	}
*/
import "C"
import "unsafe"

func main() {
	v := C.CString("Don't Forget My Memory Is Not Visible To Go!")
	C.Coutput(v)
	C.free(unsafe.Pointer(v))
}
```

### Distributed Go

- distributed lock
- consistent hash

Common consistency models:

- Distribued shared memory (DSM)
- First-in-first-out - PRAM
- master-slave model

producer-consumer problem

```go
package main

import "fmt"

var comm = make(chan bool)
var done = make(chan bool)

func producer() {
	for i := 0; i < 10; i++ {
		comm <- true
	}
	done <- true
}

func consumer() {
	for {
		communication := <-comm
		fmt.Println("Communiction from producer received!", communication)
	}
}
func main() {
	go producer()
	go consumer()
	<-done
	fmt.Println("All Done!")
}
```

### Go Circuit

https://github.com/gocircuit/circuit
need Zookpeer

跨进程 channel


# 5. Locks, Blocks, and Better Channels

### Understanding Blocking Methods in GO

- 1.a listening, waiting channel
- 2.the select statement in a loop
- 3.network connections and reads

```go
<- myChannel  // same way as the following code snippet

select {
case mc := <- myChannel:
	// do something
}
```


Cleaning up goroutines:Any channel that is left waiting and/or left receiving will result in a deadlock.

### Pprof

runtime.ppof package

```go
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"runtime/pprof"
)

var profile = flag.String("cpuprofile", "", "output pprof data to file")

func generateString(length int, seed *rand.Rand, chHater chan string) string {
	bytes := make([]byte, length)
	for i := 0; i < length; i++ {
		bytes[i] = byte(rand.Int())
	}
	chHater <- string(bytes[:length])
	return string(bytes[:length])
}

func generateChannel() <-chan int {
	ch := make(chan int)
	return ch
}

func main() {
	iterations := 99999
	goodbye := make(chan bool, iterations)
	channelThatHatesLetters := make(chan string)

	runtime.GOMAXPROCS(2)
	flag.Parse()
	if *profile != "" {
		flag, err := os.Create(*profile)
		if err != nil {
			fmt.Println("Could not create profile", err)
		}
		pprof.StartCPUProfile(flag)
		defer pprof.StopCPUProfile()

	}

	seed := rand.New(rand.NewSource(19))

	initString := ""

	for i := 0; i < iterations; i++ {
		go func() {
			initString = generateString(300, seed, channelThatHatesLetters)
			goodbye <- true
		}()

	}
	select {
	case <-channelThatHatesLetters:

	}
	<-goodbye

	fmt.Println(initString)
}
```

### Handling deadlocks and errors


```package maingo

import "os"

func main() {
	panic("Oh no, forgot to write a program")
	os.Exit(1)
}
```


# 6. C10K - A non-blocking web server in GO



# 10 Advanced Concurrency and Best Practices 

使用 tomb(need go get) 可以对 goroutine 更细粒度控制，比如超时 kill

```go
package main

import (
	"fmt"
	"time"
)

func main() {

	myChan := make(chan int)

	go func() {
		time.Sleep(6 * time.Second)
		myChan <- 1
	}()

	for {
		select {
			case <-time.After(5 * time.Second):
				fmt.Println("This took too long!")
				return
			case <-myChan:
				fmt.Println("Too little, too late")
		}
	}
}
```
