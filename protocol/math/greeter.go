package math

import "context"

type HelloRequest struct {
	Name string
}

type HelloReply struct {
	Msg string
}

// GreeterServer is the server API for Greeter service.
type Greeter interface {
	// Sends a greeting
	SayHello(context.Context, *HelloRequest) (*HelloReply, error)
	SayHi(context.Context, *HelloRequest) (*HelloReply, error)
}
