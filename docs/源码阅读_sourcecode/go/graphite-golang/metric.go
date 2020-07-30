// Package  mygraphite provides ...
package mygraphite

import (
	"fmt"
	"time"
)

type Metric struct {
	Name      string
	Value     string
	Timestamp int64
}

func NewMetric(name, value string, ts int64) Metric {
	return Metric{
		Name:      name,
		Value:     value,
		Timestamp: ts,
	}
}

// 协议比较简单，就是  "name value timestamp"
func (m Metric) String() string {
	return fmt.Sprintf("%s %s %s", m.Name, m.Value, time.Unix(m.Timestamp, 0).Format("2006-01-02 15:04:05"))
}
