package parser_test

import (
	"os"
	"testing"

	"github.com/edenzhong7/xrpc/pkg/generator/parser"
)

func TestParserMath(t *testing.T) {
	meta := parser.NewMetaData()
	meta.Parse("/Users/edenzhong/go/src/zen/xrpc/protocol/math/math.go")
	f, _ := os.Create("/Users/edenzhong/go/src/zen/xrpc/protocol/math/math.stub.go")
	meta.Print(parser.NewXrpcStubBuilder(), f)
}

func TestParserGreeter(t *testing.T) {
	meta := parser.NewMetaData()
	meta.Parse("/Users/edenzhong/go/src/zen/xrpc/protocol/math/greeter.go")
	f, _ := os.Create("/Users/edenzhong/go/src/zen/xrpc/protocol/math/greeter.stub.go")
	meta.Print(parser.NewXrpcStubBuilder(), f)
}
