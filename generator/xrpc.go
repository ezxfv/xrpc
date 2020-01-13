package generator

import (
	"github.com/golang/protobuf/protoc-gen-go/generator"
	_ "github.com/jhump/protoreflect/grpcreflect"
)

func init() {
	generator.RegisterPlugin(new(xrpcPlugin))
}

type xrpcPlugin struct {
	gen *generator.Generator
}

func (x *xrpcPlugin) Name() string {
	return "xrpc"
}

func (x *xrpcPlugin) Init(g *generator.Generator) {
	x.gen = g
}

func (x *xrpcPlugin) Generate(file *generator.FileDescriptor) {
	panic("implement me")
}

func (x *xrpcPlugin) GenerateImports(file *generator.FileDescriptor) {
	panic("implement me")
}
