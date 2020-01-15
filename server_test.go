package xrpc_test

import (
	"testing"

	"github.com/edenzhong7/xrpc"

	pb "github.com/edenzhong7/xrpc/protocol/greeter"
)

func TestServer_RegisterService(t *testing.T) {
	g := &pb.UnimplementedGreeterServer{}
	s := xrpc.NewServer()
	pb.RegisterGreeterServer(s, g)
}
