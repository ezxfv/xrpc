package net

import (
	"context"
	"crypto/sha1"
	"time"

	kcp "github.com/xtaci/kcp-go"

	"golang.org/x/crypto/pbkdf2"
)

var (
	kcpDialer Dialer = func(ctx context.Context, addr string) (conn Conn, err error) {
		key := pbkdf2.Key([]byte("demo pass"), []byte("demo salt"), 1024, 32, sha1.New)
		block, _ := kcp.NewAESBlockCrypt(key)

		// wait for server to become ready
		time.Sleep(time.Second)

		// dial to the echo server
		conn, err = kcp.DialWithOptions(addr, block, 10, 3)
		return
	}
)

func init() {
	RegisterDialer(KCP, kcpDialer)
	RegisterListenerBuilder(KCP, newKcpListener)
}

func newKcpListener(ctx context.Context, addr string) (lis Listener, err error) {
	key := pbkdf2.Key([]byte("demo pass"), []byte("demo salt"), 1024, 32, sha1.New)
	block, _ := kcp.NewAESBlockCrypt(key)
	lis, err = kcp.ListenWithOptions(addr, block, 10, 3)
	return
}
