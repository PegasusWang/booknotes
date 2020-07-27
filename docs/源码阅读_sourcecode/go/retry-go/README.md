# go 重试

源码：https://github.com/avast/retry-go

# 简单实现

一个简单的重试 go 函数可以这么实现

```go
func Retry(fn func() error, attempts int, sleep time.Duration) error {
	if err := fn(); err != nil {
		if s, ok := err.(stop); ok {
			return s.error
		}

		if attempts--; attempts > 0 {
			// logger.Warnf("retry func error: %s. attemps #%d after %s.", err.Error(), attempts, sleep)
			fmt.Printf("retry times[%d], sleep[%v]\n", attempts, sleep)
			time.Sleep(sleep)
			return Retry(fn, attempts, 2*sleep)
		}
		return err
	}
	return nil
}

type stop struct {
	error
}

func NoRetryError(err error) stop {
	return stop{err}
}
```
