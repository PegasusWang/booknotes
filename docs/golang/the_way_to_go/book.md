# 11 Interfaces and reflection

An Interface define a set of methods. And cannot contain variables.

```go
type Namer interface {
	Method1(param_list) return_type
	Method2(param_list) return_type
}
```

## 11.3 type assertion

```go
if v, ok := varI.(T); ok {
//
}
```

Value-received methods can be called with pointer values because they can be dereference first.

```go
package sort

type Sorter interface {
	Len() int
	Less(i, jint) bool
	Swap(i, j int)
}

fucn Sort(data Sorter) {
	for pass := 1; pass <data.Len(); pass ++ {
		for i :=0; i < data.Len()-pass; i++ {
			if data.Less(i+1,i) {
				data.Swap(i,i+1)
			}
		}
	}
}

func IsSorted(data Sorter) bool {
	n := data.Len()
	for i := n-1; i >0; i-- {
		if data.Less(i, i-1){
			return false
		}
	}
	return true
}

// convenience types for common cases

type IntArray []int
func (p IntArray) Len() int {return len(p)}
func (p IntArray) Less(i,j int) bool {return p[i] < p[j]}
func (p IntArray) Swap(i,j int) {p[i],p[j] = p[j],p[i]}
```


### 11.9.2 contsucting an array of a general type or with variables of different types

```go
type Element interface{}

type Vector struct {
	a []Element
}

func (p *Vector) At(i int) Element {
	return p.a[i]
}

func (p *Vector) Set(i int, Element e) {
	p.a[i] = e
}
```

tree struct:

```go
package main

import "fmt"

type Node struct {
	le   *Node
	data interface{}
	ri   *Node
}

func NewNode(left, right *Node) *Node {
	return &Node{left, nil, right}
}

func (n *Node) SetData(data interface{}) {
	n.data = data
}

func main() {
	root := NewNode(nil, nil)
	root.SetData("root node")
	a := NewNode(nil, nil)
	a.SetData("left node")
	b := NewNode(nil, nil)
	b.SetData("right node")
	root.le = a
	root.ri = b
	fmt.Printf("%v\n", root)
}
```


## 11.10 The reflect package


## 11.14 Structs coolections and higher order functions

```go
package main

type Any interface{}

type Car struct {
	Model        string
	Manufacturer string
	BuildYear int
}

type Cars []*Car

func (cs Cars) Map(f func(car *Car) Any) []Any {
	result := make([]Any, 0)
	ix := 0
	cs.Process(func(c *Car) {
		result[ix] = f(c)
		ix++
	})
	return result
}

allNewBMWs := allCars.FindAll(func(car *Car) bool {
	return (car.Manufacturer == "BMW") && (car.BuildYear > 2010)
})
```

# 12 Reading and writing

```go
// read input from console:
package main

import "fmt"

var (
	firstName, lastName string
)

func main() {
	fmt.Scanln(*firstName, *lastName)
	fmt.Println("Hi %s %s\n", &firstName, &lastName)
}

// or use bufio.Reader
```


# 13 Error-handling and Testing

no try/catch, defer-panic-and-recover mechanism.

```go
type error interface {
	Error() string
}
// to stop an error-state program, use os.Exit(1)


// define new errors
err := errros.New("math-square root of negative number")

// custome error field
type PathError struct {
	Op   string
	Path string
	Err  error
}
func ( e *PathError) String() string {
	return e.Op + " " + e.Path = ": " + e.Err.Error()
}
// making errror-object with fmt
if f < 0 {
	return 0, fmt.Errorf("math: square root of negative number %g", f)
}
```


## 13.2 Run-time exceptions and panic

if panic is called from a nested function, immediately stops the execution of the current function,
all defer statement are guaranteed to execute and the control is given to the function caller, which receives this call
to panic.


## 13.3 Recover

recover is only useful when called inside a deferred function: it then retrieves the error value passed through the call
of panic; when used in normal execution a call to recover will return nil and have no other effects.

like catch in java.

