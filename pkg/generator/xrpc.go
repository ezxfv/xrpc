// Go support for Protocol Buffers - Google's data interchange format
//
// Copyright 2015 The Go Authors.  All rights reserved.
// https://github.com/golang/protobuf
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//     * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Package xrpc outputs gRPC service descriptions in Go code.
// It runs as a plugin for the Go protocol buffer compiler plugin.
// It is linked in to protoc-gen-go.
package generator

import (
	"fmt"
	"strconv"
	"strings"

	pb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/generator"
)

// generatedCodeVersion indicates a version of the generated code.
// It is incremented whenever an incompatibility between the generated code and
// the xrpc package is introduced; the generated code references
// a constant, xrpc.SupportPackageIsVersionN (where N is generatedCodeVersion).
const generatedCodeVersion = 4

// Paths for packages used by code generated in this file,
// relative to the import_prefix of the generator.Generator.
const (
	contextPkgPath = "context"
	xrpcPkgPath    = "github.com/edenzhong7/xrpc"
	statusPkgPath  = "github.com/edenzhong7/xrpc/pkg/status"
	codesPkgPath   = "github.com/edenzhong7/xrpc/pkg/codes"
)

func init() {
	generator.RegisterPlugin(new(xrpc))
}

// xrpc is an implementation of the Go protocol buffer compiler's
// plugin architecture.  It generates bindings for gRPC support.
type xrpc struct {
	gen *generator.Generator
}

// name returns the name of this plugin, "xrpc".
func (x *xrpc) Name() string {
	return "xrpc"
}

// The names for packages imported in the generated code.
// They may vary from the final path component of the import path
// if the name is used by other packages.
var (
	contextPkg string
	xrpcPkg    string
)

// Init initializes the plugin.
func (x *xrpc) Init(gen *generator.Generator) {
	x.gen = gen
}

// Given a type name defined in a .proto, return its object.
// Also record that we're using it, to guarantee the associated import.
func (x *xrpc) objectNamed(name string) generator.Object {
	x.gen.RecordTypeUse(name)
	return x.gen.ObjectNamed(name)
}

// Given a type name defined in a .proto, return its name as we will print it.
func (x *xrpc) typeName(str string) string {
	return x.gen.TypeName(x.objectNamed(str))
}

// P forwards to g.gen.P.
func (x *xrpc) P(args ...interface{}) { x.gen.P(args...) }

// Generate generates code for the services in the given file.
func (x *xrpc) Generate(file *generator.FileDescriptor) {
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}

	contextPkg = string(x.gen.AddImport(contextPkgPath))
	xrpcPkg = string(x.gen.AddImport(xrpcPkgPath))

	x.P("// Reference imports to suppress errors if they are not otherwise used.")
	x.P("var _ ", contextPkg, ".Context")
	x.P("var _ ", xrpcPkg, ".ClientConn")
	x.P()

	// Assert version compatibility.
	x.P("// This is a compile-time assertion to ensure that this generated file")
	x.P("// is compatible with the xrpc package it is being compiled against.")
	x.P("const _ = ", xrpcPkg, ".SupportPackageIsVersion", generatedCodeVersion)
	x.P()

	for i, service := range file.FileDescriptorProto.Service {
		x.generateService(file, service, i)
	}
}

// GenerateImports generates the import declaration for this file.
func (x *xrpc) GenerateImports(file *generator.FileDescriptor) {
}

// reservedClientName records whether a client name is reserved on the client side.
var reservedClientName = map[string]bool{
	// TODO: do we need any in gRPC?
}

func unexport(s string) string { return strings.ToLower(s[:1]) + s[1:] }

// deprecationComment is the standard comment added to deprecated
// messages, fields, enums, and enum values.
var deprecationComment = "// Deprecated: Do not use."

