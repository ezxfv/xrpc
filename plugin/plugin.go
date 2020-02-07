package plugin

import (
	"context"
	"errors"
	"plugin"
	"sync"

	"github.com/edenzhong7/xrpc/pkg/net"

	"google.golang.org/grpc"
)

type Plugin interface{}

//PluginContainer represents a plugin container that defines all methods to manage plugins.
//And it also defines all extension points.
type Container interface {
	Add(plugin Plugin)
	Remove(plugin Plugin)

	DoRegisterService(sd *grpc.ServiceDesc, ss interface{}) error
	DoRegisterCustomService(sd *grpc.ServiceDesc, ss interface{}, metadata string) error
	DoRegisterFunction(serviceName, fname string, fn interface{}, metadata string) error

	DoConnect(net.Conn) (net.Conn, bool)
	DoDisconnect(net.Conn) bool

	DoOpenStream(ctx context.Context, conn net.Conn) (context.Context, error)
	DoCloseStream(ctx context.Context, conn net.Conn) (context.Context, error)

	DoHandle(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error)

	IOContainer
}

type IOContainer interface {
	DoPreReadRequest(ctx context.Context) error
	DoPostReadRequest(ctx context.Context, r interface{}, e error) error

	DoPreWriteResponse(ctx context.Context, req interface{}, resp interface{}) error
	DoPostWriteResponse(ctx context.Context, req interface{}, resp interface{}, e error) error
}

type (
	RegisterServicePlugin interface {
		RegisterService(sd *grpc.ServiceDesc, ss interface{}) error
	}

	RegisterCustomServicePlugin interface {
		RegisterCustomService(sd *grpc.ServiceDesc, ss interface{}, metadata string) error
	}

	RegisterFunctionPlugin interface {
		RegisterFunction(serviceName, fname string, fn interface{}, metadata string) error
	}

	ConnectPlugin interface {
		Connect(conn net.Conn) (net.Conn, bool)
	}

	DisconnectPlugin interface {
		Disconnect(conn net.Conn) bool
	}

	OpenStreamPlugin interface {
		OpenStream(ctx context.Context, conn net.Conn) (context.Context, error)
	}

	CloseStreamPlugin interface {
		CloseStream(ctx context.Context, conn net.Conn) (context.Context, error)
	}

	PreReadRequestPlugin interface {
		PreReadRequest(ctx context.Context) error
	}

	PostReadRequestPlugin interface {
		PostReadRequest(ctx context.Context, r interface{}, e error) error
	}

	PreHandlePlugin interface {
		PreHandle(ctx context.Context, r interface{}, info *grpc.UnaryServerInfo) (context.Context, error)
	}

	PostHandlePlugin interface {
		PostHandle(ctx context.Context, req interface{}, resp interface{}, info *grpc.UnaryServerInfo, e error) (context.Context, error)
	}

	PreWriteResponsePlugin interface {
		PreWriteResponse(ctx context.Context, req interface{}, resp interface{}) error
	}

	PostWriteResponsePlugin interface {
		PostWriteResponse(ctx context.Context, req interface{}, resp interface{}, e error) error
	}

	AlmightyPlugin interface {
		RegisterServicePlugin
		RegisterCustomServicePlugin
		RegisterFunctionPlugin
		ConnectPlugin
		DisconnectPlugin
		OpenStreamPlugin
		CloseStreamPlugin
		PreReadRequestPlugin
		PostReadRequestPlugin
		PreHandlePlugin
		PostHandlePlugin
		PreWriteResponsePlugin
		PostWriteResponsePlugin
	}
)

func NewPluginContainer() Container {
	pc := &pluginContainer{
		plugins: map[Plugin]bool{},
		rsp:     map[RegisterServicePlugin]bool{},
		rcsp:    map[RegisterCustomServicePlugin]bool{},
		rfp:     map[RegisterFunctionPlugin]bool{},
		cp:      map[ConnectPlugin]bool{},
		dp:      map[DisconnectPlugin]bool{},
		osp:     map[OpenStreamPlugin]bool{},
		csp:     map[CloseStreamPlugin]bool{},
		prrp:    map[PreReadRequestPlugin]bool{},
		porrp:   map[PostReadRequestPlugin]bool{},
		php:     map[PreHandlePlugin]bool{},
		pohp:    map[PostHandlePlugin]bool{},
		pwrp:    map[PreWriteResponsePlugin]bool{},
		powrp:   map[PostWriteResponsePlugin]bool{},

		mu: &sync.Mutex{},
	}
	return pc
}

