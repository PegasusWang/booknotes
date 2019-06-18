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
### 14.2.10 For-range applied to channels

It reads from the given channel ch until the channel is closed and then the code following for continue to execute. Obviously another goroutine must be writing to ch(otherwise the executino blocks in the for-loop) and must close ch when it's done writing.

producer consumer pattern:


```
for {
	Consume(Product())
}
```

### 14.2.11 Channel directionality

```
var send_only chan<- int // channel can only receive data (<-chan T)
var recv_only <-chan int // channel can only send data
```

IDIOM: Pipe and filter pattern

```
sendChan := make(chan int)
receiveChan := make(chan string)
go processChannel(sendChan, receiveChan)

func processChannel(in <- chan int, out chan<- string) {
	for inValue := range in {
		result := // process inValue
		out<-result
	}
}
```

```
// sieve primer number
package main

import "fmt"

func generate() chan int {
	ch := make(chan int)
	go func() {
		for i := 2; ; i++ {
			ch <- i
		}
	}()
	return ch
}

func filter(in chan int, prime int) chan int {
	out := make(chan int)
	go func() {
		for {
			if i := <-in; i%prime != 0 {
				out <- i
			}
		}
	}()
	return out
}

func sieve() chan int {
	out := make(chan int)
	go func() {
		ch := generate()
		for {
			prime := <-ch
			ch = filter(ch, prime)
			out <- prime
		}
	}()
	return out
}

func main() {
	primes := sieve()
	for {
		fmt.Println(<-primes)
	}
}
```


## 14.3 Synchronization of goroutine: closing a channel - testing for blocked channels

Only the sender should close a channel, never receiver.
Sending or Closing a closed channel causes a run-time panic.

```
if v, ok := <-ch; ok {
	process(v)
}
```

To do a non-blocking channel you need to use a select.
for range will automatically detect when the channel is closed.

## 14.4 Switching between goroutines with select

Getting the values out of dirrerent concurrently executing goroutines can be accomplisehd with
the select ekyworkd. A select is terminated when a break or return is executed in one of its cases.

- if all are blocked, it waits until one can proceed
- if multiple can proceed, it choose one at random
- when none of the channel operations can proceed and default clause is present, default is always runnable

```
// sieve primer number
package main

import (
	"fmt"
	"runtime"
	"time"
)

func main() {
	runtime.GOMAXPROCS(2)
	ch1 := make(chan int)
	ch2 := make(chan int)

	go pump1(ch1)
	go pump2(ch2)
	go suck(ch1, ch2)
	time.Sleep(1e9)
}

func pump1(ch chan int) {
	for i := 0; ; i++ {
		ch <- i * 2
	}
}

func pump2(ch chan int) {
	for i := 0; ; i++ {
		ch <- i + 5
	}
}
func suck(ch1 chan int, ch2 chan int) {
	for {
		select {
		case v := <-ch1:
			fmt.Printf("received on channel 1 :%d\n", v)
		case v := <-ch2:
			fmt.Printf("received on channel 2 :%d\n", v)
		}
	}
}
```

## 14.5 Channels, Timeouts and Tickers

time.Ticker is an object that repeatedly sends a time value on a contained channel C at a specified time interval.

```
func main() {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	select {
	case u := <-ch1:
		//...
	case v := <-ch2:
		//...
	case <-ticker.C:
		logState(status) // call some logging function logState
	default: // novalue ready to be received
		//...
	}
}
```

rate limiter:


```
rate_per_sec := 10
var dur Dration = 1e9
chRate := time.Tick(dur)
for req := range requests {
	<- chRate
	go client.Call("service.Method", req, ...)
}
```

time.After(d) only send time once:


```
package main

import (
	"fmt"
	"time"
)

func main() {
	tick := time.Tick(1e8)
	boom := time.After(5e8)
	for {
		select {
		case <-tick:
			fmt.Println("tick.")
		case <-boom:
			fmt.Println("Boom.")
			return
		default:
			fmt.Println("     .")
			time.Sleep(5e7)
		}
	}
}
```

## 14.6 Using recover with goroutines

