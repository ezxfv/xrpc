package types

import (
	"context"
)

type (
	MD              map[string][]string
	UnaryServerInfo struct {
		// Server is the service implementation the user provides. This is read-only.
		Server interface{}
		// FullMethod is the full RPC method string, i.e., /package.service/method.
		FullMethod string
	}

	MethodDesc struct {
		MethodName string
		Handler    methodHandler
	}
	// Stream defines the common interface a client or server stream has to satisfy.
	Stream interface {
		Context() context.Context
		SendMsg(ctx context.Context, m interface{}) error
		RecvMsg(ctx context.Context, m interface{}) (context.Context, error)
		Close() error
	}

	ClientStream interface {
		Stream
		Header() (MD, error)
		Trailer() MD
		CloseSend() error
	}

	ServerStream interface {
		Stream
		SetHeader(MD) error
		SendHeader(MD) error
		SetTrailer(MD)
	}
	StreamHandler func(srv interface{}, stream ServerStream) error

	// StreamDesc represents a streaming RPC service's method specification.
	StreamDesc struct {
		StreamName string
		Handler    StreamHandler

		ServerStreams bool
		ClientStreams bool
	}
	ServiceDesc struct {
		ServiceName string
		HandlerType interface{}
		Methods     []MethodDesc
		Streams     []StreamDesc
		Metadata    interface{}
	}
	methodHandler          func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor UnaryServerInterceptor) (interface{}, error)
	UnaryHandler           func(ctx context.Context, req interface{}) (interface{}, error)
	UnaryServerInterceptor func(ctx context.Context, req interface{}, info *UnaryServerInfo, handler UnaryHandler) (resp interface{}, err error)
)
