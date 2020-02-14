package main

import (
	"flag"
	"fmt"
	"log"

	"x.io/xrpc/app/chord"
)

var (
	host = flag.String("host", "localhost", "host ip")
	port = flag.Int("port", 9899, "chord serer port")
	join = flag.String("join", "", "join node addr")
	api  = flag.Bool("enable_api", true, "enable http api")
)

func main() {
	flag.Parse()
	h := chord.NewBlake2bHasher()
	store := chord.NewSimpleKVStore()
	c := chord.NewChord(*host, *port, h, store)
	//*join = "localhost:9100"
	if len(*join) > 0 {
		c.JoinNode(*join)
	}
	if *api {
		var apiAddr = fmt.Sprintf("%s:%d", *host, *port+1)
		go chord.ServerAPI(apiAddr, chord.NewChordAPI(c))
	}
	if err := c.Server(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
