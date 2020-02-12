package grace_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"x.io/xrpc/pkg/grace"
)

const (
	id   int = 10
	val      = 100
	exit     = -1
)

func TestShmWrite(t *testing.T) {
	grace.ShmSet(id, val)
}

func TestShmRead(t *testing.T) {
	assert.Equal(t, val, grace.ShmGet(id))
}

func TestShmDel(t *testing.T) {
	assert.Equal(t, "errno 0", grace.ShmDel(id))
}

func TestStart(t *testing.T) {
	grace.ShmSet(id, val)
	for {
		time.Sleep(time.Millisecond * 100)
		if grace.ShmGet(id) == exit {
			break
		}
	}
	grace.ShmDel(id)
}

func TestStop(t *testing.T) {
	time.Sleep(time.Second)
	grace.ShmSet(id, exit)
}

func TestRunWithActions(t *testing.T) {
	c := make(chan struct{})
	s := struct{}{}
	go grace.RunWithActions(id, map[int]func(){
		grace.Start: func() {
			println("start")
			c <- s
		},
		grace.Reload: func() {
			println("reload")
			c <- s
		},
		grace.Restart: func() {
			println("restart")
			c <- s
		},
		grace.Stop: func() {
			println("stop")
			c <- s
		},
	})
	grace.ShmSet(id, grace.Start)
	<-c
	grace.ShmSet(id, grace.Reload)
	<-c
	grace.ShmSet(id, grace.Restart)
	<-c
	grace.ShmSet(id, grace.Stop)
	<-c
	close(c)
}
