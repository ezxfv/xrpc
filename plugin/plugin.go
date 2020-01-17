package plugin

import (
	"context"
	"errors"
	"plugin"

	"github.com/edenzhong7/xrpc/pkg/common"
	"github.com/edenzhong7/xrpc/pkg/log"
	"github.com/edenzhong7/xrpc/pkg/net"

	"github.com/gogo/protobuf/proto"
)

var (
	allPlugins map[string]Plugin
)

type Message = proto.Message

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

//PluginContainer represents a plugin container that defines all methods to manage plugins.
//And it also defines all extension points.
type Container interface {
	Add(plugin Plugin)
	Remove(plugin Plugin)
	All() []Plugin

	DoRegister(name string, rcvr interface{}, metadata string) error
	DoRegisterFunction(serviceName, fname string, fn interface{}, metadata string) error
	DoUnregister(name string) error

	DoPostConnAccept(net.Conn) (net.Conn, bool)
	DoPostConnClose(net.Conn) bool

	DoPreReadRequest(ctx context.Context) error
	DoPostReadRequest(ctx context.Context, r Message, e error) error

	DoPreHandleRequest(ctx context.Context, req Message) error

	DoPreWriteResponse(context.Context, Message, Message) error
	DoPostWriteResponse(context.Context, Message, Message, error) error

	DoPreWriteRequest(ctx context.Context) error
	DoPostWriteRequest(ctx context.Context, r Message, e error) error
}

type (
	//// RegisterPlugin is .
	RegisterPlugin interface {
		Register(name string, rcvr interface{}, metadata string) error
		Unregister(name string) error
	}

	// RegisterFunctionPlugin is .
	RegisterFunctionPlugin interface {
		RegisterFunction(serviceName, fname string, fn interface{}, metadata string) error
	}

	// PostConnAcceptPlugin represents connection accept plugin.
	// if returns false, it means subsequent IPostConnAcceptPlugins should not contiune to handle this conn
	// and this conn has been closed.
	PostConnAcceptPlugin interface {
		HandleConnAccept(net.Conn) (net.Conn, bool)
	}

	// PostConnClosePlugin represents client connection close plugin.
	PostConnClosePlugin interface {
		HandleConnClose(net.Conn) bool
	}

	//PreReadRequestPlugin represents .
	PreReadRequestPlugin interface {
		PreReadRequest(ctx context.Context) error
	}

	//PostReadRequestPlugin represents .
	PostReadRequestPlugin interface {
		PostReadRequest(ctx context.Context, r Message, e error) error
	}

	//PreHandleRequestPlugin represents .
	PreHandleRequestPlugin interface {
		PreHandleRequest(ctx context.Context, r Message) error
	}

	//PreWriteResponsePlugin represents .
	PreWriteResponsePlugin interface {
		PreWriteResponse(context.Context, Message, Message) error
	}

	//PostWriteResponsePlugin represents .
	PostWriteResponsePlugin interface {
		PostWriteResponse(context.Context, Message, Message, error) error
	}

	//PreWriteRequestPlugin represents .
	PreWriteRequestPlugin interface {
		PreWriteRequest(ctx context.Context) error
	}

	//PostWriteRequestPlugin represents .
	PostWriteRequestPlugin interface {
		PostWriteRequest(ctx context.Context, r Message, e error) error
	}
)

// pluginContainer implements PluginContainer interface.
type pluginContainer struct {
	plugins []Plugin
}

// Add adds a plugin.
func (p *pluginContainer) Add(plugin Plugin) {
	p.plugins = append(p.plugins, plugin)
}

// Remove removes a plugin by it's name.
func (p *pluginContainer) Remove(plugin Plugin) {
	if p.plugins == nil {
		return
	}

	plugins := make([]Plugin, 0, len(p.plugins))
	for _, p := range p.plugins {
		if p != plugin {
			plugins = append(plugins, p)
		}
	}

	p.plugins = plugins
}

func (p *pluginContainer) All() []Plugin {
	return p.plugins
}

// DoRegister invokes DoRegister plugin.
func (p *pluginContainer) DoRegister(name string, rcvr interface{}, metadata string) error {
	var es []error
	for _, rp := range p.plugins {
		if p, ok := rp.(RegisterPlugin); ok {
			err := p.Register(name, rcvr, metadata)
			if err != nil {
				es = append(es, err)
			}
		}
	}

	if len(es) > 0 {
		return es[0]
	}
	return nil
}

