// Package  mystatsd provides ...
package mystatsd

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"math/rand"
)

type conn struct {
	addr          string
	errorHandler  func(error)
	flushPeriod   time.Duration
	maxPacketSize int
	network       string
	tagFormat     TagFormat

	mu        sync.Mutex //  go 里 不用显示初始化 mutex
	closed    bool
	w         io.WriteCloser
	buf       []byte
	rateCache map[float32]string
}

func newConn(conf connConfig, muted bool) (*conn, error) {
	c := &conn{
		addr:          conf.Addr,
		errorHandler:  conf.ErrorHandler,
		flushPeriod:   conf.FlushPeriod,
		maxPacketSize: conf.MaxPacketSize,
		network:       conf.Network,
		tagFormat:     conf.TagFormat,
	}
	if muted {
		return c, nil
	}

	var err error
	c.w, err = dialTimeout(c.network, c.addr, 5*time.Second)
	if err != nil {
		return c, err
	}

	// 如果是 udp 发送一个检查
	if c.network[:3] == "udp" {
		for i := 0; i < 2; i++ {
			_, err = c.w.Write(nil)
			if err != nil {
				_ = c.w.Close()
				c.w = nil
				return c, err
			}
		}
	}

	c.buf = make([]byte, 0, c.maxPacketSize+200) // 防止溢出

	if c.flushPeriod > 0 {
		go func() {
			ticker := time.NewTicker(c.flushPeriod)
			for range ticker.C {
				c.mu.Lock()
				if c.closed {
					ticker.Stop()
					c.mu.Unlock()
					return
				}
				c.flush(0)
				c.mu.Unlock()
			}
		}()
	}
	return c, nil
}

// echo "system.load.1:0.5|g" | nc 127.0.0.1 8251

/*
协议： https://my.oschina.net/solate/blog/3076247/print

<bucket>:<value>|<type>[|@sample_rate]

bucket是一个metric的标识，可以看成一个metric的变量。
value: metric的值，通常是数字。
metric的类型，通常有timer、counter、gauge和set四种。
sample_rate


Counting: gorets:1|c
Sampling: gorets:1|c|@0.1
Timing: glork:320|ms|@0.1
Gauges: gaugor:333|g
Sets: uniques:765|s
*/
func (c *conn) metric(prefix, bucket string, n interface{}, typ string, rate float32, tags string) {
	fmt.Println(prefix, bucket, n, typ, rate, tags)
	c.mu.Lock()
	l := len(c.buf)
	// 分别按照协议 append 值
	c.appendBucket(prefix, bucket, tags)
	c.appendNumber(n)
	c.appendType(typ)
	c.appendRate(rate) // 最后的采样值
	c.closeMetric(tags)
	c.flushIfBufferFull(l)
	c.mu.Unlock()
}

func (c *conn) appendByte(b byte) {
	c.buf = append(c.buf, b)
}

func (c *conn) appendString(s string) {
	c.buf = append(c.buf, s...)

}
func (c *conn) appendBucket(prefix, bucket string, tags string) {
	c.appendString(prefix)
	c.appendString(bucket)
	if c.tagFormat == InfluxDB {
		c.appendString(tags)
	}
	c.appendByte(':')
}

// 主要调用 strconv 包，把数字写入到 buf
func (c *conn) appendNumber(v interface{}) {
	switch n := v.(type) {
	case int:
		c.buf = strconv.AppendInt(c.buf, int64(n), 10)
	case uint:
		c.buf = strconv.AppendUint(c.buf, uint64(n), 10)
	case int64:
		c.buf = strconv.AppendInt(c.buf, n, 10)
	case uint64:
		c.buf = strconv.AppendUint(c.buf, n, 10)
	case int32:
		c.buf = strconv.AppendInt(c.buf, int64(n), 10)
	case uint32:
		c.buf = strconv.AppendUint(c.buf, uint64(n), 10)
	case int16:
		c.buf = strconv.AppendInt(c.buf, int64(n), 10)
	case uint16:
		c.buf = strconv.AppendUint(c.buf, uint64(n), 10)
	case int8:
		c.buf = strconv.AppendInt(c.buf, int64(n), 10)
	case uint8:
		c.buf = strconv.AppendUint(c.buf, uint64(n), 10)
	case float64:
		c.buf = strconv.AppendFloat(c.buf, n, 'f', -1, 64)
	case float32:
		c.buf = strconv.AppendFloat(c.buf, float64(n), 'f', -1, 32)
	}
}

