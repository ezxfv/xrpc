package parser_test

import (
	"os"
	"testing"

	"x.io/xrpc/pkg/generator/parser"
)

func TestRpcStub(t *testing.T) {
	meta := parser.NewMetaData()
	meta.Parse("/Users/edenzhong/go/src/zen/xrpc/protocol/math/math.go")
	f, _ := os.Create("/Users/edenzhong/go/src/zen/xrpc/protocol/math/math.rpc.go")
	parser.RpcStub(meta, parser.NewXrpcStubBuilder(), f)
}

func TestHttpStub(t *testing.T) {
	meta := parser.NewMetaData()
	meta.Parse("/Users/edenzhong/go/src/zen/xrpc/protocol/imdb/imdb.go")
	f, _ := os.Create("/Users/edenzhong/go/src/zen/xrpc/protocol/imdb/imdb.http.go")
	parser.HttpStub(meta, parser.NewHttpStubBuilder(), f)
}
