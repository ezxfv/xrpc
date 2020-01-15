package xrpc

import (
	"context"
	"reflect"
	"sync"

	"github.com/edenzhong7/xrpc/pkg/net"

	"github.com/edenzhong7/xrpc/middleware"
	"github.com/edenzhong7/xrpc/pkg/transport"
	"github.com/edenzhong7/xrpc/plugin"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

type UnaryServerInfo = grpc.UnaryServerInfo
type UnaryServerInterceptor = grpc.UnaryServerInterceptor

type UnaryHandler func(ctx context.Context, req interface{}) (interface{}, error)

func NewServer() *Server {
	s := &Server{
		m:  map[string]*service{},
		mu: &sync.Mutex{},
	}
	return s
}

// service consists of the information of the server serving this service and
// the methods in this service.
type service struct {
	server interface{} // the server for service methods
	md     map[string]*MethodDesc
	sd     map[string]*StreamDesc
	mdata  interface{}
}

type Server struct {
	opts *options

	serve  bool
	m      map[string]*service // service name -> service info
	ctx    context.Context
	cancel context.CancelFunc

	lis         map[net.Listener]bool
	conns       map[net.Conn]bool
	middlewares []middleware.ServerMiddleware
	plugins     map[string]plugin.Plugin

	mu       *sync.Mutex
	cv       *sync.Cond
	quit     chan struct{}
	done     chan struct{}
	quitOnce sync.Once
	doneOnce sync.Once
}

func (s *Server) Serve(lis net.Listener) (err error) {
	return
}

func (s *Server) RegisterService(sd *ServiceDesc, ss interface{}) {
	ht := reflect.TypeOf(sd.HandlerType).Elem()
	st := reflect.TypeOf(ss)
	if !st.Implements(ht) {
		grpclog.Fatalf("grpc: Server.RegisterService found the handler of type %v that does not satisfy %v", st, ht)
	}
	s.register(sd, ss)
}

func (s *Server) AddMiddleware(ms ...middleware.ServerMiddleware) {
	s.middlewares = append(s.middlewares, ms...)
}

func (s *Server) register(sd *ServiceDesc, ss interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.serve {
		grpclog.Fatalf("grpc: Server.RegisterService after Server.Serve for %q", sd.ServiceName)
	}
	if _, ok := s.m[sd.ServiceName]; ok {
		grpclog.Fatalf("grpc: Server.RegisterService found duplicate service registration for %q", sd.ServiceName)
	}
	srv := &service{
		server: ss,
		md:     make(map[string]*MethodDesc),
		sd:     make(map[string]*StreamDesc),
		mdata:  sd.Metadata,
	}
	for i := range sd.Methods {
		d := &sd.Methods[i]
		srv.md[d.MethodName] = d
	}
	for i := range sd.Streams {
		d := &sd.Streams[i]
		srv.sd[d.StreamName] = d
	}
	s.m[sd.ServiceName] = srv
}

func (s *Server) processStreamingRPC(t transport.ServerTransport, stream *transport.Stream, srv *service, sd *StreamDesc) (err error) {
	return
}
