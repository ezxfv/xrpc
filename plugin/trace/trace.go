package trace

import (
	"context"

	"github.com/edenzhong7/xrpc/plugin"
)

const (
	Name = "trace"
)

func init() {
	plugin.RegisterBuilder(&builder{})
}

type builder struct{}

func (b *builder) Name() string {
	return Name
}

func (b *builder) NewClientMiddleware() plugin.ClientMiddleware {
	panic("implement me")
}

func (b *builder) NewServerMiddleware() plugin.ServerMiddleware {
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
