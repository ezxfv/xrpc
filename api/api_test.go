package api_test

import (
	"testing"

	"x.io/xrpc/api"
)

func TestServerAPI(t *testing.T) {
	api.Server(":8080")
}
