# 代码阅读

https://github.com/patrickmn/go-cache

```sh
go get github.com/patrickmn/go-cache
```

# 进程缓存使用场景

除了使用 redis 等内存缓存之外，有时候还需要单机 cache 场景(不用和 redis server 进行网络请求了)。
举一个例子就是像是微博这种服务，对于一些超高热点数据(比如某明星又出轨），缓存的 key 会集中在某台 redis
上，很多个服务集中请求某台 redis 导致其压力过高。这个时候可以使用单机 cache，直接换存到服务进程内存中，连 redis
不用请求了，同时将压力分散到多个服务上，缓解缓存服务器压力。这个是一个很典型的使用场景(需要保证一致性，一般是先从 redis
拉过来再缓存到进程内存）。

注意：该场景下不同机器的缓存不一致的问题。

# go-cache

star 数量较高，也比较稳定了。实现了 goroutine 安全的 map[string]interface{}，包含超时。支持序列化到文件并且重新加载

示例代码:

```go
import (
	"fmt"
	"github.com/patrickmn/go-cache"
	"time"
)

func main() {
	// Create a cache with a default expiration time of 5 minutes, and which
	// purges expired items every 10 minutes
	c := cache.New(5*time.Minute, 10*time.Minute)

	// Set the value of the key "foo" to "bar", with the default expiration time
	c.Set("foo", "bar", cache.DefaultExpiration)

	// Set the value of the key "baz" to 42, with no expiration time
	// (the item won't be removed until it is re-set, or removed using
	// c.Delete("baz")
	c.Set("baz", 42, cache.NoExpiration)

	// Get the string associated with the key "foo" from the cache
	foo, found := c.Get("foo")
	if found {
		fmt.Println(foo)
	}

	// Since Go is statically typed, and cache values can be anything, type
	// assertion is needed when values are being passed to functions that don't
	// take arbitrary types, (i.e. interface{}). The simplest way to do this for
	// values which will only be used once--e.g. for passing to another
	// function--is:
	foo, found := c.Get("foo")
	if found {
		MyFunction(foo.(string))
	}

	// This gets tedious if the value is used several times in the same function.
	// You might do either of the following instead:
	if x, found := c.Get("foo"); found {
		foo := x.(string)
		// ...
	}
	// or
	var foo string
	if x, found := c.Get("foo"); found {
		foo = x.(string)
	}
	// ...
	// foo can then be passed around freely as a string

	// Want performance? Store pointers!
	c.Set("foo", &MyStruct, cache.DefaultExpiration)
	if x, found := c.Get("foo"); found {
		foo := x.(*MyStruct)
			// ...
	}
}
```
