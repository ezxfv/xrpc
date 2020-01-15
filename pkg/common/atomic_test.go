package common_test

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/edenzhong7/xrpc/pkg/common"
	"github.com/stretchr/testify/assert"
)

var n int32 = 0
var mu = &sync.Mutex{}

func Hello(x int) {
	atomic.AddInt32(&n, 1)
}

func MutexIncN(wg *sync.WaitGroup) {
	mu.Lock()
	n++
	mu.Unlock()
	wg.Done()
}

func CASIncN(wg *sync.WaitGroup) {
	for {
		val := atomic.LoadInt32(&n)
		if atomic.CompareAndSwapInt32(&n, val, val+1) {
			wg.Done()
			return
		}
	}
}

func TestSemaphore_Go(t *testing.T) {
	s := common.S
	num := 10000
	for i := 0; i < num; i++ {
		s.Go(Hello, i)
		assert.True(t, runtime.NumGoroutine() < s.MaxG())
	}
	assert.Equal(t, int32(num), n)
}

func TestSyncCond(t *testing.T) {
	c := sync.NewCond(&sync.Mutex{})
	queue := make([]interface{}, 0, 10)
	removeFromQueue := func(delay time.Duration) {
		time.Sleep(delay)
		c.L.Lock()
		queue = queue[1:]
		fmt.Println("Removed from queue")
		c.L.Unlock()
		c.Signal()
	}
	for i := 0; i < 10; i++ {
		c.L.Lock()
		for len(queue) == 2 {
			c.Wait()
		}
		fmt.Println("Adding to queue")
		queue = append(queue, struct{}{})
		go removeFromQueue(1 * time.Second)
		c.L.Unlock()
	}
}

func TestMutexIncN(t *testing.T) {
	N := 10000000
	wg := &sync.WaitGroup{}
	wg.Add(N)
	for i := 0; i < N; i++ {
		go MutexIncN(wg)
	}
	wg.Wait()
	assert.True(t, n == int32(N))
}

func TestCASIncN(t *testing.T) {
	N := 10000000
	wg := &sync.WaitGroup{}
	wg.Add(N)
	for i := 0; i < N; i++ {
		go CASIncN(wg)
	}
	wg.Wait()
	assert.True(t, n == int32(N))
}
