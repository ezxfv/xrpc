package net

import (
	"context"
	"errors"
	"net"
)

type Network = string
type Addr = net.Addr
type IPNet = net.IPNet

const (
	TCP  Network = "tcp"
	UDP          = "udp"
	UNIX         = "unix"
	QUIC         = "quic"
	WS           = "ws"
	KCP          = "kcp"
	SSH          = "ssh"
)

var (
	SplitHostPort = net.SplitHostPort
	ParseIP       = net.ParseIP
)

type XAddr struct {
	network Network
	addr    string
}

func (xaddr *XAddr) Network() string {
	return xaddr.network
}

func (xaddr *XAddr) String() string {
	return xaddr.addr
}

type Dialer func(ctx context.Context, addr string) (conn Conn, err error)
type ListenerBuilder func(ctx context.Context, addr string) (lis Listener, err error)

var (
	dialers   = make(map[Network]Dialer)
	listeners = make(map[Network]ListenerBuilder)
)

func RegisterDialer(protocol Network, dialer Dialer) {
	if _, ok := dialers[protocol]; !ok {
		dialers[protocol] = dialer
	}
}

func GetDialer(protocol Network) Dialer {
	dialer := dialers[protocol]
	return dialer
}

func RegisterListenerBuilder(protocol Network, builder ListenerBuilder) {
	if _, ok := listeners[protocol]; !ok {
		listeners[protocol] = builder
	}
}

type Listener interface {
	// Accept waits for and returns the next connection to the listener.
	Accept() (Conn, error)

	// Stop closes the listener.
	// Any blocked Accept operations will be unblocked and return errors.
	Close() error

	// Addr returns the listener's network address.
	Addr() Addr
}

type Conn = net.Conn

func Listen(ctx context.Context, protocol Network, addr string) (lis Listener, err error) {
	if builder, ok := listeners[protocol]; ok {
		lis, err = builder(ctx, addr)
		return
	}
	err = errors.New("unsupported network " + protocol)
	return
}

func Dial(ctx context.Context, protocol Network, addr string) (conn Conn, err error) {
	if dialer, ok := dialers[protocol]; ok {
		conn, err = dialer(ctx, addr)
		return
	}
	err = errors.New("unsupported network " + protocol)
	return
}
