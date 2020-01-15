package common

import "sync"

var (
	S *Semaphore
	R *ReflectDemo

	once = &sync.Once{}
)

func init() {
	once.Do(func() {
		S = MakeSemaphore(1000)
		R = NewReflectDemo()
	})
}