// generateService generates all the code for the named service.
func (x *xrpc) generateService(file *generator.FileDescriptor, service *pb.ServiceDescriptorProto, index int) {
	path := fmt.Sprintf("6,%d", index) // 6 means service.

	origServName := service.GetName()
	fullServName := origServName
	if pkg := file.GetPackage(); pkg != "" {
		fullServName = pkg + "." + fullServName
	}
	servName := generator.CamelCase(origServName)
	deprecated := service.GetOptions().GetDeprecated()

	x.P()
	x.P(fmt.Sprintf(`// %sClient is the client API for %s service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/github.com/edenzhong7/xrpc#ClientConn.NewStream.`, servName, servName))

	// Client interface.
	if deprecated {
		x.P("//")
		x.P(deprecationComment)
	}
	x.P("type ", servName, "Client interface {")
	for i, method := range service.Method {
		x.gen.PrintComments(fmt.Sprintf("%s,2,%d", path, i)) // 2 means method in a service.
		if method.GetOptions().GetDeprecated() {
			x.P("//")
			x.P(deprecationComment)
		}
		x.P(x.generateClientSignature(servName, method))
	}
	x.P("}")
	x.P()

	// Client structure.
	x.P("type ", unexport(servName), "Client struct {")
	x.P("cc *", xrpcPkg, ".ClientConn")
	x.P("}")
	x.P()

	// NewClient factory.
	if deprecated {
		x.P(deprecationComment)
	}
	x.P("func New", servName, "Client (cc *", xrpcPkg, ".ClientConn) ", servName, "Client {")
	x.P("return &", unexport(servName), "Client{cc}")
	x.P("}")
	x.P()

	var methodIndex, streamIndex int
	serviceDescVar := "_" + servName + "_serviceDesc"
	// Client method implementations.
	for _, method := range service.Method {
		var descExpr string
		if !method.GetServerStreaming() && !method.GetClientStreaming() {
			// Unary RPC method
			descExpr = fmt.Sprintf("&%s.Methods[%d]", serviceDescVar, methodIndex)
			methodIndex++
		} else {
			// Streaming RPC method
			descExpr = fmt.Sprintf("&%s.Streams[%d]", serviceDescVar, streamIndex)
			streamIndex++
		}
		x.generateClientMethod(servName, fullServName, serviceDescVar, method, descExpr)
	}

	// Server interface.
	serverType := servName + "Server"
	x.P("// ", serverType, " is the server API for ", servName, " service.")
	if deprecated {
		x.P("//")
		x.P(deprecationComment)
	}
	x.P("type ", serverType, " interface {")
	for i, method := range service.Method {
		x.gen.PrintComments(fmt.Sprintf("%s,2,%d", path, i)) // 2 means method in a service.
		if method.GetOptions().GetDeprecated() {
			x.P("//")
			x.P(deprecationComment)
		}
		x.P(x.generateServerSignature(servName, method))
	}
	x.P("}")
	x.P()

	// Server Unimplemented struct for forward compatibility.
	if deprecated {
		x.P(deprecationComment)
	}
	x.generateUnimplementedServer(servName, service)

	// Server registration.
	if deprecated {
		x.P(deprecationComment)
	}
	x.P("func Register", servName, "Server(s *", xrpcPkg, ".Server, srv ", serverType, ") {")
	x.P("s.RegisterService(&", serviceDescVar, `, srv)`)
	x.P("}")
	x.P()

	// Server handler implementations.
	var handlerNames []string
	for _, method := range service.Method {
		hname := x.generateServerMethod(servName, fullServName, method)
		handlerNames = append(handlerNames, hname)
	}

	// Service descriptor.
	x.P("var ", serviceDescVar, " = ", xrpcPkg, ".ServiceDesc {")
	x.P("ServiceName: ", strconv.Quote(fullServName), ",")
	x.P("HandlerType: (*", serverType, ")(nil),")
	x.P("Methods: []", xrpcPkg, ".MethodDesc{")
	for i, method := range service.Method {
		if method.GetServerStreaming() || method.GetClientStreaming() {
			continue
		}
		x.P("{")
		x.P("MethodName: ", strconv.Quote(method.GetName()), ",")
		x.P("Handler: ", handlerNames[i], ",")
		x.P("},")
	}
	x.P("},")
	x.P("Streams: []", xrpcPkg, ".StreamDesc{")
	for i, method := range service.Method {
		if !method.GetServerStreaming() && !method.GetClientStreaming() {
			continue
		}
		x.P("{")
		x.P("StreamName: ", strconv.Quote(method.GetName()), ",")
		x.P("Handler: ", handlerNames[i], ",")
		if method.GetServerStreaming() {
			x.P("ServerStreams: true,")
		}
		if method.GetClientStreaming() {
			x.P("ClientStreams: true,")
		}
		x.P("},")
	}
	x.P("},")
	x.P("Metadata: \"", file.GetName(), "\",")
	x.P("}")
	x.P()
}

