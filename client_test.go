package xrpc_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"x.io/xrpc"
	_ "x.io/xrpc/pkg/encoding/gzip"
	_ "x.io/xrpc/pkg/encoding/json"
	_ "x.io/xrpc/pkg/encoding/proto"
	_ "x.io/xrpc/pkg/encoding/snappy"
	"x.io/xrpc/plugin/crypto"

	"github.com/stretchr/testify/assert"
	greeter_pb "x.io/xrpc/protocol/greeter"
	pb "x.io/xrpc/protocol/math"
)

var (
	sessionID  = "session_math_0"
	sessionKey = "1234"

	user = "admin"
	pass = "1234"
	ctx  = context.Background()
)

func setupConn(conn *xrpc.ClientConn) {
	cryptoPlugin := crypto.New()
	cryptoPlugin.SetKey(sessionID, sessionKey)
	conn.ApplyPlugins(cryptoPlugin)

	conn.SetHeaderArg("user", user)
	conn.SetHeaderArg("pass", pass)
	conn.SetHeaderArg(crypto.Key, sessionID)
}

func newMathClient(protocol, addr string) pb.MathClient {
	conn, err := xrpc.Dial(protocol, addr, xrpc.WithInsecure(), xrpc.WithJsonCodec())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	setupConn(conn)

	client := pb.NewMathClient(conn)
	return client
}

func newGreeterClient(protocol, addr string) greeter_pb.GreeterClient {
	conn, err := xrpc.Dial(protocol, addr, xrpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	setupConn(conn)

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

func TestCustomClientTrace(t *testing.T) {
	client, err := xrpc.NewRawClient("tcp", serverAddr, xrpc.WithJsonCodec())
	assert.Equal(t, nil, err)
	client.Setup(setupConn)
	var c int
	xctx := xrpc.XBackground()
	err = client.RawCall(xctx, "default.RpcDouble", &c, 10)
	assert.Equal(t, nil, err)
	assert.Equal(t, 20, c)
}

func TestCustomClient(t *testing.T) {
	client, err := xrpc.NewRawClient("tcp", serverAddr, xrpc.WithJsonCodec())
	assert.Equal(t, nil, err)
	client.Setup(setupConn)

	var c int
	err = client.RawCall(ctx, "math.Add", &c, 1, 2)
	assert.Equal(t, nil, err)
	assert.Equal(t, 3, c)

	err = client.RawCall(ctx, "default.Double", &c, 10)
	assert.Equal(t, nil, err)
	assert.Equal(t, 20, c)

	var d int
	var f float64
	reply := []interface{}{&d, &f}
	err = client.RawCall(ctx, "math.Calc", &reply, 4, 2)
	assert.Equal(t, nil, err)
	assert.Equal(t, 8, d)
	assert.Equal(t, 2.0, f)
}

func TestMathClientTrace(t *testing.T) {
	if client == nil {
		client = newMathClient("tcp", serverAddr)
	}

	r := client.XRpcDouble(ctx, 10)
	assert.Equal(t, 20, r)
}

func TestMathClient(t *testing.T) {
	if client == nil {
		client = newMathClient("tcp", serverAddr)
	}

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

func TestChordMathClient(t *testing.T) {
	N := 3
	for i := 0; i < N; i++ {
		client := newMathClient("chord", "math.Math")
		r := client.XRpcDouble(ctx, 10)
		assert.Equal(t, 20, r)
	}
}

func TestGreeterClient(t *testing.T) {
	client := newGreeterClient("tcp", serverAddr)
	r, err := client.SayHello(context.Background(), &greeter_pb.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())
}

func TestMathAndGreeterClient(t *testing.T) {
	mathClient := newMathClient("tcp", serverAddr)
	greeterClient := newGreeterClient("tcp", serverAddr)
	c := mathClient.Add(ctx, 2, 3)
	assert.Equal(t, 5, c)

	r, err := greeterClient.SayHello(ctx, &greeter_pb.HelloRequest{Name: name})
	assert.Equal(t, nil, err)
	assert.Equal(t, "Hi xrpc_test!!", r.GetMessage())
}

func TestMathClient1K(t *testing.T) {
	if client == nil {
		client = newMathClient("tcp", serverAddr)
	}
	now := time.Now()
	N := 1000
	for i := 0; i < N; i++ {
		client.Add(ctx, a, b)
	}
	fmt.Printf("%2f ms/op\n", float64(time.Since(now).Milliseconds())/float64(N))
}

func BenchmarkMathClientInc(tb *testing.B) {
	if client == nil {
		client = newMathClient("tcp", serverAddr)
	}
	for i := 0; i < tb.N; i++ {
		client.Inc(ctx, n)
	}
}

func BenchmarkMathClientAdd(tb *testing.B) {
	if client == nil {
		client = newMathClient("tcp", serverAddr)
	}
	for i := 0; i < tb.N; i++ {
		client.Add(ctx, a, b)
	}
}

func BenchmarkMathClientCalc(tb *testing.B) {
	if client == nil {
		client = newMathClient("tcp", serverAddr)
	}
	for i := 0; i < tb.N; i++ {
		client.Calc(ctx, a, b)
	}
}