//
//var (
//	nilPlugins = []reflect.Type{
//		reflect.TypeOf(Plugin(nil)),
//		reflect.TypeOf(RegisterServicePlugin(nil)),
//		reflect.TypeOf(RegisterCustomServicePlugin(nil)),
//		reflect.TypeOf(RegisterFunctionPlugin(nil)),
//		reflect.TypeOf(ConnectPlugin(nil)),
//		reflect.TypeOf(DisconnectPlugin(nil)),
//		reflect.TypeOf(OpenStreamPlugin(nil)),
//		reflect.TypeOf(CloseStreamPlugin(nil)),
//		reflect.TypeOf(PreReadRequestPlugin(nil)),
//		reflect.TypeOf(PostReadRequestPlugin(nil)),
//		reflect.TypeOf(PreHandlePlugin(nil)),
//		reflect.TypeOf(PostHandlePlugin(nil)),
//		reflect.TypeOf(PreWriteResponsePlugin(nil)),
//		reflect.TypeOf(PostWriteResponsePlugin(nil)),
//	}
//)

// pluginContainer implements PluginContainer interface.
type pluginContainer struct {
	plugins map[Plugin]bool
	rsp     map[RegisterServicePlugin]bool
	rcsp    map[RegisterCustomServicePlugin]bool
	rfp     map[RegisterFunctionPlugin]bool
	cp      map[ConnectPlugin]bool
	dp      map[DisconnectPlugin]bool
	osp     map[OpenStreamPlugin]bool
	csp     map[CloseStreamPlugin]bool
	prrp    map[PreReadRequestPlugin]bool
	porrp   map[PostReadRequestPlugin]bool
	php     map[PreHandlePlugin]bool
	pohp    map[PostHandlePlugin]bool
	pwrp    map[PreWriteResponsePlugin]bool
	powrp   map[PostWriteResponsePlugin]bool

	mu *sync.Mutex
}

// Add adds a plugin.
func (pc *pluginContainer) Add(plugin Plugin) {
	if plugin == nil {
		return
	}

	pc.plugins[plugin] = true

	if p, ok := plugin.(RegisterServicePlugin); ok {
		pc.rsp[p] = true
	}
	if p, ok := plugin.(RegisterCustomServicePlugin); ok {
		pc.rcsp[p] = true
	}
	if p, ok := plugin.(RegisterFunctionPlugin); ok {
		pc.rfp[p] = true
	}
	if p, ok := plugin.(ConnectPlugin); ok {
		pc.cp[p] = true
	}
	if p, ok := plugin.(DisconnectPlugin); ok {
		pc.dp[p] = true
	}
	if p, ok := plugin.(OpenStreamPlugin); ok {
		pc.osp[p] = true
	}
	if p, ok := plugin.(CloseStreamPlugin); ok {
		pc.csp[p] = true
	}
	if p, ok := plugin.(PreReadRequestPlugin); ok {
		pc.prrp[p] = true
	}
	if p, ok := plugin.(PostReadRequestPlugin); ok {
		pc.porrp[p] = true
	}
	if p, ok := plugin.(PreHandlePlugin); ok {
		pc.php[p] = true
	}
	if p, ok := plugin.(PostHandlePlugin); ok {
		pc.pohp[p] = true
	}
	if p, ok := plugin.(PreWriteResponsePlugin); ok {
		pc.pwrp[p] = true
	}
	if p, ok := plugin.(PostWriteResponsePlugin); ok {
		pc.powrp[p] = true
	}
}

