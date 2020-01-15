package plugin

import (
	"context"
	"errors"
	"plugin"

	"github.com/edenzhong7/xrpc/log"
)

type Plugin interface {
	Name() string
	Init(ctx context.Context) error
	SetLogger(logger log.Logger)

	OnConnect(ctx context.Context) error
	OnSend(ctx context.Context) error
	OnRecv(ctx context.Context) error
	OnDisconnect(ctx context.Context) error

	Destroy()
}

func LoadPluginDLL(ctx context.Context, libPath string) (p Plugin, err error) {
	dll, err := plugin.Open(libPath)
	if err != nil {
		return
	}
	builderName := "NewPlugin"
	builderSymbol, err := dll.Lookup(builderName)
	if err != nil {
		return
	}
	builder, ok := builderSymbol.(func(ctx context.Context) Plugin)
	if !ok {
		return nil, errors.New("unexpected builder " + builderName)
	}
	p = builder(ctx)
	if !ok {
		return nil, errors.New("already loaded plugin: " + p.Name())
	}
	log.GLogger().Debug("loaded Plugin " + p.Name() + " from dll...")
	return
}
