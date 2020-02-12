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

func (g *Generator) Tab() {
	g.ident += 4
	g.c++
}

func (g *Generator) UnTab() {
	g.ident -= 4
	g.c--
}

func (g *Generator) F(format string, args ...interface{}) {
	g.P(fmt.Sprintf(format, args...))
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

func genClientVars(method *Method) (string, string, string, string) {
	w := bytes.NewBuffer([]byte{})
	w.WriteString(method.Name + "(")
	w.WriteString("ctx context.Context")
	if len(method.Params) > 0 {
		w.WriteString(", ")
	}
	k := 1
	var ins, outs, starOuts []string
	for i, pb := range method.Params {
		if strings.Contains(pb.Type, "xrpc.XContext") {
			continue
		}
		var ns []string
		if len(pb.Names) != 0 {
			for range pb.Names {
				nn := fmt.Sprintf("in_%d", k)
				k++
				ns = append(ns, nn)
			}
		} else {
			nn := fmt.Sprintf("in_%d", k)
			k++
			ns = append(ns, nn)
		}
		ins = append(ins, ns...)
		w.WriteString(strings.Join(ns, ","))
		w.WriteString(" ")
		w.WriteString(pb.Type)
		if i < len(method.Params)-1 {
			w.WriteString(", ")
		}
	}
	w.WriteString(") ")
	if len(method.Results) == 0 {
		return w.String(), strings.Join(ins, ", "), "", ""
	}
	w.WriteString("(")
	k = 1
	for i, rb := range method.Results {
		var ns []string
		if len(rb.Names) != 0 {
			for range rb.Names {
				nn := fmt.Sprintf("out_%d", k)
				k++
				ns = append(ns, nn)
			}
		} else {
			nn := fmt.Sprintf("out_%d", k)
			k++
			ns = append(ns, nn)
		}
		for _, nn := range ns {
			//if strings.HasPrefix(rb.Type, "*") {
			//	starOuts = append(starOuts, nn)
			//} else {
			starOuts = append(starOuts, "&"+nn)
			//}
		}
		outs = append(outs, ns...)
		w.WriteString(strings.Join(ns, ","))
		w.WriteString(" ")
		w.WriteString(rb.Type)
		if i < len(method.Results)-1 {
			w.WriteString(", ")
		}
	}
	w.WriteString(")")
	return w.String(), strings.Join(ins, ", "), strings.Join(outs, ", "), strings.Join(starOuts, ", ")
}

func (b *xrpcStubBuilder) ClientStub(meta *MetaData, x *Generator) error {
	// Client interface.
	for _, service := range meta.Interfaces() {
		servName := service.Name
		fullServName := fmt.Sprintf("%s.%s", meta.Name(), servName)
		x.P(fmt.Sprintf(`// %sClient is the client API for %s service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/github.com/edenzhong7/xrpc#ClientConn.NewStream.`, servName, servName))
		x.P("type ", servName, "Client interface {")
		for _, method := range service.AllMethods() {
			//method.Params = append(method.Params, &ArgBlock{
			//	Names: []string{"opts"},
			//	Type:  "...xrpc.CallOption",
			//})
			funcSign, _, _, _ := genClientVars(method)
			x.P(funcSign)
		}
		x.P("}")
		x.P()
		// Client structure.
		x.P("type ", unexport(servName), "Client struct {")
		x.P("cc *xrpc.ClientConn")
		x.P("opts []xrpc.CallOption")
		x.P("}")
		x.P()
		// NewClient factory.
		x.P("func New", servName, "Client(cc *xrpc.ClientConn, opts ...xrpc.CallOption) ", servName, "Client {")
		x.P("return &", unexport(servName), "Client{cc, opts}")
		x.P("}")
		x.P()
		// Client method implementations.
		for _, method := range service.AllMethods() {
			funcSign, ins, outs, starOuts := genClientVars(method)
			// 命名参数
			x.F("func (c *%sClient) %s {", unexport(servName), funcSign)
			x.P("var ins, outs []interface{}")
			if len(ins) > 0 {
				x.F("ins = append(ins, %s)", ins)
			}
			if len(starOuts) > 0 {
				x.F("outs = append(outs, %s)", starOuts)
			}
			x.P(`err := c.cc.Invoke(ctx, "`, fmt.Sprintf("/%s/%s", fullServName, method.Name), `", ins, &outs, c.opts...)`)
			x.F("if err != nil { return %s }", outs)
			x.P("return ", outs)
			// 反序列化结果
			x.P("}")
			x.P()
		}
	}
	return nil
}

func genServerVars(method *Method, x *Generator) (string, string) {
	x.P("var ins []interface{}")
	var paramsNames []string
	var var_args bool
	var resultsNames []string
	var hasCtx bool
	if len(method.Params) > 0 {
		ins_str := ""
		x.P("var (")
		x.Tab()
		k := 1
		for _, p := range method.Params {
			var t string
			var ctx bool
			if strings.HasPrefix(p.Type, "*") {
				if strings.Contains(p.Type, "XContext") {
					t = " = xrpc.XBackground()"
					ctx = true
				} else {
					t = "= new(" + p.Type[1:] + ")"
				}
			} else if strings.HasPrefix(p.Type, "...") {
				t = "[]" + p.Type[3:]
				var_args = true
			} else {
				t = p.Type
			}
			var in_name string
			if ctx {
				in_name = "xctx"
				x.F("%s %s", in_name, t)
				paramsNames = append(paramsNames, in_name)
				hasCtx = true
				continue
			}
			if len(p.Names) == 0 {
				in_name = fmt.Sprintf("in_%d", k)
				x.F("in_%d %s", k, t)
				paramsNames = append(paramsNames, in_name)
				if strings.HasPrefix(p.Type, "*") {
					ins_str = ins_str + in_name + ", "
				} else {
					ins_str += ins_str + "&" + in_name + ", "
				}
				k++
			} else {
				for range p.Names {
					in_name = fmt.Sprintf("in_%d", k)
					x.F("in_%d %s", k, t)
					paramsNames = append(paramsNames, in_name)
					if strings.HasPrefix(p.Type, "*") {
						ins_str = ins_str + in_name + ", "
					} else {
						ins_str = ins_str + "&" + in_name + ", "
					}
					k++
				}
			}
		}
		x.UnTab()
		x.P(")")
		if hasCtx {
			x.P("xctx.SetCtx(ctx)")
		}
		end := len(ins_str) - 2
		x.F("ins = append(ins, %s)", ins_str[:end])
	}
	if len(method.Results) > 0 {
		x.P("var (")
		x.Tab()
		k := 1
		for _, p := range method.Results {
			var t string
			if strings.HasPrefix(p.Type, "*") {
				t = "= new(" + p.Type[1:] + ")"
			} else {
				t = p.Type
			}
			if len(p.Names) == 0 {
				x.F("out_%d %s", k, t)
				resultsNames = append(resultsNames, fmt.Sprintf("out_%d", k))
				k++
			} else {
				for range p.Names {
					x.F("out_%d %s", k, t)
					resultsNames = append(resultsNames, fmt.Sprintf("out_%d", k))
					k++
				}
			}
		}
		x.UnTab()
		x.P(")")
	}
	ins := strings.Join(paramsNames, ", ")
	if var_args {
		ins += "..."
	}
	outs := strings.Join(resultsNames, ", ")
	return ins, outs
}

func (b *xrpcStubBuilder) ServerStub(meta *MetaData, x *Generator) error {
	for _, service := range meta.Interfaces() {
		servName := service.Name
		serviceDescVar := "_" + servName + "_serviceDesc"
		fullServName := fmt.Sprintf("%s.%s", meta.Name(), service.Name)

		// Server Unimplemented struct for forward compatibility.
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
		// Server registration.
		x.P("func Register", servName, "Server(s *xrpc.Server, srv ", servName, ") {")
		x.P("s.RegisterService(&", serviceDescVar, `, srv)`)
		x.P("}")
		x.P()

		// handler implementations.
		var handlerNames []string
		for _, method := range service.AllMethods() {
			methName := method.Name
			hname := fmt.Sprintf("_%s_%s_Handler", servName, methName)
			x.P("func ", hname, "(srv interface{}, ctx ", contextPkg, ".Context, dec func(interface{}) error, interceptor ", xrpcPkg, ".UnaryServerInterceptor) (interface{}, error) {")
			ins, outs := genServerVars(method, x)
			if len(ins) > 0 {
				x.P("if err := dec(&ins); err != nil { return nil, err }")
			}
			x.F("if interceptor == nil { ")
			if len(outs) > 0 {
				x.P("var results []interface{}")
				x.F("%s = srv.(%s).%s(%s)", outs, servName, methName, ins)
				x.F("results = append(results, %s)", outs)
				x.F("return results, nil")
			} else {
				x.F("srv.(%s).%s(%s)", servName, methName, ins)
				x.P("return nil, nil")
			}
			x.P("}")
			x.P("info := &", xrpcPkg, ".UnaryServerInfo{")
			x.P("Server: srv,")
			x.P("FullMethod: ", strconv.Quote(fmt.Sprintf("/%s/%s", fullServName, methName)), ",")
			x.P("}")
			x.P("handler := func(ctx ", contextPkg, ".Context, req interface{}) (interface{}, error) {")
			if len(outs) > 0 {
				x.P("var results []interface{}")
				x.F("%s = srv.(%s).%s(%s)", outs, servName, methName, ins)
				x.F("results = append(results, %s)", outs)
				x.F("return results, nil")
			} else {
				x.F("srv.(%s).%s(%s)", servName, methName, ins)
				x.P("return nil, nil")
			}
			x.P("}")
			x.P("return interceptor(ctx, ins, info, handler)")
			x.P("}")
			x.P()
			handlerNames = append(handlerNames, hname)
		}
		// Service descriptor.
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
