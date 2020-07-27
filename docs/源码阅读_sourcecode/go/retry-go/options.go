package main

import (
	"math/rand"
	"time"
)

// Function signature of retry if function
type RetryIfFunc func(error) bool

// Function signature of OnRetry function
// n = count of attempts
type OnRetryFunc func(n uint, err error)

type DelayTypeFunc func(n uint, config *Config) time.Duration

type Config struct {
	attempts      uint
	delay         time.Duration
	maxDelay      time.Duration
	maxJitter     time.Duration
	onRetry       OnRetryFunc
	retryIf       RetryIfFunc
	delayType     DelayTypeFunc
	lastErrorOnly bool
}

type Option func(*Config)

// 是否只返回最后一个错误
func LastErrorOnly(lastErrorOnly bool) Option {
	return func(c *Config) {
		c.lastErrorOnly = lastErrorOnly
	}
}

// 设置重试次数
func Attempts(attempts uint) Option {
	return func(c *Config) {
		c.attempts = attempts
	}
}

// 设置重试间隔，默认 100ms
func Delay(delay time.Duration) Option {
	return func(c *Config) {
		c.delay = delay
	}
}

func MaxDelay(maxDelay time.Duration) Option {
	return func(c *Config) {
		c.maxDelay = maxDelay
	}
}

// MaxJitter sets the maximum random Jitter between retries for RandomDelay
func MaxJitter(maxJitter time.Duration) Option {
	return func(c *Config) {
		c.maxJitter = maxJitter
	}
}

// DelayType set type of the delay between retries
// default is BackOff (指数退避类型，每次 sleep 乘以 2)
func DelayType(delayType DelayTypeFunc) Option {
	return func(c *Config) {
		c.delayType = delayType
	}
}

// 指数退避
func BackOffDelay(n uint, config *Config) time.Duration {
	return config.delay * (1 << n)
}

// 固定时间 sleep
func FixedDelay(_ uint, config *Config) time.Duration {
	return config.delay
}

// 随机 ，最大 maxJitter
func RandomDelay(_ uint, config *Config) time.Duration {
	return time.Duration(rand.Int63n(int64(config.maxJitter)))
}

// 组合多个 delay 方式，比如指数退避+随机
func CombineDelay(delays ...DelayTypeFunc) DelayTypeFunc {
	return func(n uint, config *Config) time.Duration {
		var total time.Duration
		for _, delay := range delays {
			total += delay(n, config)
		}
		return total
	}
}

// OnRetry function callback are called each retry
//
// log each retry example:
//
//	retry.Do(
//		func() error {
//			return errors.New("some error")
//		},
//		retry.OnRetry(func(n uint, err error) {
//			log.Printf("#%d: %s\n", n, err)
//		}),
//	)
// 重试的时候执行的回调函数
func OnRetry(onRetry OnRetryFunc) Option {
	return func(c *Config) {
		c.onRetry = onRetry
	}
}

// RetryIf controls whether a retry should be attempted after an error
// (assuming there are any retry attempts remaining)
//
// skip retry if special error example:
//
//	retry.Do(
//		func() error {
//			return errors.New("special error")
//		},
//		retry.RetryIf(func(err error) bool {
//			if err.Error() == "special error" {
//				return false
//			}
//			return true
//		})
//	)
//
// By default RetryIf stops execution if the error is wrapped using `retry.Unrecoverable`,
// so above example may also be shortened to:
//
//	retry.Do(
//		func() error {
//			return retry.Unrecoverable(errors.New("special error"))
//		}
//	)
// 可以指定重试的条件。
func RetryIf(retryIf RetryIfFunc) Option {
	return func(c *Config) {
		c.retryIf = retryIf
	}
}
