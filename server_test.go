package xrpc_test

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	"x.io/xrpc/plugin/chord"

	"x.io/xrpc/api"

	"x.io/xrpc"
	"x.io/xrpc/pkg/net"
	"x.io/xrpc/plugin/crypto"
	"x.io/xrpc/plugin/prom"
	"x.io/xrpc/plugin/trace"
	"x.io/xrpc/plugin/whitelist"

	greeter_pb "x.io/xrpc/protocol/greeter"
	math_pb "x.io/xrpc/protocol/math"
)

const (
	serverAddr = "localhost:9898"
)

var (
	enablePlugin = true
	enableAuth   = true
	enableCrypto = true
	enableAPI    = true
	enableChord  = true
)

type MathImpl struct {
	math_pb.UnimplementedMath
}

func (m *MathImpl) XRpcDouble(c *xrpc.XContext, a int) int {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
	conn, err := xrpc.Dial("tcp", serverAddr, xrpc.WithInsecure(), xrpc.WithJsonCodec())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	setupConn(conn)
	client := math_pb.NewMathClient(conn)

	client.Double(c.Context(), a)
	return client.XRpcAdd(c.Context(), a, a)
}

func (m *MathImpl) XRpcAdd(c *xrpc.XContext, a, b int) int {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
	conn, err := xrpc.Dial("tcp", serverAddr, xrpc.WithInsecure(), xrpc.WithJsonCodec())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	setupConn(conn)
	client := math_pb.NewMathClient(conn)

	return client.Add(c.Context(), a, b)
}

func (m *MathImpl) Double(a int) int {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
	return a * 2
}

func (m *MathImpl) Inc(n *math_pb.Num) (int32, *math_pb.Num) {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
	n.Val += n.S.S
	return n.Val, n
}

func (m *MathImpl) Add(a, b int) int {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
	return a + b
}

func (m *MathImpl) Calc(ns ...int) (int, float64) {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
	var sum int
	for _, i := range ns {
		sum += i
	}
	return sum, float64(sum) / float64(len(ns))
}

type GreeterImpl struct {
	greeter_pb.UnimplementedGreeterServer
}

func (g *GreeterImpl) SayHello(ctx context.Context, req *greeter_pb.HelloRequest) (*greeter_pb.HelloReply, error) {
	reply := &greeter_pb.HelloReply{
		Message: "Hi " + req.GetName(),
	}
	return reply, nil
}

func newServer(protocol, addr string) (lis net.Listener, svr *xrpc.Server) {
	lis, err := net.Listen(context.Background(), protocol, addr)
	if err != nil {
		log.Fatal(err)
	}
	s := xrpc.NewServer()
	if enablePlugin {
		promPlugin := prom.New()
		//logPlugin := logp.New()
		tracePlugin := trace.New()
		whitelistPlugin := whitelist.New(map[string]bool{"127.0.0.1": true}, nil)
		s.ApplyPlugins(promPlugin, tracePlugin, whitelistPlugin)
		if enableCrypto {
			cryptoPlugin := crypto.New()
			cryptoPlugin.SetKey(sessionID, sessionKey)
			s.ApplyPlugins(cryptoPlugin)
		}
		if enableChord {
			chordPlugin := chord.New(fmt.Sprintf("%s://%s", protocol, addr), "http://localhost:9900/chord")
			s.ApplyPlugins(chordPlugin)
		}
		s.StartPlugins()
	}
	if enableAuth {
		admin := xrpc.NewAdminAuthenticator(user, pass)
		s.SetAuthenticator(admin)
	}
	if enableAPI {
		go api.Server(":8080")
	}
	return lis, s
}

func TestCustomServer(t *testing.T) {
	lis, s := newServer("tcp", serverAddr)
	s.RegisterCustomService("math", &Math{})
	s.RegisterFunction("default", "Double", func(a int) int {
		return a * 2
	})
	s.RegisterFunction("default", "Add", func(a, b int) int {
		return a + b
	})
	s.RegisterFunction("default", "RpcDouble", func(c *xrpc.XContext, a int) int {
		client, err := xrpc.NewRawClient("tcp", serverAddr, xrpc.WithJsonCodec())
		if err != nil {
			return 0
		}
		client.Setup(setupConn)

		var r1, r2 int
		err = client.RawCall(c.Context(), "default.Double", &r1, a)
		err = client.RawCall(c.Context(), "default.Add", &r2, a, a)
		return r1
	})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	s.Start()
}

func TestMathServer(t *testing.T) {
	lis, s := newServer("tcp", serverAddr)
	math_pb.RegisterMathServer(s, &MathImpl{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	s.Start()
}

func TestGreeterServer(t *testing.T) {
	lis, s := newServer("tcp", serverAddr)
	greeter_pb.RegisterGreeterServer(s, &GreeterImpl{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	s.Start()
}

func TestMathAndGreeterServer(t *testing.T) {
	lis, s := newServer("tcp", serverAddr)
	math_pb.RegisterMathServer(s, &MathImpl{})
	greeter_pb.RegisterGreeterServer(s, &GreeterImpl{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	s.Start()
}

func TestServer_RegisterService(t *testing.T) {
	g := &greeter_pb.UnimplementedGreeterServer{}
	s := xrpc.NewServer()
	greeter_pb.RegisterGreeterServer(s, g)
}
