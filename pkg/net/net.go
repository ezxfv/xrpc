package net

import (
	"context"
	"errors"
	"net"
)

type Protocol = string
type Addr = net.Addr

type XAddr struct {
	protocol Protocol
	addr     string
}

func (xaddr *XAddr) Network() string {
	return xaddr.protocol
}

func (xaddr *XAddr) String() string {
	return xaddr.addr
}

const (
	TCP  Protocol = "tcp"
	UDP           = "udp"
	QUIC          = "quic"
	WS            = "ws"
	KCP           = "kcp"
)

type Dialer func(ctx context.Context, addr string) (conn Conn, err error)
type ListenerBuilder func(ctx context.Context, addr string) (lis Listener, err error)

var (
	dialers   = make(map[Protocol]Dialer)
	listeners = make(map[Protocol]ListenerBuilder)
)

func RegisterDialer(protocol Protocol, dialer Dialer) {
	if _, ok := dialers[protocol]; !ok {
		dialers[protocol] = dialer
	}
}

func GetDialer(protocol Protocol) Dialer {
	dialer := dialers[protocol]
	return dialer
}

func RegisterListenerBuilder(protocol Protocol, builder ListenerBuilder) {
	if _, ok := listeners[protocol]; !ok {
		listeners[protocol] = builder
	}
}

type Listener interface {
	// Accept waits for and returns the next connection to the listener.
	Accept() (Conn, error)

	// Close closes the listener.
	// Any blocked Accept operations will be unblocked and return errors.
	Close() error

	// Addr returns the listener's network address.
	Addr() Addr
}

type Conn interface {
	net.Conn
	SupportMux() bool
}

func Listen(ctx context.Context, protocol Protocol, addr string) (lis Listener, err error) {
	if builder, ok := listeners[protocol]; ok {
		lis, err = builder(ctx, addr)
		return
	}
	err = errors.New("unsupported protocol " + protocol)
	return
}

func Dial(ctx context.Context, protocol Protocol, addr string) (conn Conn, err error) {
	if dialer, ok := dialers[protocol]; ok {
		conn, err = dialer(ctx, addr)
		return
	}
	err = errors.New("unsupported protocol " + protocol)
	return
}
