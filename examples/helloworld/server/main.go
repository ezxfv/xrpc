package main

import (
	"context"
	"fmt"
	"log"
	"os"

	cli "github.com/urfave/cli/v2"

	"github.com/edenzhong7/xrpc"
	_ "github.com/edenzhong7/xrpc/pkg/encoding/gzip"
	_ "github.com/edenzhong7/xrpc/pkg/encoding/proto"
	"github.com/edenzhong7/xrpc/pkg/net"

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
	s.Start()
}

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "addr",
				Value: ":9898",
				Usage: "addr to listen",
			},
			&cli.StringFlag{
				Name:  "protocol",
				Value: "tcp",
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
