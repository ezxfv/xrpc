package net_test

import (
	"context"
	"testing"

	"github.com/xtaci/smux"

	"github.com/edenzhong7/xrpc/pkg/net"

	"github.com/stretchr/testify/assert"
)

var (
	req  = []byte("hello")
	resp = []byte("world!")

	addr = "localhost:9898"
)

func testLisConn(t *testing.T, protocol net.Protocol) {
	lis, err := net.Listen(context.Background(), protocol, addr)
	assert.Equal(t, nil, err)
	conn, err := lis.Accept()
	assert.Equal(t, nil, err)
	d := make([]byte, len(req))
	n, err := conn.Read(d)
	assert.Equal(t, len(req), n)
	assert.Equal(t, req, d)
	n, err = conn.Write(resp)
	assert.Equal(t, nil, err)
	assert.Equal(t, len(resp), n)
	assert.Equal(t, nil, conn.Close())
	assert.Equal(t, nil, lis.Close())
}

func testConn(t *testing.T, protocol net.Protocol) {
	conn, err := net.Dial(context.Background(), protocol, addr)
	assert.Equal(t, nil, err)
	n, err := conn.Write(req)
	assert.Equal(t, nil, err)
	assert.Equal(t, len(req), n)

	d := make([]byte, len(resp))
	n, err = conn.Read(d)
	assert.Equal(t, len(resp), n)
	assert.Equal(t, resp, d)
	assert.Equal(t, nil, conn.Close())
}

func TestKCPListener(t *testing.T) {
	testLisConn(t, net.KCP)
}

func TestKCPConn(t *testing.T) {
	testConn(t, net.KCP)
}

func TestWSListener(t *testing.T) {
	testLisConn(t, net.WS)
}

func TestWSConn(t *testing.T) {
	testConn(t, net.WS)
}

func TestWsMuxSession(t *testing.T) {
	lis, err := net.Listen(context.Background(), net.WS, addr)
	assert.Equal(t, nil, err)
	conn, err := lis.Accept()
	assert.Equal(t, nil, err)
	_, err = smux.Server(conn, nil)
	assert.Equal(t, nil, err)
}

func TestWsMuxConn(t *testing.T) {

}

func TestQuicListener(t *testing.T) {
	testLisConn(t, net.QUIC)
}

func TestQuicConn(t *testing.T) {
	testConn(t, net.QUIC)
}
