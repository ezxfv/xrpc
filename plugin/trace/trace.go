package trace

import (
	"context"

	"google.golang.org/grpc"
)

const (
	Name = "trace"
)

func init() {
	registerJaeger()
}

func New() *tracePlugin {
	t := &tracePlugin{}
	return t
}

type tracePlugin struct {
}

func (t *tracePlugin) PreHandle(ctx context.Context, r interface{}, info *grpc.UnaryServerInfo) (context.Context, error) {
	println("trace pre")
	return ctx, nil
}

func (t *tracePlugin) PostHandle(ctx context.Context, req interface{}, resp interface{}, info *grpc.UnaryServerInfo, e error) (context.Context, error) {
	println("trace post")
	return ctx, nil
}
