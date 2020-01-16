package xrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/edenzhong7/xrpc/pkg/encoding"

	"github.com/edenzhong7/xrpc/middleware"
	"github.com/edenzhong7/xrpc/pkg/net"

	"github.com/xtaci/smux"
)

type ClientConn struct {
	dopts       *dialOptions
	protocol    net.Protocol
	session     *smux.Session
	streamCache map[string]ClientStream
	middlewares []middleware.ClientMiddleware
}

type CallOption struct {
}

func Dial(protocol net.Protocol, addr string, opts ...DialOption) (cc *ClientConn, err error) {
	conn, err := net.Dial(protocol, addr)
	session, err := smux.Client(conn, nil)
	n, err := conn.Write([]byte(Preface))
	if err != nil {
		return
	}
	if n != len(Preface) {
		return nil, errors.New("write Preface unmatch")
	}

	cc = &ClientConn{
		dopts:       &dialOptions{copts: ConnectOptions{dialer: net.GetDialer(protocol)}},
		protocol:    protocol,
		session:     session,
		streamCache: map[string]ClientStream{},
		middlewares: nil,
	}
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

func genStreamKey(protocol net.Protocol, addr string, method string) string {
	return fmt.Sprintf("%s://%s%s", protocol, addr, method)
}

func (cc *ClientConn) NewStream(ctx context.Context, desc *StreamDesc, method string, opts ...CallOption) (cs ClientStream, err error) {
	log.Println("new client session")
	var stream *smux.Stream
	var ok bool
	streamKey := genStreamKey(cc.protocol, cc.session.RemoteAddr().String(), method)
	if cs, ok = cc.streamCache[streamKey]; ok {
		return
	}

	stream, err = cc.session.OpenStream()
	header := &streamHeader{
		Cmd:        Init,
		FullMethod: method,
	}
	headerJson, err := json.Marshal(&header)
	if err != nil {
		return nil, err
	}
	hdr, data := msgHeader(headerJson, nil)
	hdr[0] = byte(CmdHeader)
	if _, err = stream.Write(hdr); err != nil {
		return nil, err
	}
	if _, err = stream.Write(data); err != nil {
		return nil, err
	}
	cs = &clientStream{
		stream: stream,
		header: header,
		codec:  encoding.GetCodec("proto"),
		cp:     encoding.GetCompressor("gzip"),
	}
	cc.streamCache[streamKey] = cs
	return cs, err
}

func (cc *ClientConn) Close() (err error) {
	err = cc.session.Close()
	return
}

func invoke(ctx context.Context, method string, req, reply interface{}, cc *ClientConn, opts ...CallOption) (err error) {
	cs, err := cc.NewStream(ctx, nil, method, opts...)
	if err != nil {
		return
	}
	log.Println("client invoke SendMsg")
	if err := cs.SendMsg(req); err != nil {
		return err
	}
	log.Println("client invoke RecvMsg")
	return cs.RecvMsg(reply)
}
