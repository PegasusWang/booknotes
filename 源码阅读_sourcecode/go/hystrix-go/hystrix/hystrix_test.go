package myhystrix

import (
	"fmt"
	"testing"
)

func TestDo(t *testing.T) {
	err := Do("test", func() error {
		// call service
		fmt.Println("call")
		return nil
	}, func(err error) error {
		return nil
	})
	fmt.Println(err)
}
