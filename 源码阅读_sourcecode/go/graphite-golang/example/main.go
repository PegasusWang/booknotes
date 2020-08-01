// TODO 机器上测试一下
package main

import (
	"flag"
	"fmt"
	"log"
	graphite "test/source_code_read/graphite-golang"
)

func main() {
	var host = flag.String("host", "localhost", "Input Your Name")
	var port = flag.Int("port", 2333, "Input Your Name")
	flag.Parse()
	fmt.Println(*host, *port)

	g, err := graphite.NewGraphite(*host, *port)
	if err != nil {
		log.Print(err)
	}

	log.Printf("Loaded Graphite connection: %#v", g)
	g.SimpleSend("stats.graphite_loaded", "1")
}