// echo "system.load.1:0.5|g" | nc 127.0.0.1 8251
func (c *conn) appendType(t string) {
	c.appendByte('|') // 按照协议用 | 分隔
	c.appendString(t)
}

func (c *conn) appendRate(rate float32) {
	if rate == 1 {
		return
	}
	if c.rateCache == nil {
		c.rateCache = make(map[float32]string)
	}

	c.appendString("|@")
	if s, ok := c.rateCache[rate]; ok {
		c.appendString(s)
	} else {
		s = strconv.FormatFloat(float64(rate), 'f', -1, 32)
		c.rateCache[rate] = s
		c.appendString(s)
	}
}

func (c *conn) closeMetric(tags string) {
	if c.tagFormat == Datadog {
		c.appendString(tags)
	}
	// StatsD supports receiving multiple metrics in a single packet by separating them with a newline.
	// 根据文档，多个用换行符分隔 gorets:1|c\nglork:320|ms\ngaugor:333|g\nuniques:765|s
	c.appendByte('\n')
}

func (c *conn) flushIfBufferFull(lastSafeLen int) {
	if len(c.buf) > c.maxPacketSize {
		c.flush(lastSafeLen)
	}
}

// flush buf 的前 n 个字节。n = 0 全部刷
func (c *conn) flush(n int) {
	fmt.Println("flush", c.buf)
	if len(c.buf) == 0 {
		return
	}
	if n == 0 {
		n = len(c.buf)
	}

	_, err := c.w.Write(c.buf[:n-1]) // 去掉最后的 \n
	c.handleError(err)
	if n < len(c.buf) {
		copy(c.buf, c.buf[n:]) // copy buf 的后边内容到前边
	}
	c.buf = c.buf[:len(c.buf)-n]
}

func (c *conn) handleError(err error) {
	if err != nil && c.errorHandler != nil {
		fmt.Println(err)
		c.errorHandler(err)
	}
}

//  https://github.com/statsd/statsd/blob/master/docs/metric_types.md#gauges
func (c *conn) gauge(prefix, bucket string, value interface{}, tags string) {
	c.mu.Lock()
	l := len(c.buf)
	if isNegative(value) { // This implies you can't explicitly set a gauge to a negative number without first setting it to zero.
		c.appendBucket(prefix, bucket, tags)
		c.appendGauge(0, tags)
	} // 这里确实不是 bug，按照文档就是需要先 append 一个 0
	c.appendBucket(prefix, bucket, tags)
	c.appendGauge(value, tags)
	c.flushIfBufferFull(l)
	c.mu.Unlock()
}

func (c *conn) appendGauge(value interface{}, tags string) {
	c.appendNumber(value)
	c.appendType("g")
	c.closeMetric(tags)
}

func (c *conn) unique(prefix, bucket string, value string, tags string) {
	c.mu.Lock()
	l := len(c.buf)
	c.appendBucket(prefix, bucket, tags)
	c.appendString(value)
	c.appendType("s")
	c.closeMetric(tags)
	c.flushIfBufferFull(l)
	c.mu.Unlock()
}

// 判断是否小于 0 （真麻烦）
func isNegative(v interface{}) bool {
	switch n := v.(type) {
	case int:
		return n < 0
	case uint:
		return n < 0
	case int64:
		return n < 0
	case uint64:
		return n < 0
	case int32:
		return n < 0
	case uint32:
		return n < 0
	case int16:
		return n < 0
	case uint16:
		return n < 0
	case int8:
		return n < 0
	case uint8:
		return n < 0
	case float64:
		return n < 0
	case float32:
		return n < 0
	}
	return false
}

// 方便 mock
var (
	dialTimeout = net.DialTimeout
	now         = time.Now
	randFloat   = rand.Float32
)
