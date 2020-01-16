package xrpc

import (
	"context"
	"encoding/json"
	"log"
	"reflect"
	"sync"
	"time"

	"github.com/edenzhong7/xrpc/pkg/encoding"

	"github.com/xtaci/smux"

	"github.com/edenzhong7/xrpc/pkg/net"

	"github.com/edenzhong7/xrpc/middleware"
	"github.com/edenzhong7/xrpc/plugin"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

type UnaryServerInfo = grpc.UnaryServerInfo
type UnaryServerInterceptor = grpc.UnaryServerInterceptor

type UnaryHandler func(ctx context.Context, req interface{}) (interface{}, error)

func NewServer() *Server {
	s := &Server{
		m:           map[string]*service{},
		mu:          &sync.Mutex{},
		lis:         map[net.Listener]bool{},
		conns:       map[net.Conn]bool{},
		sessions:    map[*smux.Session]bool{},
		middlewares: []middleware.ServerMiddleware{},
		plugins:     map[string]plugin.Plugin{},
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
	sessions    map[*smux.Session]bool
	middlewares []middleware.ServerMiddleware
	plugins     map[string]plugin.Plugin

	mu       *sync.Mutex
	cv       *sync.Cond
	quit     chan struct{}
	done     chan struct{}
	quitOnce sync.Once
	doneOnce sync.Once
}

func (s *Server) Serve(lis net.Listener) error {
	go s.listen(lis)
	return nil
}

func (s *Server) Start() {
	for {
		time.Sleep(time.Millisecond * 100)
	}
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

func (s *Server) listen(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			break
		}
		p := make([]byte, len(Preface))
		n, err := conn.Read(p)
		if err != nil || n != len(Preface) {
			continue
		}
		session, err := smux.Server(conn, nil)
		s.sessions[session] = true
		go s.handleSession(session)
	}
}

func (s *Server) handleSession(session *smux.Session) {
	log.Println("handle server session")
	defer log.Println("close server session")
	for {
		stream, err := session.AcceptStream()
		if err != nil {
			break
		}
		pf, data, err := recv(stream)
		if err != nil {
			return
		}
		ss := &serverStream{
			stream: stream,
			codec:  encoding.GetCodec("proto"),
			cp:     encoding.GetCompressor("gzip"),
		}
		if pf == CmdHeader {
			header := &streamHeader{}
			err = json.Unmarshal(data, header)
			if err != nil {
				continue
			}
			ss.header = header
			go s.processStream(s.ctx, ss, header)
		}
	}
}

func (s *Server) processStream(ctx context.Context, stream ServerStream, header *streamHeader) {
	log.Println("process server stream")
	var err error
	var reply interface{}
	service, method := header.splitMethod()
	if service == "" || method == "" {
		return
	}
	srv := s.m[service].server
	desc := s.m[service].md[method]
	for {
		reply, err = desc.Handler(srv, ctx, stream.RecvMsg, nil)
		if err != nil {
			continue
		}
		if err = stream.SendMsg(reply); err != nil {
			break
		}
	}
}
