package net_test

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"x.io/xrpc/pkg/net"

	"github.com/stretchr/testify/assert"
	"github.com/xtaci/smux"
)

var (
	req  = []byte("hello")
	resp = []byte("world!")

	addr = "localhost:9898"

	smuxCfg = &smux.Config{
		Version:           2,
		KeepAliveInterval: 10 * time.Second,
		KeepAliveTimeout:  30 * time.Second,
		MaxFrameSize:      32768,
		MaxReceiveBuffer:  4194304,
		MaxStreamBuffer:   65536,
	}
)

func testLisConn(t *testing.T, network net.Network) {
	lis, err := net.Listen(context.Background(), network, addr)
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

func testConn(t *testing.T, network net.Network) {
	conn, err := net.Dial(context.Background(), network, addr)
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

func TestUDPListener(t *testing.T) {
	testLisConn(t, net.UDP)
}

func TestUDPConn(t *testing.T) {
	testConn(t, net.UDP)
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
	c, err := lis.Accept()
	assert.Equal(t, nil, err)

	session, err := smux.Server(c, nil)
	assert.Equal(t, nil, err)
	conn, err := session.AcceptStream()
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

func TestWsMuxConn(t *testing.T) {
	c, err := net.Dial(context.Background(), net.WS, addr)
	assert.Equal(t, nil, err)

	session, err := smux.Client(c, nil)
	assert.Equal(t, nil, err)
	conn, err := session.OpenStream()
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

func TestQuicListener(t *testing.T) {
	testLisConn(t, net.QUIC)
}

func TestQuicConn(t *testing.T) {
	testConn(t, net.QUIC)
}

func TestReusePort(t *testing.T) {
	_, err := net.Listen(context.Background(), net.WS, addr)
	assert.Equal(t, nil, err)
	_, err = net.Listen(context.Background(), net.KCP, addr)
	assert.Equal(t, nil, err)
	_, err = net.Listen(context.Background(), net.TCP, addr)
	assert.Equal(t, nil, err)
	_, err = net.Listen(context.Background(), net.QUIC, addr)
	assert.Equal(t, nil, err)
}

func TestUDP2(t *testing.T) {
	go func() {
		lis, err := net.Listen(context.Background(), "udp", addr)
		assert.Equal(t, nil, err)

		for {
			conn, err := lis.Accept()
			if err != nil {
				log.Fatalln(err.Error())
			}
			go func() {
				for {
					buf := make([]byte, 11, 11)
					n, err := conn.Read(buf)
					if err != nil {
						log.Fatalln(err.Error())
					}
					fmt.Printf("recv %s\n", string(buf[:n]))
					time.Sleep(time.Second)
				}
			}()
		}
	}()
	time.Sleep(time.Second)
	conn1, err := net.Dial(nil, "udp", addr)
	if err != nil {
		log.Fatalln(err.Error())
	}
	conn2, err := net.Dial(nil, "udp", addr)
	if err != nil {
		log.Fatalln(err.Error())
	}
	wg := &sync.WaitGroup{}
	wg.Add(2)
	f := func(p string, conn net.Conn) {
		for i := 0; i < 10; i++ {
			conn.Write([]byte(p + " hello"))
			time.Sleep(time.Second)
		}
		wg.Done()
	}
	go f("user1", conn1)
	go f("user2", conn2)
	wg.Wait()
}
