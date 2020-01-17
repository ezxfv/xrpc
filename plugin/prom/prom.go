package prom

import (
	"context"
	"fmt"
	"net/http"

	"github.com/edenzhong7/xrpc/plugin"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	plugin.RegisterBuilder(&builder{})
}

const (
	Name = "prom"
)

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

type promMiddleware struct{}

func (m *promMiddleware) Name() string {
	return "prom"
}
func (m *promMiddleware) Handle(ctx context.Context, handler interface{}, args ...interface{}) (newHandler interface{}) {
	panic("implement me")
}

// StartServe 在指定地址上开启prometheus http，未提供Gatherer的情况下使用默认Gatherer
func StartServe(uri string, port uint, gatherer prometheus.Gatherer) {
	if gatherer == nil {
		gatherer = prometheus.DefaultGatherer
	}
	http.Handle(uri, promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{}))
	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
