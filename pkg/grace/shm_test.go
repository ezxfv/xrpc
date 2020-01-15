package grace_test

import (
	"testing"
	"time"

	"github.com/edenzhong7/xrpc/pkg/grace"
	"github.com/stretchr/testify/assert"
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
