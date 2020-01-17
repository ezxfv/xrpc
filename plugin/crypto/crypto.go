package crypto

import (
	"context"

	"github.com/edenzhong7/xrpc/pkg/log"
	"github.com/edenzhong7/xrpc/plugin"
)

func init() {
	plugin.RegisterPlugin(&cryptoPlugin{})
}

type cryptoPlugin struct {
}

func (p *cryptoPlugin) Name() string {
	return "crypto"
}

func (p *cryptoPlugin) Init(ctx context.Context) error {
	panic("implement me")
}

func (p *cryptoPlugin) SetLogger(logger log.Logger) {
	panic("implement me")
}

func (p *cryptoPlugin) OnConnect(ctx context.Context) error {
	panic("implement me")
}

func (p *cryptoPlugin) OnSend(ctx context.Context) error {
	panic("implement me")
}

func (p *cryptoPlugin) OnRecv(ctx context.Context) error {
	panic("implement me")
}

func (p *cryptoPlugin) OnDisconnect(ctx context.Context) error {
	panic("implement me")
}

func (p *cryptoPlugin) Destroy() {
	panic("implement me")
}
