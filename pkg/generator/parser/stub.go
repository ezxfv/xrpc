package parser

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

var (
	contextPkg = "context"
	xrpcPkg    = "xrpc"
)

type StubBuilder interface {
	ClientStub(meta *MetaData, gen *Generator) error
	ServerStub(meta *MetaData, gen *Generator) error
}

func NewXrpcStubBuilder() StubBuilder {
	return &xrpcStubBuilder{}
}

type Generator struct {
	w     *bytes.Buffer
	ident int
	c     int
}

func (g *Generator) P(args ...interface{}) {
	s := fmt.Sprint(args...)
	s = strings.TrimSpace(s)

	if len(s) > 0 {
		if strings.HasSuffix(s, "{") {
			for i := 0; i < g.ident; i++ {
				g.w.WriteString(" ")
			}
			g.w.WriteString(s)
			g.ident += 4
			g.c++
		} else if strings.HasPrefix(s, "}") {
			if g.c > 0 {
				g.c--
				g.ident -= 4
			}
			if g.c == 0 {
				g.ident = 0
			}
			for i := 0; i < g.ident; i++ {
				g.w.WriteString(" ")
			}
			g.w.WriteString(s)
		} else {
			for i := 0; i < g.ident; i++ {
				g.w.WriteString(" ")
			}
			g.w.WriteString(s)
		}
	}
	g.w.WriteString("\n")
}

func (g *Generator) String() string {
	return g.w.String()
}

type xrpcStubBuilder struct {
}

func (b *xrpcStubBuilder) ClientStub(meta *MetaData, x *Generator) error {
	// TODO Client interface.
	for _, service := range meta.Interfaces() {
		servName := service.Name
		x.P(fmt.Sprintf(`// %sClient is the client API for %s service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/github.com/edenzhong7/xrpc#ClientConn.NewStream.`, servName, servName))
		x.P("type ", servName, "Client interface {")
		for _, method := range service.AllMethods() {
			//method.Params = append(method.Params, &ArgBlock{
			//	Names: []string{"opts"},
			//	Type:  "...xrpc.CallOption",
			//})
			x.P(method.String())
		}
		x.P("}")
		x.P()
		// TODO Client structure.
		x.P("type ", unexport(servName), "Client struct {")
		x.P("cc *xrpc.ClientConn")
		x.P("}")
		x.P()
		// TODO NewClient factory.
		x.P("func New", servName, "Client(cc *xrpc.ClientConn) ", servName, "Client {")
		x.P("return &", unexport(servName), "Client{cc}")
		x.P("}")
		x.P()
		// TODO Client method implementations.
		for _, method := range service.AllMethods() {
			x.P("func (c *", unexport(servName), "Client) ", method.String(), " {")
			x.P(fmt.Sprintf(`panic("unimplemented client method: %s")`, method.Name))
			x.P("}")
			x.P()
		}
	}
	return nil
}

func (b *xrpcStubBuilder) ServerStub(meta *MetaData, x *Generator) error {
	for _, service := range meta.Interfaces() {
		servName := service.Name
		serviceDescVar := "_" + servName + "_serviceDesc"
		fullServName := fmt.Sprintf("%s.%s", meta.Name(), service.Name)

		// TODO Server Unimplemented struct for forward compatibility.
		x.P("// Unimplemented", servName, " can be embedded to have forward compatible implementations.")
		x.P("type Unimplemented", servName, " struct {")
		x.P("}")
		x.P()
		for _, method := range service.AllMethods() {
			x.P("func (*Unimplemented", servName, ") ", method.String(), " {")
			x.P(`panic(fmt.Sprint(codes.Unimplemented, "method `, method.Name, ` not implemented"))`)
			x.P("}")
			x.P()
		}
		// TODO Server registration.
		x.P("func Register", servName, "Server(s *xrpc.Server, srv ", servName, ") {")
		x.P("s.RegisterService(&", serviceDescVar, `, srv)`)
		x.P("}")
		x.P()

		// TODO Server handler implementations.
		var handlerNames []string
		for _, method := range service.AllMethods() {
			methName := method.Name
			hname := fmt.Sprintf("_%s_%s_Handler", servName, methName)
			inType := method.Params[0].Type
			x.P("func ", hname, "(srv interface{}, ctx ", contextPkg, ".Context, dec func(interface{}) error, interceptor ", xrpcPkg, ".UnaryServerInterceptor) (interface{}, error) {")
			x.P("in := new(", inType, ")")
			x.P("if err := dec(in); err != nil { return nil, err }")
			x.P("if interceptor == nil { _ = srv.(", servName, ").", methName, " }")
			x.P("info := &", xrpcPkg, ".UnaryServerInfo{")
			x.P("Server: srv,")
			x.P("FullMethod: ", strconv.Quote(fmt.Sprintf("/%s/%s", fullServName, methName)), ",")
			x.P("}")
			x.P("handler := func(ctx ", contextPkg, ".Context, req interface{}) (interface{}, error) {")
			x.P("_ = srv.(", servName, ").", methName)
			x.P(`panic("gg")`)
			x.P("}")
			x.P("return interceptor(ctx, in, info, handler)")
			x.P("}")
			x.P()
			handlerNames = append(handlerNames, hname)
		}
		// TODO Service descriptor.
		x.P("var ", serviceDescVar, " = ", xrpcPkg, ".ServiceDesc {")
		x.P("ServiceName: ", strconv.Quote(fullServName), ",")
		x.P("HandlerType: (*", servName, ")(nil),")
		x.P("Methods: []", xrpcPkg, ".MethodDesc{")
		for i, method := range service.AllMethods() {
			x.P("{")
			x.P("MethodName: ", strconv.Quote(method.Name), ",")
			x.P("Handler: ", handlerNames[i], ",")
			x.P("},")
		}
		x.P("},")
		x.P("Streams: []", xrpcPkg, ".StreamDesc{},")
		x.P("Metadata: \"", meta.Name(), "\",")
		x.P("}")
		x.P()
	}

	return nil
}

func unexport(s string) string { return strings.ToLower(s[:1]) + s[1:] }
