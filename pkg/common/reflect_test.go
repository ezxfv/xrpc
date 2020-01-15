package common_test

import (
	"context"
	"testing"

	"github.com/edenzhong7/xrpc/pkg/common"
	"github.com/edenzhong7/xrpc/protocol/greeter"
	"github.com/stretchr/testify/assert"
)

type Man struct {
	Name string `demo:"name"`
	Age  int    `demo:"age"`
}

func TestReflectDemo_Call(t *testing.T) {
	name := "xxx"
	r := common.R
	var g greeter.GreeterServer
	g = &greeter.UnimplementedGreeterServer{}

	r.RegisterService("greeter", g)
	ctx := context.Background()
	req := &greeter.HelloRequest{
		Name: name,
	}
	rs, err := r.Call("greeter.SayHello", ctx, req)

	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(rs))
	assert.Equal(t, nil, rs[1])
	reply, ok := rs[0].(*greeter.HelloReply)
	assert.True(t, ok)
	assert.Equal(t, "Hi "+name, reply.Message)
}

func TestReflectDemo_Marshal(t *testing.T) {
	m := &Man{
		Name: "xxx",
		Age:  1,
	}
	r := common.R
	bs, err := r.Marshal(m)
	assert.True(t, err == nil)
	assert.Equal(t, []byte("name:`xxx`,age:`1`"), bs)
}

func TestReflectDemo_Unmarshal(t *testing.T) {
	bs := []byte("name:`xxx`,age:`1`")
	m := &Man{}
	r := common.R
	err := r.Unmarshal(bs, m)
	assert.True(t, err == nil)
	assert.True(t, m.Name == "xxx")
	assert.True(t, m.Age == 1)
}
