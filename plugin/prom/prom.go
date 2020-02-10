package prom

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

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
	p.uri = "/metrics"
	p.port = 13140
	p.gatherer = reg
	return p
}

type promPlugin struct {
	metrics  *DefaultMetrics
	uri      string
	port     uint
	gatherer prometheus.Gatherer
	s        *http.Server
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

// Start 在指定地址上开启prometheus http，未提供Gatherer的情况下使用默认Gatherer
func (p *promPlugin) Start() error {
	if p.gatherer == nil {
		p.gatherer = prometheus.DefaultGatherer
	}

	http.Handle(p.uri, promhttp.HandlerFor(p.gatherer, promhttp.HandlerOpts{}))
	addr := fmt.Sprintf(":%d", p.port)
	server := &http.Server{
		Addr:    addr,
		Handler: http.DefaultServeMux,
	}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	go server.Serve(lis)
	p.s = server
	return nil
}

func (p *promPlugin) Stop() error {
	return p.s.Close()
}
