package xrpc_test

import (
	"context"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/edenzhong7/xrpc/api"

	"github.com/edenzhong7/xrpc"
	"github.com/edenzhong7/xrpc/pkg/net"
	"github.com/edenzhong7/xrpc/plugin/crypto"
	"github.com/edenzhong7/xrpc/plugin/logp"
	"github.com/edenzhong7/xrpc/plugin/prom"
	"github.com/edenzhong7/xrpc/plugin/trace"
	"github.com/edenzhong7/xrpc/plugin/whitelist"

	greeter_pb "github.com/edenzhong7/xrpc/protocol/greeter"
	math_pb "github.com/edenzhong7/xrpc/protocol/math"
)

var (
	enablePlugin = true
	enableAuth   = true
	enableCrypto = true
	enableAPI    = true
)

type MathImpl struct {
	math_pb.UnimplementedMath
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
		logPlugin := logp.New()
		tracePlugin := trace.New()
		whitelistPlugin := whitelist.New(map[string]bool{"127.0.0.1": true}, nil)
		s.ApplyPlugins(promPlugin, logPlugin, tracePlugin, whitelistPlugin)
		if enableCrypto {
			cryptoPlugin := crypto.New()
			cryptoPlugin.SetKey(sessionID, sessionKey)
			s.ApplyPlugins(cryptoPlugin)
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
	lis, s := newServer("tcp", "localhost:9898")
	s.RegisterCustomService("math", &Math{})
	s.RegisterFunction("default", "Double", func(a int) int {
		return a * 2
	})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	s.Start()
}

func TestMathServer(t *testing.T) {
	lis, s := newServer("tcp", "localhost:9898")
	math_pb.RegisterMathServer(s, &MathImpl{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	s.Start()
}

func TestGreeterServer(t *testing.T) {
	lis, s := newServer("tcp", "localhost:9898")
	greeter_pb.RegisterGreeterServer(s, &GreeterImpl{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	s.Start()
}

func TestMathAndGreeterServer(t *testing.T) {
	lis, s := newServer("tcp", "localhost:9898")
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
