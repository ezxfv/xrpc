package net

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

var (
	udpInit = [8]byte{'1'}
)

var (
	udpDialer Dialer = func(ctx context.Context, addr string) (conn Conn, err error) {
		uaddr, err := ResolveUDPAddr("udp", addr)
		c, err := net.DialUDP("udp", nil, uaddr)
		if err != nil {
			return nil, err
		}
		n, err := c.Write(udpInit[:])
		if n != len(udpInit) {
			return nil, errors.New("write udp init header failed")
		}
		if err != nil {
			return nil, err
		}
		quit := make(chan struct{})
		conn = newUdpConn(true, c.LocalAddr(), uaddr, DefaultTimeout, nil, nil, c, quit)
		return
	}
)

func init() {
	RegisterDialer(UDP, udpDialer)
	RegisterListenerBuilder(UDP, newUDPListener)
}

func newUDPListener(ctx context.Context, addr string) (lis Listener, err error) {
	l, err := UDPListen("udp", addr)
	ul := &udpListener{
		lis:     l,
		buffers: map[string]chan<- []byte{},
		qs:      map[string]chan struct{}{},
		conns:   make(chan *udpConn, 128),
	}
	go ul.loop()
	return ul, err
}

func newUdpConn(client bool, laddr, raddr net.Addr, timeout time.Duration, r <-chan []byte, w net.PacketConn, cw *net.UDPConn, quit chan struct{}) *udpConn {
	conn := &udpConn{
		client:     client,
		localAddr:  laddr,
		remoteAddr: raddr,
		r:          r,
		w:          w,
		cw:         cw,
		mu:         &sync.Mutex{},
		timeout:    timeout,
		quit:       quit,
	}
	return conn
}

type udpConn struct {
	client     bool
	r          <-chan []byte
	buf        []byte
	w          net.PacketConn
	cw         *net.UDPConn
	mu         *sync.Mutex
	remoteAddr net.Addr
	localAddr  net.Addr

	readDeadline  time.Time
	writeDeadline time.Time
	timeout       time.Duration
	quit          chan struct{}
}

func (u *udpConn) loop() {
	for {
		select {
		case buf := <-u.r:
			u.mu.Lock()
			u.buf = append(u.buf, buf...)
			u.mu.Unlock()
		case <-u.quit:
			return
		}
	}
}

func (u *udpConn) Read(b []byte) (n int, err error) {
	l := len(b)
	if u.client {
		n, err = u.cw.Read(b)
		return
	}
	now := time.Now()
	for {
		if len(u.buf) >= l {
			break
		}
		if time.Since(now) > u.timeout {
			break
		}
		time.Sleep(time.Nanosecond * 10)
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	if l <= len(u.buf) {
		copy(b, u.buf[:l])
		b = b[l:]
		n = l
		return
	}
	n = len(u.buf)
	copy(b, u.buf)
	u.buf = nil
	return
}

func (u *udpConn) Write(b []byte) (n int, err error) {
	if !u.client {
		return u.w.WriteTo(b, u.remoteAddr)
	}
	buf := make([]byte, 8, 8)
	for i := 0; i < len(b); i += 7 {
		if i+7 <= len(b) {
			copy(buf[1:], b[i:i+7])
			_, err = u.cw.Write(buf)
			if err != nil {
				return 0, err
			}
			n += 7
		} else {
			copy(buf[1:], b[i:])
			buf[0] = string(len(b) - i)[0]
			_, err = u.cw.Write(buf)
			if err != nil {
				return 0, err
			}
			n += len(b) - i
		}
	}
	return n, err
}

func (u *udpConn) Close() (err error) {
	close(u.quit)
	if u.client {
		err = u.cw.Close()
	}
	return err
}

func (u *udpConn) LocalAddr() net.Addr {
	if u.client {
		return u.cw.LocalAddr()
	}
	return u.w.LocalAddr()
}

func (u *udpConn) RemoteAddr() net.Addr {
	return u.remoteAddr
}

func (u *udpConn) SetDeadline(t time.Time) error {
	if u.client {
		return u.cw.SetDeadline(t)
	}
	u.readDeadline = t
	u.writeDeadline = t
	return nil
}

func (u *udpConn) SetReadDeadline(t time.Time) error {
	if u.client {
		return u.cw.SetReadDeadline(t)
	}
	u.readDeadline = t
	return nil
}

func (u *udpConn) SetWriteDeadline(t time.Time) error {
	if u.client {
		return u.cw.SetWriteDeadline(t)
	}
	u.writeDeadline = t
	return nil
}

type udpListener struct {
	lis     PacketConn
	buffers map[string]chan<- []byte
	qs      map[string]chan struct{}
	conns   chan *udpConn

	mu *sync.Mutex
}

func (ul *udpListener) loop() {
	for {
		buf := make([]byte, 8, 8)
		n, addr, err := ul.lis.ReadFrom(buf)
		if err != nil {
			break
		}
		if n != len(buf) {
			break
		}
		if buf[0] == byte('1') {
			// new conn
			raddr := addr.String()
			println("conn ", raddr)
			recv := make(chan []byte, 1024)
			cquit := make(chan struct{})
			ul.buffers[raddr] = recv
			ul.qs[raddr] = cquit
			conn := newUdpConn(false, ul.lis.LocalAddr(), addr, DefaultTimeout, recv, ul.lis, nil, cquit)
			go conn.loop()
			ul.conns <- conn
			continue
		}
		// get data
		bufLen := int(buf[0])
		raddr := addr.String()
		r, ok := ul.buffers[raddr]
		if ok {
			var data []byte
			if bufLen == 0 {
				data = buf[1:]
			} else {
				data = buf[1 : bufLen+1]
			}
			r <- data
		}
	}
	return
}

func (ul *udpListener) Accept() (conn Conn, err error) {
	c, ok := <-ul.conns
	if !ok {
		return nil, errors.New("udp listener closed")
	}
	return c, nil
}

func (ul *udpListener) Close() error {
	//for _, q := range ul.qs {
	//	close(q)
	//}
	//for _, b := range ul.buffers {
	//	close(b)
	//}
	return ul.lis.Close()
}

func (ul *udpListener) Addr() Addr {
	return ul.lis.LocalAddr()
}
