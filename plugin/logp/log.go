package logp

import (
	"context"
	"net/http"
	"os"

	"x.io/xrpc/pkg/log"
	"x.io/xrpc/pkg/net"
	"x.io/xrpc/plugin"

	echo "github.com/labstack/echo/v4"
	"google.golang.org/grpc"
)

func New() plugin.Plugin {
	return &logPlugin{
		l: log.NewSimpleDefaultLogger(os.Stdout, log.DEBUG, "log_plugin", true),
	}
}

type logPlugin struct {
	l log.Logger
}

func (p *logPlugin) Start() error {
	p.l.Debug("starting log plugin")
	return nil
}

func (p *logPlugin) Stop() error {
	p.l.Debug("stopping log plugin")
	return nil
}

func (p *logPlugin) RegisterAPI(e *echo.Echo) {
	g := e.Group("log")
	g.GET("", func(c echo.Context) error {
		return c.String(http.StatusOK, "log plugin api")
	})
}

func (p *logPlugin) RegisterService(sd *grpc.ServiceDesc, ss interface{}) error {
	var methods []string
	for _, m := range sd.Methods {
		methods = append(methods, m.MethodName)
	}
	p.l.Debugf("Register Service %s: %#v\n", sd.ServiceName, methods)
	return nil
}

func (p *logPlugin) RegisterCustomService(sd *grpc.ServiceDesc, ss interface{}, metadata string) error {
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

func (p *logPlugin) PreHandle(ctx context.Context, r interface{}, info *grpc.UnaryServerInfo) (context.Context, error) {
	p.l.Debugf("PreHandle\n")
	return ctx, nil
}

func (p *logPlugin) PostHandle(ctx context.Context, req interface{}, resp interface{}, info *grpc.UnaryServerInfo, e error) (context.Context, error) {
	p.l.Debugf("PostHandle\n")
	return ctx, nil
}

func (p *logPlugin) PreWriteResponse(ctx context.Context, data []byte) ([]byte, error) {
	p.l.Debugf("PreWriteResponse\n")
	return data, nil
}

func (p *logPlugin) PostWriteResponse(ctx context.Context, req interface{}, resp interface{}, e error) error {
	p.l.Debugf("PostWriteResponse\n")
	return nil
}
