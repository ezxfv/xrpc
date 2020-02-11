package xrpc_test

import (
	"testing"

	"github.com/edenzhong7/xrpc"

	jsoniter "github.com/json-iterator/go"
)

type Math struct{}

func (m *Math) Add(a, b int) int {
	return a + b
}

func (m *Math) Calc(a, b int) (int, float64) {
	return a * b, float64(a) / float64(b)
}

func Sub(a, b int) int {
	return a - b
}

func Add(a, b int) int {
	return a + b
}

func TestCustomService_Call(t *testing.T) {
	cs := xrpc.NewCustomServer()
	cs.RegisterCustomService("math", &Math{})
	var (
		a = 1
		b = 2
		c = 0
		d = 0
	)
	args := []interface{}{a, b}
	data, _ := jsoniter.Marshal(args)
	res, _ := cs.DirectCall("math.Add", data)
	xrpc.Dispatch(res, &c, &d)
}

func TestCustomService_RegisterFunction(t *testing.T) {
	cs := xrpc.NewCustomServer()
	cs.RegisterFunction("math", "Sub", Sub)
	cs.RegisterFunction("math", "Add", Add)
	var (
		a = 1
		b = 2
		c = 0
	)
	args := []interface{}{&a, &b}
	data, _ := jsoniter.Marshal(args)
	res, _ := cs.DirectCall("math.Sub", data)
	xrpc.Dispatch(res, &c)
	println(c)
	res, _ = cs.DirectCall("math.Add", data)
	xrpc.Dispatch(res, &c)
	println(c)
}

func BenchmarkCustomService_RegisterFunction(b *testing.B) {
	cs := xrpc.NewCustomServer()
	cs.RegisterFunction("math", "Sub", Sub)
	var (
		a1 = 1
		a2 = 2
		c  = 0
	)
	for i := 0; i < b.N; i++ {
		args := []interface{}{&a1, &a2}
		data, _ := jsoniter.Marshal(args)
		res, _ := cs.DirectCall("math.Sub", data)
		xrpc.Dispatch(res, &c)
	}
}
