package api_test

import (
	"testing"

	"github.com/edenzhong7/xrpc/api"
)

func TestServerAPI(t *testing.T) {
	api.Server(":8080")
}