// generateUnimplementedServer creates the unimplemented server struct
func (x *xrpc) generateUnimplementedServer(servName string, service *pb.ServiceDescriptorProto) {
	serverType := servName + "Server"
	x.P("// Unimplemented", serverType, " can be embedded to have forward compatible implementations.")
	x.P("type Unimplemented", serverType, " struct {")
	x.P("}")
	x.P()
	// Unimplemented<service_name>Server's concrete methods
	for _, method := range service.Method {
		x.generateServerMethodConcrete(servName, method)
	}
	x.P()
}

// generateServerMethodConcrete returns unimplemented methods which ensure forward compatibility
func (x *xrpc) generateServerMethodConcrete(servName string, method *pb.MethodDescriptorProto) {
	header := x.generateServerSignatureWithParamNames(servName, method)
	x.P("func (*Unimplemented", servName, "Server) ", header, " {")
	var nilArg string
	if !method.GetServerStreaming() && !method.GetClientStreaming() {
		nilArg = "nil, "
	}
	methName := generator.CamelCase(method.GetName())
	statusPkg := string(x.gen.AddImport(statusPkgPath))
	codesPkg := string(x.gen.AddImport(codesPkgPath))
	x.P("return ", nilArg, statusPkg, `.Errorf(`, codesPkg, `.Unimplemented, "method `, methName, ` not implemented")`)
	x.P("}")
}

// generateClientSignature returns the client-side signature for a method.
func (x *xrpc) generateClientSignature(servName string, method *pb.MethodDescriptorProto) string {
	origMethName := method.GetName()
	methName := generator.CamelCase(origMethName)
	if reservedClientName[methName] {
		methName += "_"
	}
	reqArg := ", in *" + x.typeName(method.GetInputType())
	if method.GetClientStreaming() {
		reqArg = ""
	}
	respName := "*" + x.typeName(method.GetOutputType())
	if method.GetServerStreaming() || method.GetClientStreaming() {
		respName = servName + "_" + generator.CamelCase(origMethName) + "Client"
	}
	return fmt.Sprintf("%s(ctx %s.Context%s, opts ...%s.CallOption) (%s, error)", methName, contextPkg, reqArg, xrpcPkg, respName)
}

