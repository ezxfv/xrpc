package echo

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"x.io/xrpc/pkg/log"

	"github.com/patrickmn/go-cache"
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
		HandleFunc(path string, handler Handler, ms ...HttpMethod)
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
		ms       []Handler
		routes   []string

		enableCache      bool
		routeCache       *cache.Cache
		cacheExpiredTime time.Duration

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

func New() *Echo {
	e := &Echo{
		services:         map[string]*service{},
		trees:            map[HttpMethod]*Tree{},
		routeCache:       cache.New(time.Second*10, time.Minute),
		cacheExpiredTime: time.Second * 5,
	}
	for _, m := range SupportedMethods {
		e.trees[m] = NewRadixTree()
	}
	return e
}

func (g *Group) HandleFunc(path string, handler Handler) {
	path = g.prefix + path
	g.r.HandleFunc(path, handler)
}
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

func (e *Echo) HandleFunc(path string, handler Handler, ms ...HttpMethod) {
	if len(ms) == 0 {
		ms = SupportedMethods
	}
	for _, method := range ms {
		e.trees[method].Insert(path, handler)
	}
}

func (e *Echo) Cache(enable bool) {
	e.enableCache = enable
}

func (e *Echo) GET(path string, handler Handler) {
	e.addRoute(GET, path, handler)
}

func (e *Echo) POST(path string, handler Handler) {
	e.addRoute(POST, path, handler)
}

func (e *Echo) DELETE(path string, handler Handler) {
	e.addRoute(DELETE, path, handler)
}

func (e *Echo) PUT(path string, handler Handler) {
	e.addRoute(PUT, path, handler)
}

func (e *Echo) CONNECT(path string, handler Handler) {
	e.addRoute(CONNECT, path, handler)
}

func (e *Echo) HEAD(path string, handler Handler) {
	e.addRoute(HEAD, path, handler)
}

func (e *Echo) OPTIONS(path string, handler Handler) {
	e.addRoute(OPTIONS, path, handler)
}

func (e *Echo) PATCH(path string, handler Handler) {
	e.addRoute(PATCH, path, handler)
}

func (e *Echo) TRACE(path string, handler Handler) {
	e.addRoute(TRACE, path, handler)
}

func (e *Echo) addRoute(method HttpMethod, path string, handler Handler) {
	if t, ok := e.trees[method]; ok {
		_, ok = t.Insert(path, handler)
		if !ok {
			e.routes = append(e.routes, path)
		}
	}
}

func (e *Echo) Use(ms ...Handler) {
	e.ms = append(e.ms, ms...)
}

func (e *Echo) GetRoutes() []string {
	return e.routes
}

func (e *Echo) EnableListRoutes() {
	e.GET("/routes", func(c Context) error {
		s := strings.Join(e.GetRoutes(), "\n")
		return c.String(http.StatusOK, s)
	})
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

type routeCache struct {
	v    Handler
	args map[string]string
}

func (e *Echo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	path := r.URL.Path
	c := &context{
		request:  r,
		response: w,
		path:     r.RequestURI,
		values:   map[string]interface{}{},
	}
	if e.enableCache {
		t := time.Now()
		rc, ok := e.routeCache.Get(path)
		if ok {
			rcc, ok := rc.(*routeCache)
			if ok {
				if e.Debug {
					log.Debugf("read cache %s in %dns [%v]", c.Request().RequestURI, time.Since(t).Nanoseconds(), ok)
				}
				c.ms = e.ms
				c.pathParams = rcc.args
				c.handler = rcc.v
				c.Next()
				return
			}
		}
	}
	t := time.Now()
	if _, ok := e.trees[method]; !ok {
		c.String(http.StatusMethodNotAllowed, "unsupported "+method)
		return
	}
	v, ok, args := e.trees[method].Get(path)
	if e.Debug {
		fmt.Printf("route %s in %dns [%v]\n", c.Request().RequestURI, time.Since(t).Nanoseconds(), ok)
	}
	if !ok {
		c.String(http.StatusNotFound, "can't find page: "+path)
		return
	}
	c.ms = e.ms
	c.pathParams = args
	c.handler = v.(Handler)
	if e.enableCache {
		rc := &routeCache{
			v:    c.handler,
			args: args,
		}
		e.routeCache.Set(path, rc, e.cacheExpiredTime)
	}
	c.Next()
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
		e.addRoute(m.HttpMethod, m.Path, m.Handler)
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
