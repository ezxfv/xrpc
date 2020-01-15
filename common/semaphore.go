package common

import (
	"reflect"
)

type Semaphore struct {
	cs   chan struct{}
	maxG int
}

func MakeSemaphore(n int) *Semaphore {
	cs := make(chan struct{}, n)
	return &Semaphore{cs: cs, maxG: n}
}

func (s *Semaphore) MaxG() int {
	return s.maxG
}
func (s *Semaphore) Acquire() {
	s.cs <- struct{}{}
}

func (s *Semaphore) Release() {
	<-s.cs
}

func (s *Semaphore) Go(f interface{}, params ...interface{}) {
	s.Acquire()
	v := reflect.ValueOf(f)
	_call(v, params...)
	s.Release()
}
