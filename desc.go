package xrpc

import "google.golang.org/grpc"

type (
	ServiceDesc = grpc.ServiceDesc
	MethodDesc  = grpc.MethodDesc
	StreamDesc  = grpc.StreamDesc
)
