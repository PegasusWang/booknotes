# 一个简洁的并发库 async，代码阅读

使用的时候可以传递一系列函数进去，并发执行然后获取结果。

https://github.com/rafaeldias/async

# async

async is a utility library for manipulating synchronous and asynchronous flows.

### Install

```bash
$ go get github.com/rafaeldias/async
```
### Concurrent
```go
async.Concurrent(tasks taskier) (Results, error)
```

Concurrent will execute all the functions concurrently using goroutines.

- `tasks` is an internal interface which accept two async public types:
    - `Tasks` is a list of functions to be executed concurrently.
    - `MapTasks` is a map of string of functions to be executed concurrently.

The type of `Results` value depends on the type of tasks passed to the function. See [type Results](#type-results)

All errors ocurred in the functions will be returned. See [returning error](#returning-error) and [type Errors](#type-errors).

### Parallel
```go
async.Parallel(tasks taskier) (Results, error)
```

Parallel will execute all the functions in parallel. It creates multiple goroutines and distributes the functions execution among them.
The number of goroutines  defaults to `GOMAXPROCS`. If the number of active goroutines is equal to `GOMAXPROCS` and there're more functions to execute, these functions will wait until one of functions being executed finishes its job.

- `tasks` is an internal interface which accept two async public types: 
    - `async.Tasks` is a list of functions to be executed in parallel.
    - `async.MapTasks` is a map of string of functions to be executed in parallel.

The type of `Results` value depends on the type of tasks passed to the function. See [type Results](#type-results).

All errors ocurred in the functions will be returned. See [returning error](#returning-error) and [type Errors](#type-errors).

### Waterfall
```go
async.Waterfall(tasks Tasks, args ...interface{}) ([]interface[}, error)
```

Waterfall will execute all the functions in sequence, each returning their results to the next. If the last returning value of the function is of type `error`, then this value will not be passed to the next function. 

- `tasks` is a list of functions that will be executed in series.
- `args` are optional parameters that will be passed to the first task.

Waterfall returns the results of the last task as `[]interface{}` and `error`. 

If an error occur in any of the functions to be executed, the next function will not be executed, and the error will be returned to the caller. See [returning error](#returning-error).

### <a name="type-results"></a>Type Results

The `Results` is the type that is returned by `Parallel` and `Concurrent`:

```go
type Results interface {
      Index(int) []interface{}  // Gets values by index
      Key(string) []interface{} // Gets values by key
      Len() int                 // Gets the length of the results
      Keys() []string           // Gets the keys of the results
}
```

The underlying type of `Results` will be different depending on the type of tasks passed to either `Parallel` or `Concurrent`.
There are two async public types that can be passed to both functions:

 - `Tasks` is a list of functions that will be executed.
 - `MapTasks` is a map of string of functions to be executed.

When using `Tasks`, the underlying `Results` will be `[][]interface{}`:
```go
res, err : = async.Parallel(async.Tasks{
        func1,
        func2,
        ...funcN,
})
```
We can get the results of the the second function:

```go
res.Index(1)
```

Or we can iterate over the results using the  `Len()` method:

```go
for i := 0; i < res.Len(); i++ {
    fmt.Println(res.Index(i))
}
```

When using `MapTasks`, the underlying `Results` will be `map[string][]interface{}`:

```go
res, err : = async.Concurrent(async.MapTasks{
        "one"  : funcOne,
        "two"  : funcTwo,
        "three": funcThree,
})
```

We can get the results of function "three":

```go
res.Key("three")
```
Or we can also iterate over the map of results by using the `Keys()` method:

```go
for k := range res.Keys() {
    fmt.Println(res.Key(k))
}
```

### <a name="type-errors"></a>Type Errors

If errors occur in any function executed by `Concurrent` or `Parallel` an instance of `Errors` will be returned.
`Errors` implements the `error` interface, so in order to test if an error occurred, check if the returned error is not nil,
if it's not type cast it to `Errors`:

```go
_, err : = async.Parallel(async.Tasks{func1, func2, ...funcN})

if err != nil {
        parallelErrors := err.(async.Errors)

        for _, e := range parallelErrors {
                fmt.Println(e.Error())
        }
}
```

### <a name="returning-error"></a>Returning error

In order for async to identify if an error occured, the error **must** be the last returning value of the function:

```go
_, err := async.Waterfall(async.Tasks{
        func () (int, error) {
                return 1, nil
        },
        // Function with error
        func (i int) (string, error) {
                if i > 0 {
                    // This line will interrupt the execution flow
                    return "", errors.New("Error occurred")
                }
                return "Ok", nil
        },
        // This function will not be executed.
        func (s string) {
            return
        }
});

if err != nil {
      fmt.Println(err.Error()); // "Error occurred"
}
```

# Examples 

### Waterfall

```go
import (
        "fmt"

        "github.com/rafaeldias/async"
)

func fib(p, c int) (int, int) {
  return c, p + c
}

func main() {

        // execution in series.
        res, e := async.Waterfall(async.Tasks{
                fib,
                fib,
                fib,
                func(p, c int) (int, error) {
                        return c, nil
                },
        }, 0, 1)

        if e != nil {
              fmt.Printf("Error executing a Waterfall (%s)\n", e.Error())
        }

        fmt.Println(res[0].(int)) // Prints 3
}

```

### Parallel

```go
import (
        "fmt"
        "time"

        "github.com/rafaeldias/async"
)

func main() {

        res, e := async.Parallel(async.MapTasks{
                "one": func() int {
                        for i := 'a'; i < 'a'+26; i++ {
                                fmt.Printf("%c ", i)
                        }
                        
                        return 1
                },
                "two": func() int {
                        time.Sleep(2 * time.Microsecond)
                        for i := 0; i < 27; i++ {
                                fmt.Printf("%d ", i)
                        }
                        
                        return 2
                },
                "three": func() int {
                        for i := 'z'; i >= 'a'; i-- {
                                fmt.Printf("%c ", i)
                        }
                        
                        return 3
                },
        })

        if e != nil {
                fmt.Printf("Errors [%s]\n", e.Error())
        }
        
        fmt.Println("Results from task 'two': %v", res.Key("two"))
}
```

### Concurrent

```go
import (
        "errors"
        "fmt"

        "github.com/rafaeldias/async"
)

func main() {

        res, e := async.Concurrent(async.Tasks{
                func() int {
                        for i := 'a'; i < 'a'+26; i++ {
                                fmt.Printf("%c ", i)
                        }
                        return 0
                },
                func() error {
                        time.Sleep(3 * time.Microsecond)
                        for i := 0; i < 27; i++ {
                                fmt.Printf("%d ", i)
                        }
                        return errors.New("Error executing concurently")
                },
        })

        if e != nil {
                fmt.Printf("Errors [%s]\n", e.Error()) // output errors separated by space
        }

        fmt.Println("Result from function 0: %v", res.Index(0))
}
```

# License
Distributed under MIT License. See LICENSE file for more details.
