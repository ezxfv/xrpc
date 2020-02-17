package chord_test

import (
	"fmt"
	"log"
	"testing"
	"time"

	"x.io/xrpc/app/chord"
)

var (
	host = "localhost"
)

// port: rpc server, port+1: http api, port+2: prom metrics
func TestNewChord(t *testing.T) {
	h := chord.NewBlake2bHasher()
	store := chord.NewSimpleKVStore()
	port := 9899
	c := chord.NewChord(host, port, h, store)
	var apiAddr = fmt.Sprintf("%s:%d", host, port+1)
	go chord.ServerAPI(apiAddr, chord.NewChordAPI(c))

	if err := c.Server(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func TestChordImpl_Join(t *testing.T) {
	nodeNum := 2
	startJoin := func(port int) {
		h := chord.NewBlake2bHasher()
		store := chord.NewSimpleKVStore()
		c := chord.NewChord(host, port, h, store)
		var joinAddr = fmt.Sprintf("%s:%d", host, port-100)
		c.JoinNode(joinAddr)
		if err := c.Server(); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}
	for i := 1; i <= nodeNum; i++ {
		go startJoin(9899 + i*100)
	}
	for {
		time.Sleep(time.Second)
	}
}
