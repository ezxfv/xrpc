package echo

import (
	"fmt"
	"io"
	"net"
	"net/http"

	"x.io/xrpc/pkg/log"
)

var (
	SupportedMethods = []HttpMethod{GET, POST, DELETE, PUT, CONNECT, HEAD, OPTIONS, PATCH, TRACE}
)

type (
	URLPath    = string
	HttpMethod = string

	Handler          func(c Context) error
	HTTPErrorHandler func(error, Context)

	Renderer interface {
		Render(io.Writer, string, interface{}, Context) error
	}
	Router interface {
		GET(path string, handler Handler)
		POST(path string, handler Handler)
		DELETE(path string, handler Handler)
		PUT(path string, handler Handler)
		CONNECT(path string, handler Handler)
		HEAD(path string, handler Handler)
		OPTIONS(path string, handler Handler)
		PATCH(path string, handler Handler)
		TRACE(path string, handler Handler)
	}

	MethodDesc struct {
		MethodName string
		HttpMethod HttpMethod
		Path       URLPath
		Handler    Handler
	}

	ServiceDesc struct {
		ServiceName string
		HandlerType interface{}
		Methods     []MethodDesc
		Metadata    interface{}
	}

	service struct {
		server interface{}
		md     map[string]*MethodDesc
		mdata  interface{}
	}

	Echo struct {
		addr     string
		services map[string]*service
		trees    map[HttpMethod]*Tree
		render   Renderer

		Listener         net.Listener
		Debug            bool
		Logger           log.Logger
		HTTPErrorHandler HTTPErrorHandler
	}

	Group struct {
		prefix string
		r      Router
	}
)

func (g *Group) GET(path string, handler Handler) {
	path = g.prefix + path
	g.r.GET(path, handler)
}

func (g Group) POST(path string, handler Handler) {
	path = g.prefix + path
	g.r.POST(path, handler)
}

func (g *Group) DELETE(path string, handler Handler) {
	path = g.prefix + path
	g.r.DELETE(path, handler)
}

func (g *Group) PUT(path string, handler Handler) {
	path = g.prefix + path
	g.r.PUT(path, handler)
}

func (g *Group) CONNECT(path string, handler Handler) {
	path = g.prefix + path
	g.r.DELETE(path, handler)
}

func (g *Group) HEAD(path string, handler Handler) {
	path = g.prefix + path
	g.r.HEAD(path, handler)
}
func (g Group) OPTIONS(path string, handler Handler) {
	path = g.prefix + path
	g.r.OPTIONS(path, handler)
}

func (g *Group) PATCH(path string, handler Handler) {
	path = g.prefix + path
	g.r.PATCH(path, handler)
}

func (g *Group) TRACE(path string, handler Handler) {
	path = g.prefix + path
	g.r.TRACE(path, handler)
}

func (e *Echo) GET(path string, handler Handler) {
	e.trees[GET].Insert(path, handler)
}

func (e *Echo) POST(path string, handler Handler) {
	e.trees[POST].Insert(path, handler)
}

func (e *Echo) DELETE(path string, handler Handler) {
	e.trees[DELETE].Insert(path, handler)
}

func (e *Echo) PUT(path string, handler Handler) {
	e.trees[PUT].Insert(path, handler)
}

func (e *Echo) CONNECT(path string, handler Handler) {
	e.trees[CONNECT].Insert(path, handler)
}

func (e *Echo) HEAD(path string, handler Handler) {
	e.trees[HEAD].Insert(path, handler)
}

func (e *Echo) OPTIONS(path string, handler Handler) {
	e.trees[OPTIONS].Insert(path, handler)
}

func (e *Echo) PATCH(path string, handler Handler) {
	e.trees[PATCH].Insert(path, handler)
}

func (e *Echo) TRACE(path string, handler Handler) {
	e.trees[TRACE].Insert(path, handler)
}

func New() *Echo {
	e := &Echo{
		services: map[string]*service{},
		trees:    map[HttpMethod]*Tree{},
	}
	for _, m := range SupportedMethods {
		e.trees[m] = NewRadixTree()
	}
	return e
}

func (e *Echo) Start(addr string) error {
	e.addr = addr
	fmt.Printf(banner, Version, addr)
	err := http.Serve(e.Listener, e)
	if err != nil {
		log.Fatal("start echo server failed")
	}
	return err
}

func (e *Echo) ListenAndServe(addr string) error {
	fmt.Printf(banner, Version, addr)
	err := http.ListenAndServe(addr, e)
	if err != nil {
		log.Fatal("start echo server failed")
	}
	return err
}

func (e *Echo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	path := r.URL.Path
	ctx := &context{
		request:  r,
		response: w,
		values:   map[string]interface{}{},
	}
	if _, ok := e.trees[method]; !ok {
		ctx.String(http.StatusMethodNotAllowed, "unsupported "+method)
		return
	}
	v, ok, args := e.trees[method].Get(path)
	if !ok {
		ctx.String(http.StatusNotFound, "can't find page: "+path)
		return
	}
	ctx.pathParams = args
	v.(Handler)(ctx)
}

func (e *Echo) RegisterService(sd *ServiceDesc, ss interface{}) error {
	srv := &service{
		server: ss,
		md:     make(map[string]*MethodDesc),
		mdata:  sd.Metadata,
	}
	for i := range sd.Methods {
		d := &sd.Methods[i]
		srv.md[d.MethodName] = d
	}

	e.services[sd.ServiceName] = srv

	// update radix tree
	for _, m := range sd.Methods {
		if _, ok := e.trees[m.HttpMethod]; !ok {
			continue
		}
		e.trees[m.HttpMethod].Insert(m.Path, m.Handler)
	}
	return nil
}

func (e *Echo) Group(prefix string) *Group {
	g := &Group{
		prefix: prefix,
		r:      e,
	}
	return g
}
