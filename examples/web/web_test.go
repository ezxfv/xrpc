package web_test

import (
	"net/http"
	"testing"

	"x.io/xrpc/pkg/log"

	"github.com/lucas-clemente/quic-go/http3"
)

func TestQuic(t *testing.T) {
	http.Handle("/", http.FileServer(http.Dir("/home/edenz/Pictures")))
	log.Fatal(http3.ListenAndServeQUIC("localhost:4242", "../../testdata/server.crt", "../../testdata/server.key", nil))
}
