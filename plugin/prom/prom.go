package prom

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"x.io/xrpc/pkg/codes"
)

const (
	Name = "prom"
)

func New() *promPlugin {
	p := &promPlugin{
		metrics: newDefaultMetrics(Server),
	}
	reg := prometheus.NewRegistry()
	reg.MustRegister(p.metrics)
	p.uri = "/metrics"
	p.port = 13140
	p.reg = reg
	return p
}

type promPlugin struct {
	metrics *DefaultMetrics
	uri     string
	port    uint
	reg     *prometheus.Registry
	s       *http.Server
}

func (p *promPlugin) Collect(cs ...prometheus.Collector) {
	if p.reg != nil {
		p.reg.MustRegister(cs...)
	}
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
	var gather prometheus.Gatherer = p.reg
	if gather == nil {
		gather = prometheus.DefaultGatherer
	}

	http.Handle(p.uri, promhttp.HandlerFor(gather, promhttp.HandlerOpts{}))
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
