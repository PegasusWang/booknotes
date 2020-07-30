// Package  mystatsd provides ...
package mystatsd

import (
	"bytes"
	"strings"
	"time"
)

type config struct {
	Conn   connConfig
	Client clientConfig
}

type clientConfig struct {
	Muted  bool
	Rate   float32
	Prefix string
	Tags   []tag
}

type tag struct {
	K, V string
}

// 连接配置
type connConfig struct {
	Addr          string
	ErrorHandler  func(error)
	FlushPeriod   time.Duration
	MaxPacketSize int
	Network       string
	TagFormat     TagFormat
}

type Option func(*config)

// 一堆设置函数
func Address(addr string) Option {
	return Option(func(c *config) {
		c.Conn.Addr = addr
	})
}

func ErrorHandler(h func(error)) Option {
	return Option(func(c *config) {
		c.Conn.ErrorHandler = h
	})
}

func FlushPeriod(p time.Duration) Option {
	return Option(func(c *config) {
		c.Conn.FlushPeriod = p
	})
}

func MaxPacketSize(n int) Option {
	return Option(func(c *config) {
		c.Conn.MaxPacketSize = n
	})
}

func Mute(b bool) Option {
	return Option(func(c *config) {
		c.Client.Muted = b
	})
}

// SampleRate sets the sample rate of the Client. It allows sending the metrics
// less often which can be useful for performance intensive code paths.
func SampleRate(rate float32) Option {
	return Option(func(c *config) {
		c.Client.Rate = rate
	})
}

func Prefix(p string) Option {
	return Option(func(c *config) {
		// 注意不要多余的 dot
		c.Client.Prefix += strings.TrimSuffix(p, ".") + "."
	})
}

type TagFormat uint8

func TagsFormat(tf TagFormat) Option {
	return Option(func(c *config) {
		c.Conn.TagFormat = tf
	})
}

func joinTags(tf TagFormat, tags []tag) string {
	if len(tags) == 0 || tf == 0 {
		return ""
	}
	join := joinFuncs[tf]
	return join(tags)
}

func splitTags(tf TagFormat, tags string) []tag {
	if len(tags) == 0 || tf == 0 {
		return nil
	}
	split := splitFuncs[tf]
	return split(tags)
}

const (
	// InfluxDB tag format.
	// See https://influxdb.com/blog/2015/11/03/getting_started_with_influx_statsd.html
	InfluxDB TagFormat = iota + 1
	// Datadog tag format.
	// See http://docs.datadoghq.com/guides/metrics/#tags
	Datadog
)

var (
	joinFuncs = map[TagFormat]func([]tag) string{
		// InfluxDB tag format: ,tag1=payroll,region=us-west
		// https://influxdb.com/blog/2015/11/03/getting_started_with_influx_statsd.html
		InfluxDB: func(tags []tag) string {
			var buf bytes.Buffer
			for _, tag := range tags {
				_ = buf.WriteByte(',')
				_, _ = buf.WriteString(tag.K)
				_ = buf.WriteByte('=')
				_, _ = buf.WriteString(tag.V)
			}
			return buf.String()
		},
		// Datadog tag format: |#tag1:value1,tag2:value2
		// http://docs.datadoghq.com/guides/dogstatsd/#datagram-format
		Datadog: func(tags []tag) string {
			buf := bytes.NewBufferString("|#")
			first := true
			for _, tag := range tags {
				if first {
					first = false
				} else {
					_ = buf.WriteByte(',')
				}
				_, _ = buf.WriteString(tag.K)
				_ = buf.WriteByte(':')
				_, _ = buf.WriteString(tag.V)
			}
			return buf.String()
		},
	}
	splitFuncs = map[TagFormat]func(string) []tag{
		InfluxDB: func(s string) []tag {
			s = s[1:]
			pairs := strings.Split(s, ",")
			tags := make([]tag, len(pairs))
			for i, pair := range pairs {
				kv := strings.Split(pair, "=")
				tags[i] = tag{K: kv[0], V: kv[1]}
			}
			return tags
		},
		Datadog: func(s string) []tag {
			s = s[2:]
			pairs := strings.Split(s, ",")
			tags := make([]tag, len(pairs))
			for i, pair := range pairs {
				kv := strings.Split(pair, ":")
				tags[i] = tag{K: kv[0], V: kv[1]}
			}
			return tags
		},
	}
)
