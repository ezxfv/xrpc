package client_test

import (
	"strconv"
	"testing"

	"x.io/xrpc/app/chord/client"

	"github.com/stretchr/testify/assert"
)

func TestNewChordClient(t *testing.T) {
	cc := client.NewChordClient("http://localhost:9900/chord")
	err := cc.Set("xrpc", "dev")
	assert.Equal(t, nil, err)
	v, err := cc.Get("xrpc")
	assert.Equal(t, nil, err)
	assert.Equal(t, "dev", v)
	err = cc.Del("xrpc")
	assert.Equal(t, nil, err)
}

func TestChordSet(t *testing.T) {
	cc := client.NewChordClient("http://localhost:9900/chord")
	for i := 0; i < 10; i++ {
		a := strconv.Itoa(i)
		err := cc.Set("key-"+a, "val-"+a)
		assert.Equal(t, nil, err)
	}
	for i := 0; i < 10; i++ {
		a := strconv.Itoa(i)
		val, err := cc.Get("key-" + a)
		assert.Equal(t, nil, err)
		assert.Equal(t, "val-"+a, val)
	}
}
