// 源码：https://github.com/avast/retry-go
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type RetryableFunc func() error

// 默认配置
var (
	DefaultAttempts      = uint(10)
	DefaultDelay         = 100 * time.Millisecond
	DefaultMaxJitter     = 100 * time.Millisecond
	DefaultOnRetry       = func(n uint, err error) {}
	DefaultRetryIf       = IsRecoverable
	DefaultDelayType     = CombineDelay(BackOffDelay, RandomDelay) // 指数退避+随机
	DefaultLastErrorOnly = false
)

// 主体重试方法
func Do(retryableFunc RetryableFunc, opts ...Option) error {
	var n uint // 表示第几次
	// 默认配置
	config := &Config{
		attempts:      DefaultAttempts,
		delay:         DefaultDelay,
		maxJitter:     DefaultMaxJitter,
		onRetry:       DefaultOnRetry,
		retryIf:       DefaultRetryIf,
		delayType:     DefaultDelayType,
		lastErrorOnly: DefaultLastErrorOnly,
	}
	for _, opt := range opts {
		opt(config)
	}

	var errorLog Error
	if !config.lastErrorOnly { // 只展示最后一个错误
		errorLog = make(Error, config.attempts)
	} else {
		errorLog = make(Error, 1)
	}

	lastErrIndex := n
	for n < config.attempts {
		err := retryableFunc()
		if err != nil {
			errorLog[lastErrIndex] = unpackUnrecoverable(err)

			if !config.retryIf(err) {
				break
			}
			config.onRetry(n, err)

			if n == config.attempts-1 { // 最后一次尝试不用 sleep
				break
			}

			delayTime := config.delayType(n, config)
			if config.maxDelay > 0 && delayTime > config.maxDelay {
				delayTime = config.maxDelay
			}

			fmt.Printf("retry [%d]\n", n)
			time.Sleep(delayTime)
		} else {
			return nil
		}

		n++
		if !config.lastErrorOnly {
			lastErrIndex = n
		}
	}

	if config.lastErrorOnly {
		return errorLog[lastErrIndex]
	}
	return errorLog // 注意如果上边重试成功了，是不返回这个 errorLog的

}

type Error []error // 记录过程中的所有 error

// 拼装所有错误信息
func (e Error) Error() string {
	logWithNumber := make([]string, lenWithoutNil(e))
	for i, l := range e {
		if l != nil {
			logWithNumber[i] = fmt.Sprintf("#%d: %s", i+1, l.Error()) // 赋值每一个 error msg
		}
	}
	return fmt.Sprintf("All attempts fail:\n%s", strings.Join(logWithNumber, "\n"))
}

func lenWithoutNil(e Error) (count int) {
	for _, v := range e {
		if v != nil {
			count++
		}
	}
	return
}

type unrecoverableError struct {
	error
}

func Unrecoverable(err error) error {
	return unrecoverableError{err}
}

func IsRecoverable(err error) bool {
	_, isUnrecoverable := err.(unrecoverableError)
	return !isUnrecoverable
}

func unpackUnrecoverable(err error) error {
	if unrecoverableError, isUnrecoverable := err.(unrecoverableError); isUnrecoverable {
		return unrecoverableError.error
	}
	return err
}

func testReq() {
	url := "http://example.com"
	var body []byte

	err := Do(
		func() error {
			resp, err := http.Get(url)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			return nil
		},
	)

	fmt.Println(body, err)
}

func test() {
	n := 0
	testerr := fmt.Errorf("test error")
	// err := retry.Do(
	err := Do(
		func() error {
			if n <= 2 {
				n++
				return testerr
			} else {
				n++
				return nil
			}
		},
		Attempts(3),
	)

	fmt.Println("final err", err) // 注意：只有所有的重试都失败了才会返回 all attempts err
}

func main() {
	test()
}
