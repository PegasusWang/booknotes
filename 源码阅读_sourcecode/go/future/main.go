package main

// package myfuture

import (
	"context"
	"errors"
	"fmt"

	pkgErrors "github.com/pkg/errors"
)

type Func func() error

type Future struct {
	ch  chan struct{}
	fn  Func
	err error
}

func New(fn Func) *Future {
	f := &Future{
		ch: make(chan struct{}),
		fn: fn,
	}
	f.start()
	return f
}

func (f *Future) start() {
	go func() {
		defer func() {
			if rval := recover(); rval != nil {
				if err, ok := rval.(error); ok {
					f.err = pkgErrors.WithStack(err)
				} else {
					rvalStr := fmt.Sprint(rval)
					f.err = pkgErrors.WithStack(errors.New(rvalStr))
				}
			}
			close(f.ch)
		}()

		f.err = f.fn()
		return
	}()
}

func (f *Future) Get(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-f.ch:
		return f.err
	}
}

func (f *Future) Done() bool {
	select {
	case <-f.ch:
		return true
	default:
		return false
	}
}

/*
func GetSomeData(ctx context.Context) (int, error) {
    return 1, nil
}

var ctx context.Context
var rv int
f := future.New(func() error {
    var err error
    rv, err = GetSomeData(ctx)
    return err
})
err := f.Get(ctx)

fmt.Println(rv)
*/
func GetSomeData(ctx context.Context) (int, error) {
	// return 1, nil
	return 0, errors.New("fake")
}

func testOne() {
	// NOTE: 这里其实 Future 没有给出参数和返回值，用的都是外部闭包变量的，可以避免各种 interface 操作
	// 参数和返回值都应该放到闭包变量
	ctx := context.Background()
	var rv int
	f := New(func() error {
		var err error
		rv, err = GetSomeData(ctx)
		return err
	})
	err := f.Get(ctx)
	fmt.Println(rv, err)
}

func main() {
	testOne()
}