func (x *xrpc) generateClientMethod(servName, fullServName, serviceDescVar string, method *pb.MethodDescriptorProto, descExpr string) {
	sname := fmt.Sprintf("/%s/%s", fullServName, method.GetName())
	methName := generator.CamelCase(method.GetName())
	inType := x.typeName(method.GetInputType())
	outType := x.typeName(method.GetOutputType())

	if method.GetOptions().GetDeprecated() {
		x.P(deprecationComment)
	}
	x.P("func (c *", unexport(servName), "Client) ", x.generateClientSignature(servName, method), "{")
	if !method.GetServerStreaming() && !method.GetClientStreaming() {
		x.P("out := new(", outType, ")")
		// TODO: Pass descExpr to Invoke.
		x.P(`err := c.cc.Invoke(ctx, "`, sname, `", in, out, opts...)`)
		x.P("if err != nil { return nil, err }")
		x.P("return out, nil")
		x.P("}")
		x.P()
		return
	}
	streamType := unexport(servName) + methName + "Client"
	x.P("stream, err := c.cc.NewStream(ctx, ", descExpr, `, "`, sname, `", opts...)`)
	x.P("if err != nil { return nil, err }")
	x.P("x := &", streamType, "{stream}")
	if !method.GetClientStreaming() {
		x.P("if err := x.ClientStream.SendMsg(in); err != nil { return nil, err }")
		x.P("if err := x.ClientStream.CloseSend(); err != nil { return nil, err }")
	}
	x.P("return x, nil")
	x.P("}")
	x.P()

	genSend := method.GetClientStreaming()
	genRecv := method.GetServerStreaming()
	genCloseAndRecv := !method.GetServerStreaming()

	// Stream auxiliary types and methods.
	x.P("type ", servName, "_", methName, "Client interface {")
	if genSend {
		x.P("Send(*", inType, ") error")
	}
	if genRecv {
		x.P("Recv() (*", outType, ", error)")
	}
	if genCloseAndRecv {
		x.P("CloseAndRecv() (*", outType, ", error)")
	}
	x.P(xrpcPkg, ".ClientStream")
	x.P("}")
	x.P()

	x.P("type ", streamType, " struct {")
	x.P(xrpcPkg, ".ClientStream")
	x.P("}")
	x.P()

	if genSend {
		x.P("func (x *", streamType, ") Send(m *", inType, ") error {")
		x.P("return x.ClientStream.SendMsg(m)")
		x.P("}")
		x.P()
	}
	if genRecv {
		x.P("func (x *", streamType, ") Recv() (*", outType, ", error) {")
		x.P("m := new(", outType, ")")
		x.P("if err := x.ClientStream.RecvMsg(m); err != nil { return nil, err }")
		x.P("return m, nil")
		x.P("}")
		x.P()
	}
	if genCloseAndRecv {
		x.P("func (x *", streamType, ") CloseAndRecv() (*", outType, ", error) {")
		x.P("if err := x.ClientStream.CloseSend(); err != nil { return nil, err }")
		x.P("m := new(", outType, ")")
		x.P("if err := x.ClientStream.RecvMsg(m); err != nil { return nil, err }")
		x.P("return m, nil")
		x.P("}")
		x.P()
	}
}

// generateServerSignatureWithParamNames returns the server-side signature for a method with parameter names.
func (x *xrpc) generateServerSignatureWithParamNames(servName string, method *pb.MethodDescriptorProto) string {
	origMethName := method.GetName()
	methName := generator.CamelCase(origMethName)
	if reservedClientName[methName] {
		methName += "_"
	}

	var reqArgs []string
	ret := "error"
	if !method.GetServerStreaming() && !method.GetClientStreaming() {
		reqArgs = append(reqArgs, "ctx "+contextPkg+".Context")
		ret = "(*" + x.typeName(method.GetOutputType()) + ", error)"
	}
	if !method.GetClientStreaming() {
		reqArgs = append(reqArgs, "req *"+x.typeName(method.GetInputType()))
	}
	if method.GetServerStreaming() || method.GetClientStreaming() {
		reqArgs = append(reqArgs, "srv "+servName+"_"+generator.CamelCase(origMethName)+"Server")
	}

	return methName + "(" + strings.Join(reqArgs, ", ") + ") " + ret
}

// generateServerSignature returns the server-side signature for a method.
func (x *xrpc) generateServerSignature(servName string, method *pb.MethodDescriptorProto) string {
	origMethName := method.GetName()
	methName := generator.CamelCase(origMethName)
	if reservedClientName[methName] {
		methName += "_"
	}

	var reqArgs []string
	ret := "error"
	if !method.GetServerStreaming() && !method.GetClientStreaming() {
		reqArgs = append(reqArgs, contextPkg+".Context")
		ret = "(*" + x.typeName(method.GetOutputType()) + ", error)"
	}
	if !method.GetClientStreaming() {
		reqArgs = append(reqArgs, "*"+x.typeName(method.GetInputType()))
	}
	if method.GetServerStreaming() || method.GetClientStreaming() {
		reqArgs = append(reqArgs, servName+"_"+generator.CamelCase(origMethName)+"Server")
	}

	return methName + "(" + strings.Join(reqArgs, ", ") + ") " + ret
}

