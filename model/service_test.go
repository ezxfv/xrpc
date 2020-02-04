package model_test

import (
	"testing"

	"github.com/edenzhong7/xrpc/model"
	jsoniter "github.com/json-iterator/go"
)

type Math struct{}

type Human struct {
	Age int
}

func (m *Math) Add(a, b *int, h Human) (*int, *int, *Human) {
	c := *a + *b
	d := 2 * c
	h.Age += 20
	return &c, &d, &h
}

func Sub(a, b int) int {
	return a - b
}

func Add(a, b int) int {
	return a + b
}

func TestCustomService_Call(t *testing.T) {
	cs := model.NewCustomService()
	cs.RegisterService("math", &Math{})
	var (
		a = 1
		b = 2
		h = Human{Age: 10}
		c = 0
		d = 0
	)
	args := []interface{}{&a, &b, &h}
	data, _ := jsoniter.Marshal(args)
	res, _ := cs.Call("math.Add", data)
	model.Dispatch(res, &c, &d, &h)
	println(c, d, h.Age)
}

func TestCustomService_RegisterFunction(t *testing.T) {
	cs := model.NewCustomService()
	cs.RegisterFunction("math", "Sub", Sub)
	cs.RegisterFunction("math", "Add", Add)
	var (
		a = 1
		b = 2
		c = 0
	)
	args := []interface{}{&a, &b}
	data, _ := jsoniter.Marshal(args)
	res, _ := cs.Call("math.Sub", data)
	model.Dispatch(res, &c)
	println(c)
	res, _ = cs.Call("math.Add", data)
	model.Dispatch(res, &c)
	println(c)
}

func BenchmarkCustomService_RegisterFunction(b *testing.B) {
	cs := model.NewCustomService()
	cs.RegisterFunction("math", "Sub", Sub)
	var (
		a1 = 1
		a2 = 2
		c  = 0
	)
	for i := 0; i < b.N; i++ {
		args := []interface{}{&a1, &a2}
		data, _ := jsoniter.Marshal(args)
		res, _ := cs.Call("math.Sub", data)
		model.Dispatch(res, &c)
	}
}
