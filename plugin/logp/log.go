package logp

import (
	"context"
	"os"

	"x.io/xrpc/types"

	"x.io/xrpc/pkg/log"
	"x.io/xrpc/pkg/net"
)

func New() *logPlugin {
	return &logPlugin{
		l: log.NewSimpleDefaultLogger(os.Stdout, log.DEBUG, "log_plugin", true),
	}
}

type logPlugin struct {
	l log.Logger
}

func (p *logPlugin) Logger() log.Logger {
	return p.l
}

func (p *logPlugin) Start() error {
	p.l.Debug("starting log plugin")
	return nil
}

func (p *logPlugin) Stop() error {
	p.l.Debug("stopping log plugin")
	return nil
}

func (p *logPlugin) RegisterService(sd *types.ServiceDesc, ss interface{}) error {
	var methods []string
	for _, m := range sd.Methods {
		methods = append(methods, m.MethodName)
	}
	p.l.Debugf("Register Service %s: %#v\n", sd.ServiceName, methods)
	return nil
}

func (p *logPlugin) RegisterCustomService(sd *types.ServiceDesc, ss interface{}, metadata string) error {
	var methods []string
	for _, m := range sd.Methods {
		methods = append(methods, m.MethodName)
	}
	p.l.Debugf("Register Custom Service %s: %#v, metadata:%s\n", sd.ServiceName, methods, metadata)
	return nil
}

func (p *logPlugin) RegisterFunction(serviceName, fname string, fn interface{}, metadata string) error {
	p.l.Debugf("Register Function [%s] -> Service [%s], metadata:%s\n", fname, serviceName, metadata)
	return nil
}

func (p *logPlugin) Connect(conn net.Conn) (net.Conn, bool) {
	p.l.Debugf("Accept connect from %s:%s\n", conn.RemoteAddr().Network(), conn.RemoteAddr().String())
	return conn, true
}

func (p *logPlugin) Disconnect(conn net.Conn) bool {
	p.l.Debugf("Disconnect to %s:%s\n", conn.RemoteAddr().Network(), conn.RemoteAddr().String())
	return true
}

func (p logPlugin) OpenStream(ctx context.Context, conn net.Conn) (context.Context, error) {
	p.l.Debugf("Open stream from %s:%s\n", conn.RemoteAddr().Network(), conn.RemoteAddr().String())
	return ctx, nil
}

func (p logPlugin) CloseStream(ctx context.Context, conn net.Conn) (context.Context, error) {
	p.l.Debugf("Close stream from %s:%s\n", conn.RemoteAddr().Network(), conn.RemoteAddr().String())
	return ctx, nil
}

func (p *logPlugin) PreReadRequest(ctx context.Context, data []byte) ([]byte, error) {
	p.l.Debugf("PreReadRequest\n")
	return data, nil
}

func (p *logPlugin) PostReadRequest(ctx context.Context, r interface{}, e error) error {
	p.l.Debugf("PostReadRequest, err:%#v\n", e)
	return nil
}

func (p *logPlugin) Intercept(ctx context.Context, req interface{}, info *types.UnaryServerInfo, handler types.UnaryHandler) (resp interface{}, err error) {
	p.l.Debugf("Intercept\n")
	return handler(ctx, req)
}

func (p *logPlugin) PreWriteResponse(ctx context.Context, data []byte) ([]byte, error) {
	p.l.Debugf("PreWriteResponse\n")
	return data, nil
}

func (p *logPlugin) PostWriteResponse(ctx context.Context, req interface{}, resp interface{}, e error) error {
	p.l.Debugf("PostWriteResponse\n")
	return nil
}
