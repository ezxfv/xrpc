package plugin

import (
	"context"
	"errors"
	"plugin"

	"github.com/edenzhong7/xrpc/pkg/common"
	"github.com/edenzhong7/xrpc/pkg/log"
)

var (
	allPlugins map[string]Plugin
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
	log.Debug("loaded Plugin " + p.Name() + " from dll...")
	RegisterPlugin(p)
	return
}

func init() {
	allPlugins = map[string]Plugin{}
}

func RegisterPlugin(p Plugin) {
	if _, ok := allPlugins[p.Name()]; !ok {
		allPlugins[p.Name()] = p
	}
}

func PickPlugin(name string) Plugin {
	return allPlugins[name]
}

func PickSomePlugins(names []string) []Plugin {
	var plugins []Plugin
	for _, name := range names {
		if p, ok := allPlugins[name]; ok {
			plugins = append(plugins, p)
		}
	}
	return plugins
}

// TODO 应用插件
func OnConnect(ctx context.Context, plugins []Plugin, handler interface{}, args ...interface{}) (result []interface{}, err error) {
	var newHandler = handler
	for _, m := range plugins {
		newHandler = m.OnSend(ctx)
	}
	result, err = common.Call(newHandler, newHandler)
	return
}

func OnDisconnect(ctx context.Context, plugins []Plugin, handler interface{}, args ...interface{}) (result []interface{}, err error) {
	var newHandler = handler
	for _, m := range plugins {
		newHandler = m.OnSend(ctx)
	}
	result, err = common.Call(newHandler, newHandler)
	return
}

func OnSend(ctx context.Context, plugins []Plugin, handler interface{}, args ...interface{}) (result []interface{}, err error) {
	var newHandler = handler
	for _, m := range plugins {
		newHandler = m.OnSend(ctx)
	}
	result, err = common.Call(newHandler, newHandler)
	return
}

func OnRecv(ctx context.Context, plugins []Plugin, handler interface{}, args ...interface{}) (result []interface{}, err error) {
	var newHandler = handler
	for _, m := range plugins {
		newHandler = m.OnSend(ctx)
	}
	result, err = common.Call(newHandler, newHandler)
	return
}