func (x *xrpc) generateServerMethod(servName, fullServName string, method *pb.MethodDescriptorProto) string {
	methName := generator.CamelCase(method.GetName())
	hname := fmt.Sprintf("_%s_%s_Handler", servName, methName)
	inType := x.typeName(method.GetInputType())
	outType := x.typeName(method.GetOutputType())

	if !method.GetServerStreaming() && !method.GetClientStreaming() {
		x.P("func ", hname, "(srv interface{}, ctx ", contextPkg, ".Context, dec func(interface{}) error, interceptor ", xrpcPkg, ".UnaryServerInterceptor) (interface{}, error) {")
		x.P("in := new(", inType, ")")
		x.P("if err := dec(in); err != nil { return nil, err }")
		x.P("if interceptor == nil { return srv.(", servName, "Server).", methName, "(ctx, in) }")
		x.P("info := &", xrpcPkg, ".UnaryServerInfo{")
		x.P("Server: srv,")
		x.P("FullMethod: ", strconv.Quote(fmt.Sprintf("/%s/%s", fullServName, methName)), ",")
		x.P("}")
		x.P("handler := func(ctx ", contextPkg, ".Context, req interface{}) (interface{}, error) {")
		x.P("return srv.(", servName, "Server).", methName, "(ctx, req.(*", inType, "))")
		x.P("}")
		x.P("return interceptor(ctx, in, info, handler)")
		x.P("}")
		x.P()
		return hname
	}
	streamType := unexport(servName) + methName + "Server"
	x.P("func ", hname, "(srv interface{}, stream ", xrpcPkg, ".ServerStream) error {")
	if !method.GetClientStreaming() {
		x.P("m := new(", inType, ")")
		x.P("if err := stream.RecvMsg(m); err != nil { return err }")
		x.P("return srv.(", servName, "Server).", methName, "(m, &", streamType, "{stream})")
	} else {
		x.P("return srv.(", servName, "Server).", methName, "(&", streamType, "{stream})")
	}
	x.P("}")
	x.P()

	genSend := method.GetServerStreaming()
	genSendAndClose := !method.GetServerStreaming()
	genRecv := method.GetClientStreaming()

	// Stream auxiliary types and methods.
	x.P("type ", servName, "_", methName, "Server interface {")
	if genSend {
		x.P("Send(*", outType, ") error")
	}
	if genSendAndClose {
		x.P("SendAndClose(*", outType, ") error")
	}
	if genRecv {
		x.P("Recv() (*", inType, ", error)")
	}
	x.P(xrpcPkg, ".ServerStream")
	x.P("}")
	x.P()

	x.P("type ", streamType, " struct {")
	x.P(xrpcPkg, ".ServerStream")
	x.P("}")
	x.P()

	if genSend {
		x.P("func (x *", streamType, ") Send(m *", outType, ") error {")
		x.P("return x.ServerStream.SendMsg(m)")
		x.P("}")
		x.P()
	}
	if genSendAndClose {
		x.P("func (x *", streamType, ") SendAndClose(m *", outType, ") error {")
		x.P("return x.ServerStream.SendMsg(m)")
		x.P("}")
		x.P()
	}
	if genRecv {
		x.P("func (x *", streamType, ") Recv() (*", inType, ", error) {")
		x.P("m := new(", inType, ")")
		x.P("if err := x.ServerStream.RecvMsg(m); err != nil { return nil, err }")
		x.P("return m, nil")
		x.P("}")
		x.P()
	}

	return hname
}