// Remove removes a plugin by it's name.
func (pc *pluginContainer) Remove(plugin Plugin) {
	if plugin == nil {
		return
	}
	delete(pc.plugins, plugin)

	if p, ok := plugin.(RegisterServicePlugin); ok {
		delete(pc.rsp, p)
	}
	if p, ok := plugin.(RegisterCustomServicePlugin); ok {
		delete(pc.rcsp, p)
	}
	if p, ok := plugin.(RegisterFunctionPlugin); ok {
		delete(pc.rfp, p)
	}
	if p, ok := plugin.(ConnectPlugin); ok {
		delete(pc.cp, p)
	}
	if p, ok := plugin.(DisconnectPlugin); ok {
		delete(pc.dp, p)
	}
	if p, ok := plugin.(OpenStreamPlugin); ok {
		delete(pc.osp, p)
	}
	if p, ok := plugin.(CloseStreamPlugin); ok {
		delete(pc.csp, p)
	}
	if p, ok := plugin.(PreReadRequestPlugin); ok {
		delete(pc.prrp, p)
	}
	if p, ok := plugin.(PostReadRequestPlugin); ok {
		delete(pc.porrp, p)
	}
	if p, ok := plugin.(PreHandlePlugin); ok {
		delete(pc.php, p)
	}
	if p, ok := plugin.(PostHandlePlugin); ok {
		delete(pc.pohp, p)
	}
	if p, ok := plugin.(PreWriteResponsePlugin); ok {
		delete(pc.pwrp, p)
	}
	if p, ok := plugin.(PostWriteResponsePlugin); ok {
		delete(pc.powrp, p)
	}
}

func (pc *pluginContainer) DoRegisterService(sd *grpc.ServiceDesc, ss interface{}) error {
	var err error
	for p := range pc.rsp {
		err = p.RegisterService(sd, ss)
		if err != nil {
			break
		}
	}
	return err
}

func (pc *pluginContainer) DoRegisterCustomService(sd *grpc.ServiceDesc, ss interface{}, metadata string) error {
	var err error
	for p := range pc.rcsp {
		err = p.RegisterCustomService(sd, ss, metadata)
		if err != nil {
			break
		}
	}
	return err
}

func (pc *pluginContainer) DoRegisterFunction(serviceName, fname string, fn interface{}, metadata string) error {
	var err error
	for p := range pc.rfp {
		err = p.RegisterFunction(serviceName, fname, fn, metadata)
		if err != nil {
			break
		}
	}
	return err
}

func (pc *pluginContainer) DoConnect(conn net.Conn) (net.Conn, bool) {
	var ok bool
	for p := range pc.cp {
		conn, ok = p.Connect(conn)
		if !ok {
			break
		}
	}
	return conn, ok
}

func (pc *pluginContainer) DoDisconnect(conn net.Conn) bool {
	var ok bool
	for p := range pc.dp {
		ok = p.Disconnect(conn)
		if !ok {
			break
		}
	}
	return ok
}

func (pc *pluginContainer) DoOpenStream(ctx context.Context, conn net.Conn) (context.Context, error) {
	var err error
	for p := range pc.osp {
		ctx, err = p.OpenStream(ctx, conn)
		if err != nil {
			break
		}
	}
	return ctx, err
}

func (pc *pluginContainer) DoCloseStream(ctx context.Context, conn net.Conn) (context.Context, error) {
	var err error
	for p := range pc.csp {
		ctx, err = p.CloseStream(ctx, conn)
		if err != nil {
			break
		}
	}
	return ctx, err
}

func (pc *pluginContainer) DoPreReadRequest(ctx context.Context) error {
	var err error
	for p := range pc.prrp {
		err = p.PreReadRequest(ctx)
		if err != nil {
			break
		}
	}
	return err
}

func (pc *pluginContainer) DoPostReadRequest(ctx context.Context, r interface{}, e error) error {
	var err error
	for p := range pc.porrp {
		err = p.PostReadRequest(ctx, r, e)
		if err != nil {
			break
		}
	}
	return err
}

func (pc *pluginContainer) DoHandle(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	for p := range pc.php {
		ctx, err = p.PreHandle(ctx, req, info)
		if err != nil {
			break
		}
	}
	resp, err = handler(ctx, req)
	e := err
	for p := range pc.pohp {
		ctx, err = p.PostHandle(ctx, req, resp, info, e)
		if err != nil {
			break
		}
	}
	return resp, err
}

func (pc *pluginContainer) DoPreWriteResponse(ctx context.Context, req interface{}, resp interface{}) error {
	var err error
	for p := range pc.pwrp {
		err = p.PreWriteResponse(ctx, req, resp)
		if err != nil {
			break
		}
	}
	return err
}

func (pc *pluginContainer) DoPostWriteResponse(ctx context.Context, req interface{}, resp interface{}, e error) error {
	var err error
	for p := range pc.powrp {
		err = p.PostWriteResponse(ctx, req, resp, e)
		if err != nil {
			break
		}
	}
	return err
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
	return
}
