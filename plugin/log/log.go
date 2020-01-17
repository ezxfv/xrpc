package log

import (
	"context"

	"github.com/edenzhong7/xrpc/pkg/log"
	"github.com/edenzhong7/xrpc/plugin"
)

func init() {
	plugin.RegisterPlugin(&logPlugin{})
}

type logPlugin struct {
}

func (p *logPlugin) Name() string {
	return "log"
}

func (p *logPlugin) Init(ctx context.Context) error {
	panic("implement me")
}

func (p *logPlugin) SetLogger(logger log.Logger) {
	panic("implement me")
}

func (p *logPlugin) OnConnect(ctx context.Context) error {
	panic("implement me")
}

func (p *logPlugin) OnSend(ctx context.Context) error {
	panic("implement me")
}

func (p *logPlugin) OnRecv(ctx context.Context) error {
	panic("implement me")
}

func (p *logPlugin) OnDisconnect(ctx context.Context) error {
	panic("implement me")
}

func (p *logPlugin) Destroy() {
	panic("implement me")
}
