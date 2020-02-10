package xrpc_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/edenzhong7/xrpc"
	_ "github.com/edenzhong7/xrpc/pkg/encoding/gzip"
	_ "github.com/edenzhong7/xrpc/pkg/encoding/json"
	_ "github.com/edenzhong7/xrpc/pkg/encoding/proto"
	_ "github.com/edenzhong7/xrpc/pkg/encoding/snappy"

	greeter_pb "github.com/edenzhong7/xrpc/protocol/greeter"
	pb "github.com/edenzhong7/xrpc/protocol/math"
	"github.com/stretchr/testify/assert"
)

func newMathClient(protocol, addr string) pb.MathClient {
	conn, err := xrpc.Dial(protocol, addr, xrpc.WithInsecure(), xrpc.WithJsonCodec())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	client := pb.NewMathClient(conn)
	return client
}

func newGreeterClient(protocol, addr string) greeter_pb.GreeterClient {
	conn, err := xrpc.Dial(protocol, addr, xrpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	client := greeter_pb.NewGreeterClient(conn)
	return client
}

var (
	client pb.MathClient
	a      = 3
	b      = 2
	n      = &pb.Num{Val: 10, S: pb.Step{S: 10}}

	name = "xrpc_test!!"
)

func TestMathClient(t *testing.T) {
	if client == nil {
		client = newMathClient("tcp", "localhost:9898")
	}
	ctx := context.Background()
	//ctx = xrpc.SetCookie(ctx, "endpoint", "client")
	r := client.Add(ctx, a, b)
	assert.Equal(t, 5, r)
	log.Printf("3 + 2 = %d", r)
	sum, avg := client.Calc(ctx, a, b)
	assert.Equal(t, 5, sum)
	assert.Equal(t, 2.50, avg)
	log.Printf("sum = %d, avg = %.2f", sum, avg)
	val, n := client.Inc(ctx, n)
	assert.Equal(t, val, n.Val)
	assert.Equal(t, int32(20), val)
	log.Printf("new num: %d", n.Val)
}

func TestGreeterClient(t *testing.T) {
	client := newGreeterClient("tcp", "localhost:9898")
	// Contact the server and print out its response.
	r, err := client.SayHello(context.Background(), &greeter_pb.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())
}

func TestMathAndGreeterClient(t *testing.T) {
	mathClient := newMathClient("tcp", "localhost:9898")
	greeterClient := newGreeterClient("tcp", "localhost:9898")
	c := mathClient.Add(context.Background(), 2, 3)
	assert.Equal(t, 5, c)

	r, err := greeterClient.SayHello(context.Background(), &greeter_pb.HelloRequest{Name: name})
	assert.Equal(t, nil, err)
	assert.Equal(t, "Hi xrpc_test!!", r.GetMessage())
}

func TestMathClient1K(t *testing.T) {
	if client == nil {
		client = newMathClient("tcp", "localhost:9898")
	}
	now := time.Now()
	N := 1000
	for i := 0; i < N; i++ {
		client.Add(context.Background(), a, b)
		client.Calc(context.Background(), a, b)
		client.Inc(context.Background(), n)
	}
	fmt.Printf("%2f ms/op", float64(time.Since(now).Milliseconds())/float64(3*N))
}

func BenchmarkMathClientInc(tb *testing.B) {
	if client == nil {
		client = newMathClient("tcp", "localhost:9898")
	}
	for i := 0; i < tb.N; i++ {
		client.Inc(context.Background(), n)
	}
}

func BenchmarkMathClientAdd(tb *testing.B) {
	if client == nil {
		client = newMathClient("tcp", "localhost:9898")
	}
	for i := 0; i < tb.N; i++ {
		client.Add(context.Background(), a, b)
	}
}

func BenchmarkMathClientCalc(tb *testing.B) {
	if client == nil {
		client = newMathClient("tcp", "localhost:9898")
	}
	for i := 0; i < tb.N; i++ {
		client.Calc(context.Background(), a, b)
	}
}
