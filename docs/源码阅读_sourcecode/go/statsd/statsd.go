// Package mystatsd provides ...
package mystatsd

import "time"

type Client struct {
	conn   *conn // 发送
	muted  bool
	rate   float32
	prefix string
	tags   string
}

func New(opts ...Option) (*Client, error) {
	conf := &config{
		Client: clientConfig{
			Rate: 1,
		},
		Conn: connConfig{
			Addr:          ":8125",
			FlushPeriod:   100 * time.Millisecond,
			MaxPacketSize: 1440,
			Network:       "udp", // 默认 udp
		},
	}
	for _, o := range opts {
		o(conf)
	}
	conn, err := newConn(conf.Conn, conf.Client.Muted)
	c := &Client{
		conn:  conn,
		muted: conf.Client.Muted,
	}
	if err != nil {
		c.muted = true
		return c, err
	}

	c.rate = conf.Client.Rate
	c.prefix = conf.Client.Prefix
	c.tags = joinTags(conf.Conn.TagFormat, conf.Client.Tags)
	return c, nil
}

// 以下基本都是调用 conn.metric 方法了
func (c *Client) Count(bucket string, n interface{}) {
	if c.skip() {
		return
	}
	c.conn.metric(c.prefix, bucket, n, "c", c.rate, c.tags)
}

func (c *Client) skip() bool {
	return c.muted || (c.rate != 1 && randFloat() > c.rate)
}

func (c *Client) Increment(bucket string) {
	c.Count(bucket, 1)
}

func (c *Client) Gauge(bucket string, value interface{}) {
	if c.skip() {
		return
	}
	c.conn.gauge(c.prefix, bucket, value, c.tags)
}

func (c *Client) Timing(bucket string, value interface{}) {
	if c.skip() {
		return
	}
	c.conn.metric(c.prefix, bucket, value, "ms", c.rate, c.tags)
}

func (c *Client) Histogram(bucket string, value interface{}) {
	if c.skip() {
		return
	}
	c.conn.metric(c.prefix, bucket, value, "h", c.rate, c.tags)
}

// 简化发送 timing
type Timing struct {
	start time.Time
	c     *Client
}

func (c *Client) NewTiming() Timing {
	return Timing{start: now(), c: c}
}

func (t Timing) Send(bucket string) {
	t.c.Timing(bucket, int(t.Duration()/time.Millisecond))
}

func (t Timing) Duration() time.Duration {
	return now().Sub(t.start)
}

func (c *Client) Unique(bucket string, value string) {
	if c.skip() {
		return
	}
	c.conn.unique(c.prefix, bucket, value, c.tags)
}

// flush client's buffer
func (c *Client) Flush() {
	if c.muted {
		return
	}
	c.conn.mu.Lock()
	c.conn.flush(0)
	c.conn.mu.Unlock()
}

func (c *Client) Close() {
	if c.muted {
		return
	}
	c.conn.mu.Lock()
	c.conn.flush(0)
	c.conn.handleError(c.conn.w.Close())
	c.conn.closed = true
	c.conn.mu.Unlock()
}
