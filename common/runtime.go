package common

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

var enableTrace = true

func myName(depth ...int) string {
	d := 0
	if depth != nil {
		d = depth[0]
	}
	pc, _, _, _ := runtime.Caller(1 + d)
	return runtime.FuncForPC(pc).Name()
}

func callerName(depth ...int) string {
	d := 0
	if depth != nil {
		d = depth[0]
	}
	pc, _, _, _ := runtime.Caller(2 + d)
	return runtime.FuncForPC(pc).Name()
}

func GoID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

func Trace() func() {
	if !enableTrace {
		return func() {}
	}
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	id := GoID()
	caller := callerName(1)
	fmt.Printf("%s:%d %s ---> %s [go routine:%d]\n", file, line, caller, f.Name(), id)
	return func() {
		fmt.Printf(" %s <--- %s [go routine:%d]\n", caller, f.Name(), id)
	}
}

func Trace2() {
	pc := make([]uintptr, 10) // at least 1 entry needed
	n := runtime.Callers(0, pc)
	frames := runtime.CallersFrames(pc[:n])
	for {
		frame, more := frames.Next()
		fmt.Printf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
		if !more {
			break
		}
	}
}

func DumpStacks() {
	buf := make([]byte, 16384)
	buf = buf[:runtime.Stack(buf, true)]
	fmt.Printf("=== BEGIN goroutine stack dump ===\n%s\n=== END goroutine stack dump ===", buf)
}
