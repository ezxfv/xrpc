package plugin

import (
	"context"
	"errors"
	"net/http"
	"plugin"
	"strconv"
	"sync"

	"x.io/xrpc/api"
	echo "x.io/xrpc/pkg/echo"
	"x.io/xrpc/pkg/net"
	"x.io/xrpc/types"
)

type Plugin interface{}

//PluginContainer represents a plugin container that defines all methods to manage plugins.
//And it also defines all extension points.
type Container interface {
	Start() error
	Stop() error

	Add(plugin Plugin)
	Remove(plugin Plugin)

	DoRegisterService(sd *types.ServiceDesc, ss interface{}) error
	DoRegisterCustomService(sd *types.ServiceDesc, ss interface{}, metadata string) error
	DoRegisterFunction(serviceName, fname string, fn interface{}, metadata string) error

	DoConnect(net.Conn) (net.Conn, bool)
	DoDisconnect(net.Conn) bool

	DoOpenStream(ctx context.Context, conn net.Conn) (context.Context, error)
	DoCloseStream(ctx context.Context, conn net.Conn) (context.Context, error)

	DoIntercept(ctx context.Context, req interface{}, info *types.UnaryServerInfo, handler types.UnaryHandler) (resp interface{}, err error)

	IOContainer
}

type IOContainer interface {
	DoPreReadRequest(ctx context.Context, data []byte) ([]byte, error)
	DoPostReadRequest(ctx context.Context, r interface{}, e error) error

	DoPreWriteResponse(ctx context.Context, data []byte) ([]byte, error)
	DoPostWriteResponse(ctx context.Context, req interface{}, resp interface{}, e error) error
}

type (
	StartPlugin interface {
		Start() error
	}

	StopPlugin interface {
		Stop() error
	}

	RegisterServicePlugin interface {
		RegisterService(sd *types.ServiceDesc, ss interface{}) error
	}

	RegisterCustomServicePlugin interface {
		RegisterCustomService(sd *types.ServiceDesc, ss interface{}, metadata string) error
	}

	RegisterFunctionPlugin interface {
		RegisterFunction(serviceName, fname string, fn interface{}, metadata string) error
	}

	UnregisterServicePlugin interface {
		Unregister(serviceName string) error
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
		PreReadRequest(ctx context.Context, data []byte) ([]byte, error)
	}

	PostReadRequestPlugin interface {
		PostReadRequest(ctx context.Context, r interface{}, e error) error
	}

	Interceptor func(ctx context.Context, req interface{}, info *types.UnaryServerInfo, handler types.UnaryHandler) (resp interface{}, err error)

	InterceptPlugin interface {
		Intercept(ctx context.Context, req interface{}, info *types.UnaryServerInfo, handler types.UnaryHandler) (resp interface{}, err error)
	}

	PreWriteResponsePlugin interface {
		PreWriteResponse(ctx context.Context, data []byte) ([]byte, error)
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
		InterceptPlugin
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
		inp:     map[InterceptPlugin]bool{},
		pwrp:    map[PreWriteResponsePlugin]bool{},
		powrp:   map[PostWriteResponsePlugin]bool{},

		mu: &sync.Mutex{},
	}
	api.Register(pc)
	return pc
}

func NewIOPluginContainer() IOContainer {
	return NewPluginContainer()
}

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
	inp     map[InterceptPlugin]bool
	pwrp    map[PreWriteResponsePlugin]bool
	powrp   map[PostWriteResponsePlugin]bool

	mu *sync.Mutex
}

func (pc *pluginContainer) Start() (err error) {
	for plugin := range pc.plugins {
		if pp, ok := plugin.(StartPlugin); ok {
			err = pp.Start()
			if err != nil {
				return
			}
		}
	}
	return
}

func (pc *pluginContainer) Stop() (err error) {
	for plugin := range pc.plugins {
		if pp, ok := plugin.(StopPlugin); ok {
			err = pp.Stop()
			if err != nil {
				return
			}
		}
	}
	return
}

// Add adds a plugin.
func (pc *pluginContainer) Add(plugin Plugin) {
	if plugin == nil {
		return
	}

	pc.plugins[plugin] = true

	if p, ok := plugin.(api.APIer); ok {
		api.Register(p)
	}

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
	if p, ok := plugin.(InterceptPlugin); ok {
		pc.inp[p] = true
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
	if p, ok := plugin.(InterceptPlugin); ok {
		delete(pc.inp, p)
	}
	if p, ok := plugin.(PreWriteResponsePlugin); ok {
		delete(pc.pwrp, p)
	}
	if p, ok := plugin.(PostWriteResponsePlugin); ok {
		delete(pc.powrp, p)
	}
}

func (pc *pluginContainer) DoRegisterService(sd *types.ServiceDesc, ss interface{}) error {
	var err error
	for p := range pc.rsp {
		err = p.RegisterService(sd, ss)
		if err != nil {
			break
		}
	}
	return err
}

func (pc *pluginContainer) DoRegisterCustomService(sd *types.ServiceDesc, ss interface{}, metadata string) error {
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

func (pc *pluginContainer) DoPreReadRequest(ctx context.Context, data []byte) ([]byte, error) {
	var err error
	for p := range pc.prrp {
		data, err = p.PreReadRequest(ctx, data)
		if err != nil {
			break
		}
	}
	return data, err
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

func (pc *pluginContainer) DoIntercept(ctx context.Context, req interface{}, info *types.UnaryServerInfo, handler types.UnaryHandler) (resp interface{}, err error) {
	if len(pc.inp) == 0 {
		return
	}
	chain := func(in Interceptor, handler types.UnaryHandler) types.UnaryHandler {
		return func(ctx context.Context, req interface{}) (resp interface{}, err error) {
			return in(ctx, req, info, handler)
		}
	}
	chainHandler := handler
	for in := range pc.inp {
		chainHandler = chain(in.Intercept, chainHandler)
	}
	return chainHandler(ctx, req)
}

func (pc *pluginContainer) DoPreWriteResponse(ctx context.Context, data []byte) ([]byte, error) {
	var err error
	for p := range pc.pwrp {
		data, err = p.PreWriteResponse(ctx, data)
		if err != nil {
			break
		}
	}
	return data, err
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

func (pc *pluginContainer) RegisterAPI(e *echo.Echo) {
	g := e.Group("/plugincontainer")
	g.GET("", func(c echo.Context) error {
		return c.String(http.StatusOK, "plugin contains api is working")
	})
	g.GET("/count", func(c echo.Context) error {
		return c.String(http.StatusOK, strconv.Itoa(len(pc.plugins)))
	})
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
