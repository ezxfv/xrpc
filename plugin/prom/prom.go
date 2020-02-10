package prom

import (
	"errors"
	"fmt"
	"net/http"

	"context"

	"github.com/edenzhong7/xrpc/pkg/codes"
	"github.com/edenzhong7/xrpc/plugin"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

const (
	Name = "prom"
)

func New() plugin.Plugin {
	p := &promPlugin{
		metrics: newDefaultMetrics(Server),
	}
	reg := prometheus.NewRegistry()
	reg.MustRegister(p.metrics)
	go StartServe("/metrics", 13140, reg)
	return p
}

type promPlugin struct {
	metrics *DefaultMetrics
}

func (p *promPlugin) PreHandle(ctx context.Context, r interface{}, info *grpc.UnaryServerInfo) (context.Context, error) {
	reporter := newDefaultReporter(p.metrics, "unary", info.FullMethod)
	ctx = context.WithValue(ctx, "reporter", reporter)
	return ctx, nil
}

func (p *promPlugin) PostHandle(ctx context.Context, req interface{}, resp interface{}, info *grpc.UnaryServerInfo, e error) (context.Context, error) {
	r, ok := ctx.Value("reporter").(*defaultReporter)
	if !ok {
		return ctx, errors.New("prom plugin PostHandle get reporter from ctx failed")
	}
	r.Handled(codes.ErrorClass(e))
	return ctx, nil
}

// StartServe 在指定地址上开启prometheus http，未提供Gatherer的情况下使用默认Gatherer
func StartServe(uri string, port uint, gatherer prometheus.Gatherer) {
	if gatherer == nil {
		gatherer = prometheus.DefaultGatherer
	}
	http.Handle(uri, promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{}))
	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
