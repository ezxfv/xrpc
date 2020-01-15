package xrpc

import (
	"context"

	"github.com/edenzhong7/xrpc/middleware"
	"github.com/edenzhong7/xrpc/net"
)

type ClientConn struct {
	dopts *dialOptions

	middlewares []middleware.ClientMiddleware
}

type CallOption struct {
}

func Dial(protocol net.Protocol, addr string, opts ...DialOption) (cc *ClientConn, err error) {
	return
}

func (cc *ClientConn) Invoke(ctx context.Context, method string, args, reply interface{}, opts ...CallOption) error {
	handler := func() error {
		return invoke(ctx, method, args, reply, cc, opts...)
	}
	for _, m := range cc.middlewares {
		handler = m.Handle(ctx, handler).(func() error)
	}
	return handler()
}

func (cc *ClientConn) AddMiddleware(ms ...middleware.ClientMiddleware) {
	cc.middlewares = append(cc.middlewares, ms...)
}

func (cc *ClientConn) NewStream(ctx context.Context, desc *StreamDesc, method string, opts ...CallOption) (cs ClientStream, err error) {
	return
}

func (cc *ClientConn) Close() (err error) {
	return
}

func invoke(ctx context.Context, method string, args, reply interface{}, cc *ClientConn, opts ...CallOption) (e error) {
	return nil
}