```go
func protect(g func()) {
	defer func() {
		log.Println("done")
		if err := recover(); err != nil {
			log.Printf("run time panic : %v", err)
		}
	}()
	log.Println("start")
	g()
}


package main

import "fmt"

func badCall() {
	panic("bad end")
}

func test() {
	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("Panicking %s\r\n", e)
		}
	}()
	badCall()
	fmt.Printf("After bad call\r\n")
}

func main() {
	fmt.Printf("calling test\r\n")
	test()
	fmt.Printf("calling test end\r\n")
}
```

## 13.4 Error-handling and panicking in a custom package


best practice:

- always recover from panic in your package: no explicit panic() should be allowed to cross a package boundary
- return errors as error values to the callers of your package.


## 13.5 An Error-handling scheme with closures

```go
func f(a type1, b type2)

fType1 = func f(a type1, b type2)


func errorHandler(fn fType1) fType1 {
	return func(a type1, b type2) {
		defer func() {
			if e, ok := recover().(error); ok {
				log.Printf("run time panic:%v", err)
			}
		}()
		fn(a,b)
	}
}

```

## 13.6 Starting an external command or program

os.StartProcess
exec.Command(name string, arg ...string)

## 13.7 Testing and benckmarking in GO



# 14 Goroutines and channels

## 14.1 Concurrency, parallelism and goroutines


CSP(Communicating Sequentiadl Processes)

parallelism is the ability to make things run quickly by using multiple processors.

An experiential rule of thumb seems to be that for n cores setting GOMAXPROCS to n-1 yields the best the performance,
and the following should also be followed: number of goroutines > 1 + GOMAXPROCS > 1

If we do not wai int main(), the program stop the goroutines die with it.
When the func main() returns, the program exits, it does not wait for other (non-main) goroutines to complete.
The logic of your code must be independent of the order in which goroutines are invoked.

### 14.1.5

difference bewteen goroutines and coroutines(c# and python):

- goroutines imply parallelism, coroutines in general do not
- goroutines communicate via channels; coroutines communicate via yield and resume operations

## 14.2 Channels for communication between goroutines

Using shared variables is not discouranged.
Only one goroutine has access to a  data item at any given time: so data races cannot occur, by design. Channels a firstclass objects.

```
var identifier chan datatype // uniniitialized channel is nil
var cha1 chan string // reference type
ch1 = make(chan string)
```

Channel send and receive operations are aotmic, they always complete without interruption.

### 14.2.3 Blocking of channels

default communication is synchronous and unbuffered: sends do not complete until there is a receiver to accept the value. So channel send/receive block until the other side is ready.

- A send opeartion on a channel blocks until a receiver is available. for the same channel.
- A receive operation for a channel blocks until a sender is available for the same channel.


```
package main

import (
	"fmt"
	"time"
)

func main() {
	ch1 := make(chan int)
	go pump(ch1)
	// fmt.Println(<-ch1)
	go suck(ch1)
	time.Sleep(1e9)
}

func pump(ch chan int) {
	for i := 0; ; i++ {
		ch <- i
	}
}

func suck(ch chan int) {
	for {
		fmt.Println(<-ch)
	}
}
```

### 14.2.5 Asynchronous channels-making a channel with a buffer

An unbuffered channel can only contain 1 item and is for that reason somethimes too restrictive.

```
buf := 100
ch1 := make(chan string, buf) //buf is the number of elements the channel can hold
```

sending to a bufferd channel will not blocked unless buffer is full, reading not blocked unless buffer is empty.
If the capacity is greater than 0, channel is asychronous

Semaphore pattern


```
	ch := make(chan int)
	go func() {
		// do something
		ch <- 1
	}()
	doSomethingElseForAWhile()
	<-ch // wait for goroutine to finish. discard send value
```

### 14.2.9 Implementing a semaphore using a buffered channel

```
type Empty interface{}
type semaphore chan Empty


sem = make(semaphore, N)

// acauire n resourdces
func (s semaphore) P(n int) {
	e := new(Empty)
	for i :=0 ; i < n; i ++ {
		s <- e
	}
}
// release n resources
func (s semaphore) V(n int) {
	for i :=0 ;i <n;i ++ {
		<-s
	}
}

// mutexes
func (s semaphore) Lock() {
	s.P(1)
}
func (s semaphore) Unlock() {
	s.V(1)
}

// signal-wait
func (s semaphore) Wait(n int) {
	s.P(n)
}
func (s semaphore) Signal() {
	s.V(1)
}
```
