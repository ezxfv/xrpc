package xrpc

import (
	"context"
	"encoding/json"
	"reflect"
	"sync"
	"time"

	"github.com/edenzhong7/xrpc/pkg/encoding"
	"github.com/edenzhong7/xrpc/pkg/log"
	"github.com/edenzhong7/xrpc/pkg/net"
	"github.com/edenzhong7/xrpc/plugin"

	"github.com/xtaci/smux"
	"google.golang.org/grpc"
)

type (
	UnaryServerInfo        = grpc.UnaryServerInfo
	UnaryServerInterceptor = grpc.UnaryServerInterceptor
)

func NewServer() *Server {
	pc := plugin.NewPluginContainer()
	s := &Server{
		m:        map[string]*service{},
		mu:       &sync.Mutex{},
		lis:      map[net.Listener]bool{},
		conns:    map[net.Conn]bool{},
		sessions: map[*smux.Session]bool{},
		pc:       pc,
		ctx:      context.Background(),
		auth:     NewEmptyAuthenticator(),
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

	auth Authenticator

	lis      map[net.Listener]bool
	conns    map[net.Conn]bool
	sessions map[*smux.Session]bool
	pc       plugin.Container

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

func (s *Server) SetAuthenticator(authenticator Authenticator) {
	s.auth = authenticator
}

func (s *Server) Shutdown() (err error) {
	if s.pc != nil {
		err = s.pc.Stop()
	}
	return
}

func (s *Server) StartPlugins() (err error) {
	if s.pc != nil {
		err = s.pc.Start()
	}
	return
}

func (s *Server) ApplyPlugins(plugins ...plugin.Plugin) {
	for _, p := range plugins {
		s.pc.Add(p)
	}
}

func (s *Server) Start() {
	for {
		time.Sleep(time.Millisecond * 100)
	}
}

func (s *Server) RegisterFunction(serviceName, fname string, fn interface{}, metadata string) {
	// TODO DoRegisterFunction
}

func (s *Server) RegisterCustomService(sd *ServiceDesc, ss interface{}) {
	// TODO DoRegisterCustomService
}

func (s *Server) RegisterService(sd *ServiceDesc, ss interface{}) {
	ht := reflect.TypeOf(sd.HandlerType).Elem()
	st := reflect.TypeOf(ss)
	if !st.Implements(ht) {
		log.Fatalf("xrpc: Server.RegisterService found the handler of type %v that does not satisfy %v", st, ht)
	}
	s.register(sd, ss)
}

func (s *Server) register(sd *ServiceDesc, ss interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.serve {
		log.Fatalf("xrpc: Server.RegisterService after Server.Serve for %q", sd.ServiceName)
	}
	if _, ok := s.m[sd.ServiceName]; ok {
		log.Fatalf("xrpc: Server.RegisterService found duplicate service registration for %q", sd.ServiceName)
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
	// TODO DoRegister
	s.pc.DoRegisterService(sd, ss)
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
		if err != nil {
			continue
		}
		s.sessions[session] = true
		// TODO DoConnect
		conn, ok := s.pc.DoConnect(conn)
		if !ok {
			conn.Close()
			continue
		}
		go s.handleSession(conn, session)
	}
}

func (s *Server) handleSession(conn net.Conn, session *smux.Session) {
	log.Debug("handle server session")
	defer log.Debug("close server session")
	for {
		stream, err := session.AcceptStream()
		if err != nil {
			break
		}
		pf, data, err := recv(stream)
		if err != nil {
			return
		}
		if pf == cmdHeader {
			header := &streamHeader{}
			err = json.Unmarshal(data, header)
			if err != nil {
				continue
			}
			if err = s.auth.Authenticate(header.Args); err != nil {
				log.Error(err.Error())
				stream.Close()
				continue
			}
			ss := &serverStream{
				stream: &streamConn{stream},
				codec:  encoding.GetCodec(getCodecArg(header)),
				cp:     encoding.GetCompressor(getCompressorArg(header)),
				sc:     s.pc,
			}
			ss.header = header
			// TODO DoOpenStream
			if _, err = s.pc.DoOpenStream(context.Background(), stream); err != nil {
				continue
			}
			go s.processStream(s.ctx, ss, header)
		}
	}
	// TODO DoDisconnect
	s.pc.DoDisconnect(conn)
}

func (s *Server) processStream(ctx context.Context, stream ServerStream, header *streamHeader) {
	log.Debug("process server stream")
	log.Debug("close server stream")
	service, method := header.splitMethod()
	if service == "" || method == "" {
		return
	}
	srv := s.m[service].server
	desc := s.m[service].md[method]
	var newCtx context.Context

	dec := func(m interface{}) (err error) {
		newCtx, err = stream.RecvMsg(newCtx, m)
		return
	}
	for {
		newCtx = ctx
		reply, err := desc.Handler(srv, newCtx, dec, s.pc.DoHandle)
		if err != nil {
			break
		}
		if err = stream.SendMsg(newCtx, reply); err != nil {
			break
		}
	}
	// TODO DoCloseStream
	s.pc.DoCloseStream(ctx, stream.(*serverStream).stream)
}
