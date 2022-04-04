package main

import (
	"flag"
	"log"
	mystatsd "test/source_code_read/statsd"
)

var addr = flag.String("addr", ":8125", "Input Your statsd addr")

// ./example --addr=127.0.0.1:8125
func main() {
	flag.Parse()
	c, err := mystatsd.New(
		mystatsd.Address(*addr),
	)
	if err != nil {
		log.Print(err)
	}
	defer c.Close()

	for i := 0; i < 100; i++ {
		c.Increment("foo.bar.10_228_100_3.stats")
	}
}
