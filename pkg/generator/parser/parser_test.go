package parser_test

import (
	"os"
	"testing"

	"github.com/edenzhong7/xrpc/pkg/generator/parser"
)

func TestParserMath(t *testing.T) {
	meta := parser.NewMetaData()
	meta.Parse("/Users/edenzhong/go/src/zen/xrpc/protocol/math/math.go")
	f, _ := os.Create("/Users/edenzhong/go/src/zen/xrpc/protocol/math/math_stub.go")
	meta.Print(parser.NewXrpcStubBuilder(), f)
	//printer.Fprint(os.Stdout, fs, f)
}

func TestParserGreeter(t *testing.T) {
	meta := parser.NewMetaData()
	meta.Parse("/Users/edenzhong/go/src/zen/xrpc/protocol/math/greeter.go")
	meta.Print(parser.NewXrpcStubBuilder(), os.Stdout)
	//printer.Fprint(os.Stdout, fs, f)
}
