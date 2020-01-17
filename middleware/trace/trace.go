package trace

import (
	"context"

	"github.com/edenzhong7/xrpc/middleware"
)

const (
	Name = "trace"
)

func init() {
	middleware.RegisterBuilder(&builder{})
}

type builder struct{}

func (b *builder) Name() string {
	return Name
}

func (b *builder) NewClientMiddleware() middleware.ClientMiddleware {
	panic("implement me")
}

func (b *builder) NewServerMiddleware() middleware.ServerMiddleware {
	panic("implement me")
}

type traceMiddleware struct {
}

func (m *traceMiddleware) Name() string {
	return "trace"
}

func (m *traceMiddleware) Handle(ctx context.Context, handler interface{}, args ...interface{}) (newHandler interface{}) {
	panic("implement me")
}