// DoRegisterFunction invokes DoRegisterFunction plugin.
func (p *pluginContainer) DoRegisterFunction(serviceName, fname string, fn interface{}, metadata string) error {
	var es []error
	for _, rp := range p.plugins {
		if p, ok := rp.(RegisterFunctionPlugin); ok {
			err := p.RegisterFunction(serviceName, fname, fn, metadata)
			if err != nil {
				es = append(es, err)
			}
		}
	}

	if len(es) > 0 {
		return es[0]
	}
	return nil
}

// DoUnregister invokes RegisterXPlugin.
func (p *pluginContainer) DoUnregister(name string) error {
	var es []error
	for _, rp := range p.plugins {
		if p, ok := rp.(RegisterPlugin); ok {
			err := p.Unregister(name)
			if err != nil {
				es = append(es, err)
			}
		}
	}

	if len(es) > 0 {
		return es[0]
	}
	return nil
}

//DoPostConnAccept handles accepted conn
func (p *pluginContainer) DoPostConnAccept(conn net.Conn) (net.Conn, bool) {
	var flag bool
	for i := range p.plugins {
		if p, ok := p.plugins[i].(PostConnAcceptPlugin); ok {
			conn, flag = p.HandleConnAccept(conn)
			if !flag { //interrupt
				conn.Close()
				return conn, false
			}
		}
	}
	return conn, true
}

//DoPostConnClose handles closed conn
func (p *pluginContainer) DoPostConnClose(conn net.Conn) bool {
	var flag bool
	for i := range p.plugins {
		if p, ok := p.plugins[i].(PostConnClosePlugin); ok {
			flag = p.HandleConnClose(conn)
			if !flag {
				return false
			}
		}
	}
	return true
}

// DoPreReadRequest invokes PreReadRequest plugin.
func (p *pluginContainer) DoPreReadRequest(ctx context.Context) error {
	for i := range p.plugins {
		if p, ok := p.plugins[i].(PreReadRequestPlugin); ok {
			err := p.PreReadRequest(ctx)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// DoPostReadRequest invokes PostReadRequest plugin.
func (p *pluginContainer) DoPostReadRequest(ctx context.Context, r Message, e error) error {
	for i := range p.plugins {
		if p, ok := p.plugins[i].(PostReadRequestPlugin); ok {
			err := p.PostReadRequest(ctx, r, e)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// DoPreHandleRequest invokes PreHandleRequest plugin.
func (p *pluginContainer) DoPreHandleRequest(ctx context.Context, r Message) error {
	for i := range p.plugins {
		if p, ok := p.plugins[i].(PreHandleRequestPlugin); ok {
			err := p.PreHandleRequest(ctx, r)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// DoPreWriteResponse invokes PreWriteResponse plugin.
func (p *pluginContainer) DoPreWriteResponse(ctx context.Context, req Message, res Message) error {
	for i := range p.plugins {
		if p, ok := p.plugins[i].(PreWriteResponsePlugin); ok {
			err := p.PreWriteResponse(ctx, req, res)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// DoPostWriteResponse invokes PostWriteResponse plugin.
func (p *pluginContainer) DoPostWriteResponse(ctx context.Context, req Message, resp Message, e error) error {
	for i := range p.plugins {
		if p, ok := p.plugins[i].(PostWriteResponsePlugin); ok {
			err := p.PostWriteResponse(ctx, req, resp, e)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// DoPreWriteRequest invokes PreWriteRequest plugin.
func (p *pluginContainer) DoPreWriteRequest(ctx context.Context) error {
	for i := range p.plugins {
		if p, ok := p.plugins[i].(PreWriteRequestPlugin); ok {
			err := p.PreWriteRequest(ctx)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// DoPostWriteRequest invokes PostWriteRequest plugin.
func (p *pluginContainer) DoPostWriteRequest(ctx context.Context, r Message, e error) error {
	for i := range p.plugins {
		if p, ok := p.plugins[i].(PostWriteRequestPlugin); ok {
			err := p.PostWriteRequest(ctx, r, e)
			if err != nil {
				return err
			}
		}
	}

	return nil
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
	RegisterXPlugin(p)
	return
}

func init() {
	allPlugins = map[string]Plugin{}
}

func RegisterXPlugin(p Plugin) {
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
