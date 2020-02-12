package client_test

import (
	"testing"

	"x.io/xrpc/app/chord/client"

	"github.com/stretchr/testify/assert"
)

func TestNewChordClient(t *testing.T) {
	cc := client.NewChordClient(client.DefaultURL)
	err := cc.Set("xrpc", "dev")
	assert.Equal(t, nil, err)
	v, err := cc.Get("xrpc")
	assert.Equal(t, nil, err)
	assert.Equal(t, "dev", v)
	err = cc.Del("xrpc")
	assert.Equal(t, nil, err)
}
