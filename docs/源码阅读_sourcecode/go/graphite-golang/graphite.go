package mygraphite

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/uber/jaeger-client-go/crossdock/log"
)

type Graphite struct {
	Host       string
	Port       int
	Protocol   string
	Timeout    time.Duration
	Prefix     string
	conn       net.Conn
	nop        bool
	DisableLog bool
}

const defaultTimeout = 5 * time.Second //5 second

func (g *Graphite) IsNop() bool {
	return g.nop
}

func (g *Graphite) Connect() error {
	if !g.IsNop() {
		if g.conn != nil {
			g.conn.Close()
		}

		addr := fmt.Sprintf("%s:%d", g.Host, g.Port)
		if g.Timeout == 0 {
			g.Timeout = defaultTimeout
		}

		var err error
		var conn net.Conn

		if g.Protocol == "udp" {
			udpAddr, err := net.ResolveUDPAddr("udp", addr)
			if err != nil {
				return err
			}
			conn, err = net.DialUDP(g.Protocol, nil, udpAddr)
		} else {
			conn, err = net.DialTimeout(g.Protocol, addr, g.Timeout)
		}
		if err != nil {
			return err
		}
		g.conn = conn
	}
	return nil
}

func (g *Graphite) Disconnect() error {
	err := g.conn.Close()
	g.conn = nil
	return err
}

func (g *Graphite) SendMetric(metric Metric) error {
	metrics := make([]Metric, 1)
	metrics[0] = metric
	return g.sendMetrics(metrics)
}

func (g *Graphite) SendMetrics(metrics []Metric) error {
	return g.sendMetrics(metrics)
}

func (g *Graphite) sendMetrics(metrics []Metric) error {
	if g.IsNop() {
		if !g.DisableLog {
			for _, m := range metrics {
				log.Printf("Graphite: %s\n", m)
			}
		}
		return nil
	}

	zeroedMetric := Metric{}
	buf := bytes.NewBufferString("")
	for _, m := range metrics {
		if m == zeroedMetric {
			continue // 跳过空
		}

		if m.Timestamp == 0 {
			m.Timestamp = time.Now().Unix()
		}
		metricName := ""
		if g.Prefix != "" {
			metricName = fmt.Sprintf("%s.%s", g.Prefix, m.Name)
		} else {
			metricName = m.Name
		}
		if g.Protocol == "udp" {
			fmt.Fprintf(g.conn, "%s %s %d\n", metricName, m.Value, m.Timestamp) // 到 g.conn
			continue
		}
		buf.WriteString(fmt.Sprintf("%s %s %d\n", metricName, m.Value, m.Timestamp)) // append buf
	}
	if g.Protocol == "tcp" {
		_, err := g.conn.Write(buf.Bytes()) // write buf
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *Graphite) SimpleSend(stat string, value string) error {
	metrics := make([]Metric, 1)
	metrics[0] = NewMetric(stat, value, time.Now().Unix())
	return g.sendMetrics(metrics)
}

// 工厂方法

func GraphiteFactory(protocol, host string, port int, prefix string) (*Graphite, error) {
	var g *Graphite

	switch protocol {
	case "tcp":
		g = &Graphite{Host: host, Port: port, Protocol: "tcp", Prefix: prefix}
	case "udp":
		g = &Graphite{Host: host, Port: port, Protocol: "udp", Prefix: prefix}
	case "nop":
		g = &Graphite{Host: host, Port: port, nop: true}
	}
	err := g.Connect()
	if err != nil {
		return nil, err
	}
	return g, nil
}

func NewGraphite(host string, port int) (*Graphite, error) {
	return GraphiteFactory("tcp", host, port, "")
}

func NewGraphiteWithMetricPrefix(host string, port int, prefix string) (*Graphite, error) {
	return GraphiteFactory("tcp", host, port, prefix)
}

func NewGraphiteUDP(host string, port int) (*Graphite, error) {
	return GraphiteFactory("udp", host, port, "")
}

func NewGraphiteNop(host string, port int) *Graphite {
	g, _ := GraphiteFactory("nop", host, port, "")
	return g
}
