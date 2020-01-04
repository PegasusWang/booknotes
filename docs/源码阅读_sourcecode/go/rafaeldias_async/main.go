package main

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"
)

// 阅读并且尝试抄一个https://github.com/rafaeldias/async

// 先来定义一个任务列表
type Tasks []interface{}
type taskier interface {
	GetFuncs() (interface{}, error)
}

// 获取需要执行的函数列表
func (t Tasks) GetFuncs() (interface{}, error) {
	l := len(t)
	fns := make([]reflect.Value, l)

	for i := 0; i < l; i++ {
		f := reflect.Indirect(reflect.ValueOf(t[i]))
		if f.Kind() != reflect.Func {
			return fns, fmt.Errorf("%T must be a Function ", f)
		}
		fns[i] = f
	}
	return fns, nil
}

// 定义结果类型，用来保存返回的并行任务结果。支持多种类型，所以用了interface{}
type Results interface {
	Index(int) []interface{}  // Get value by index
	Key(string) []interface{} // Get value by key
	Len() int                 // Get the length of the result
	Keys() []string           // Get the keys of the result
}

// 并行执行 taskiker 里的函数，然后返回结果
func Parallel(stack taskier) (Results, error) {
	return execConcurrentStack(stack, true)
}

type funcs struct {
	Stack interface{}
}

type Errors []error

func (e Errors) Error() string {
	b := bytes.NewBufferString(emptyStr)
	for _, err := range e {
		b.WriteString(err.Error())
		b.WriteString(" ")
	}
	return strings.TrimSpace(b.String())
}

// ExecConcurrent 并行执行所有 stack 中的函数
func (f *funcs) ExecConcurrent(parallel bool) (Results, error) {
	var results Results
	var errs Errors
	if funcs, ok := f.Stack.([]reflect.Value); ok {
		results, errs = execSlice(funcs, parallel)
	} else {
		panic("Stack type must be of type []reflect.Value or map[string]reflect.Value.")
	}
	if len(errs) == 0 {
		return results, nil
	}
	return results, errs
}

// 实现接口 Results 的几个method
type sliceResults [][]interface{}

// 获取第 i 个 task 的结果
func (s sliceResults) Index(i int) []interface{} {
	return s[i]
}
func (s sliceResults) Len() int {
	return len(s)
}
func (s sliceResults) Keys() []string {
	panic("Cannot get map keys from Slice")
}

// Not supported by sliceResults
func (s sliceResults) Key(k string) []interface{} {
	panic("Cannot get map key from Slice")
}

type execResult struct {
	err     error
	results []reflect.Value
	key     string
}

var (
	emptyStr    string
	emptyError  error
	emptyResult []interface{}
	emptyArgs   []reflect.Value
)

// 执行函数
func execSlice(funcs []reflect.Value, parallel bool) (sliceResults, Errors) {
	var (
		errs    Errors
		results = sliceResults{}
		ls      = len(funcs) // 需要执行函数的长度
		cr      = make(chan execResult, ls)
	)
	if parallel {
		// 限制并发数
		sem := make(chan int, runtime.GOMAXPROCS(0))
		for i := 0; i < ls; i++ {
			sem <- 1
			go execRoutineParallel(funcs[i], cr, sem, emptyStr)
		}
	} else {
		for i := 0; i < ls; i++ {
			go execRoutine(funcs[i], cr, emptyStr)
		}
	}
	// 消费结果
	for i := 0; i < ls; i++ {
		r := <-cr
		if r.err != nil {
			errs = append(errs, r.err)
		} else if lcr := len(r.results); lcr > 0 {
			res := make([]interface{}, lcr)
			for j, v := range r.results {
				res[j] = v.Interface()
			}
			results = append(results, res)
		}
	}
	return results, errs
}

func execRoutineParallel(f reflect.Value, c chan execResult, sem chan int, k string) {
	execRoutine(f, c, k)
	// Once the task has done its job, consumes message from channel `sem`
	<-sem
}

func execRoutine(f reflect.Value, c chan execResult, key string) {
	var (
		exr = execResult{}      //result
		res = f.Call(emptyArgs) // calls the function
	)
	fnt := f.Type()               //获取要执行函数的类型
	if l := fnt.NumOut(); l > 0 { // NumOut 是函数返回值个数 [0, NumOut())]
		lastArg := fnt.Out(l - 1)
		if reflect.Zero(lastArg).Interface() == emptyError {
			if e, ok := res[l-1].Interface().(error); ok {
				exr.err = e
			}
			l = l - 1 // 去掉最后一个 error 返回值的下标
		}

		if exr.err == nil && l > 0 {
			exr.results = res[:l]
			if key != "" {
				if key != "" {
					exr.key = key
				}
			}
		}
	}
	c <- exr
}

func execConcurrentStack(stack taskier, parallel bool) (Results, error) {
	var err error
	f := new(funcs)
	f.Stack, err = stack.GetFuncs()
	if err != nil {
		panic(err)
	}
	return f.ExecConcurrent(parallel)
}

func main() {
	tasks := Tasks{
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
	}
	res, err := Parallel(tasks)
	if err != nil {
		fmt.Printf("Errors [%s]\n", err.Error()) // output errors separated by space
	}
	fmt.Println("Result from function 0: %v", res.Index(0))
}