```
func server(workChan <-chan *Work) {
	for work := range workChan {
		go safelyDo(work)
	}
}

func safelyDo(work *Work) {
	defer func() {
		if err := recover(); err !=nil {
			log.Printf("work faield with %s in %v:", err, work)
		}
	}()
	do(work)
}
```

### 14.7 Tasks and Worker processes


```
// Master-worker paradigm
func Worker(in, out chan *Task){
	for {
		t := <-in
		process(t)
		out <- t
	}
}
```

when to use: a sync.Mutex or a channel?

- use locking(mutexes) when:
    - caching information in a shared data structure
    - holding state information, that is context or status of the running application

- use channels when:
    - communicating asynchronous results
    - distributing units of work
    - passing owership of data

## 14.8 Implementing a lazy generator

A generator is a function that returns the next value in a sequence each time the function is called.

```
package main

import (
	"fmt"
)

var resume chan int

func intergers() chan int {
	yield := make(chan int)
	count := 0
	go func() {
		for {
			yield <- count
			count++
		}
	}()
	return yield
}
func generateInteger() int {
	return <-resume
}
func main() {
	resume = intergers()
	fmt.Println(generateInteger())
	fmt.Println(generateInteger())
	fmt.Println(generateInteger())
}
```

## 14.9 Implementing Futures

future: somtimes you know yopu need to compute a value before you need to actually use the value.
in this case, you can potentially start computing the value on another processor and have it ready
when you need it.


```
func InverseProduct(a Matrix, b Matrix) {
	a_inv := Inverse(a)
	b_inv := Inverse(b)
	return Product(a_inv, b_inv)
}


/// a_inv and b_inv can compute parallel

fucn InverseProduct(a Matrix, b Matrix) {
	a_inv_future := InverseFuture(a) //. started as a goroutine
	b_inv_future := InverseFuture(b)
	a_inv := <-a_inv_future
	b_inv := <-b_inv_future
	return Product(a_inv,b_inv)
}

func InverseFuture(a Matrix) {
	future := make(chan Matrix)
	go func() { future<-Inverse(a) } () // launched a closure as a goroutine
	return future
}
```

## 14.10 Multiplexing
## 14.11 Limiting the number of requests processed concurrently

## 14.12 Chaining goroutines
## 14.13 Parallelzing a computation over a number of cores

```
func DoAll() {
	sem := make(chan int, NCPU)
	for i := 0; i < NCPU; i ++ {
		go DoPart(sem)
	}
	// Drain the channel sem, waiting for  NCPU tasks to complete
	for i:= 0; i < NCPU; i ++ {
		<-sem
	}
	// All done
}

func DoPart(sem chan int) {
	// do the part of the computation
	sem <- 1 // signal that thie piece is done
}

func main() {
	runtime.GOMAXPROCS = NCPU
	DoAll()
}
```

## 14.14 Parallelizing a computation over a large amount of data

a number of steps: Preprocess / StepA / StepB / ... / PostProcess

pipelining algorithm:


```
func SerialProcessData(in <- chan *Data, out <- chan *Data) {
	for data := range in {
		tmpA := PreprocessData(data)
		tmpB := ProcessStepA(tmpA)
		tmpC := ProcessStepA(tmpB)
		out <- PostProcess(tmpC)
	}
}

func ParallelProcessData(in <- chan  *Data, out <- chan *Data) {
	// make channels:
	preOut := make(chan *Data, 100)
	stepAOut := make(chan *Data, 100)
	stepBOut := make(chan *Data, 100)
	stepCOut := make(chan *Data, 100)
	// start parallel computations
	go PreprocessData(in, preOut)
	go ProcessStepA(preOut, stepAOut)
	go ProcessStepB(stepAOut, stepBOut)
	go ProcessStepC(stepBOut, stepCOut)
	go PostProcessData(stepOut, out)
}
```

## 14.15 The leaky bucket algorithm

```
var freeList = make(chan *Buffer, 100)
var serverChan = make(chan *Buffer)

func client() {
	for {
		var b *Buffer
		select {
		case b = <-freeList:
			// Got one; nothing more todo
		default:
			b = new(Buffer)
			loadInto(b) // read next message from the network
		}
		serverChan <- b //send to server
	}
}

func server() {
	for  {
		b := <-serverChan // wait for work
		process(b)
		// reuse buffer is threre's room
		select {
		case freeList <- b:
			// Reuse buffer if free slot on freeList; nothing more to do
		default:
			// freeList is full, just carry on : the buffer is 'dropped'
			// doesn't work when freeList is full, leaky bucket
		}
	}

}
```

