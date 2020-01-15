package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/edenzhong7/xrpc"
	"github.com/edenzhong7/xrpc/net"

	"github.com/urfave/cli/v2"

	pb "github.com/edenzhong7/xrpc/protocol/greeter"
)

type GreeterImpl struct {
	pb.UnimplementedGreeterServer
}

func (g *GreeterImpl) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	reply := &pb.HelloReply{
		Message: "Hi " + req.GetName(),
	}
	return reply, nil
}

func startServer(ctx *cli.Context) {
	protocol := ctx.String("protocol")
	addr := ctx.String("addr")
	lis, err := net.Listen(protocol, addr)
	if err != nil {
		log.Fatal(err)
	}
	s := xrpc.NewServer()
	pb.RegisterGreeterServer(s, &GreeterImpl{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "addr",
				Value: "127.0.0.1:9090",
				Usage: "addr to listen",
			},
			&cli.StringFlag{
				Name:  "protocol",
				Value: "kcp",
				Usage: "net protocol",
			},
		},
		Action: func(c *cli.Context) error {
			fmt.Printf("Serving on %s://%s\n", c.String("protocol"), c.String("addr"))
			startServer(c)
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
