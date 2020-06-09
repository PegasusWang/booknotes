package myratelimit

import (
	"fmt"
	"testing"
	"time"
)

func TestRateLimit(t *testing.T) {
	rl := New(100) // per second

	prev := time.Now()
	for i := 0; i < 10; i++ {
		now := rl.Take()
		fmt.Println(i, now.Sub(prev))
		prev = now
	}
}