## 14.16 Benchmarking goroutines

## 14.17 Concurrent access to object using a chnnel

To safeguard concurrent modifications of an object instead of using locking with a sync Mutex,
we can also use a backend goroutine for the sequential execution of anonymous functions.


# 16 Common Go Pitfalls or Mistakes

## 16.1 Hiding(shadowing) a variable by misusing short declaration

```
var remeber bool = false
if somehting {
	remeber := true // use = not :=
}
// use member


func shadow() (err error) {
	x, err := check1() // x is created, err is assigned to
	if err != nil {
		return   // err correctly returned
	}
	if y, err := check2(x); err !=nil { //y and inner err are created
		return // inner err shadows err so nil is wrongly returned!
	} else {
		fmt.Println(y)
	}
	return
}

```

## 16.2 Misusing strings
mind that strings in go(like java and python) are immutable, Do not use a + b in a for loop,
intead one should use a bytes.Buffer to accumulate string content.

```
func shadow() (err error) {
	x, err := check1() // x is created, err is assigned to
	if err != nil {
		return   // err correctly returned
	}
	if y, err := check2(x); err !=nil { //y and inner err are created
		return // inner err shadows err so nil is wrongly returned!
	} else {
		fmt.Println(y)
	}
	return
}
```

## 16.3 Using defer for closing a file in the wrong scope

Defer is only executed at the return of a function, not at the end of a loop or some other limited scope.

## 16.4 Confusing new() and make()

- for slices, maps and channels, use make
- for arrays, structs and all value types: use new

## 16.5 No need to pass a pointer to a slice to a function

## 16.6 Using pointers to interface types

Never use a pointer to an interface type, this is already a pointer!

## 16.7 Misusing pointers wiht value types

## 16.8 Misusing goroutines and channels
Only using goroutines and channels only where concurrency is important!

## 16.9 Using closures with goroutines

```
package main

import (
	"fmt"
	"time"
)

var values = [5]int{10, 11, 12, 13, 14}

func main() {
	// version A
	for ix := range values {
		func() {
			fmt.Print(ix, " ")
		}()
	}
	fmt.Println()   // 0,1,2,3,4

	// version B
	for ix := range values {
		go func() {  // invoke each closure as a goroutine
			fmt.Print(ix, " ")
		}() // the goroutine will probably not begin executing until after the loop
	}
	fmt.Println() // 4,4,4,4,4
	time.Sleep(5e9)
	// Version C: the right way
	for ix := range values {
		go func(ix interface{}) { // invoke each closure with ix as a parameter
			fmt.Print(ix, " ") //ix is the evaluated at each iteration and placed on the stack of goroutine
		}(ix)
	}
	fmt.Println()  // random
	time.Sleep(5e9)
	// Version D: print out values
	for ix := range values {
		val := values[ix] // variables declared witin the body of a loop are not shared between iterations
		go func() {
			fmt.Print(val, " ")
		}()
	}
	time.Sleep(1e9) //10,11,12,13,14
}
```

## 16.10 Bad error handling

wrap your error conditions in a closure wherever possible, like in the following example.

```
// seperate error checking, error reporting, and normal program logic

func httpRequestHandler(w http.ResponseWriter, req *http.Request) {
	err := func() error {
		if req.Method != "GET"{
			return errors.New("expected GET")
		}
		if input := paraseInput(req); input != "command" {
			return errors.New("malformed command")
		}
	}()
	// other error conditions can be tested here
	if err != nil {
		w.WriteHeader(400)
		io.WriterString(w, err)
		return
	}
	doSomething() //
}
```

# 17 Go Language Patterns

## 17.1 The comma, ok pattern

## 17.2 The defer pattern

## 17.3 The visibility pattern

## 17.4 The operator pattern and interface
go不支持操作符重载，可以使用函数/方法/接口 来模拟。


