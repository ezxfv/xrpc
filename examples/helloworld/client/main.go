package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/edenzhong7/xrpc"
	"github.com/urfave/cli/v2"

	pb "github.com/edenzhong7/xrpc/protocol/greeter"
)

func startClient(c *cli.Context) {
	protocol := c.String("protocol")
	addr := c.String("addr")

	conn, err := xrpc.Dial(protocol, addr, xrpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	name := "xrpc_test"
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := client.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())
}

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "addr",
				Value: "127.0.0.1:9090",
				Usage: "service addr",
			},
			&cli.StringFlag{
				Name:  "protocol",
				Value: "kcp",
				Usage: "net protocol",
			},
		},
		Action: func(c *cli.Context) error {
			fmt.Printf("call service on %s://%s\n", c.String("protocol"), c.String("addr"))
			startClient(c)
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
