package xrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/edenzhong7/xrpc/plugin"

	"github.com/edenzhong7/xrpc/pkg/encoding"
	"github.com/edenzhong7/xrpc/pkg/net"
	"github.com/xtaci/smux"
)

type ClientConn struct {
	dopts       *dialOptions
	protocol    net.Network
	session     *smux.Session
	conn        net.Conn
	streamCache map[string]ClientStream

	args map[string]interface{}
	pioc plugin.Container
}

type CallOption struct {
}

func Dial(network net.Network, addr string, opts ...DialOption) (cc *ClientConn, err error) {
	conn, err := net.Dial(context.Background(), network, addr)
	session, err := smux.Client(conn, nil)
	n, err := conn.Write([]byte(Preface))
	if err != nil {
		return
	}
	if n != len(Preface) {
		return nil, errors.New("wrote Preface length isn't match")
	}
	dopts := &dialOptions{
		copts:      ConnectOptions{dialer: net.GetDialer(network)},
		codec:      "proto",
		compressor: "gzip",
	}
	for _, opt := range opts {
		opt.apply(dopts)
	}
	cc = &ClientConn{
		dopts:       dopts,
		protocol:    network,
		session:     session,
		conn:        conn,
		streamCache: map[string]ClientStream{},
		args:        map[string]interface{}{},
		pioc:        plugin.NewPluginContainer(),
	}
	return
}

func (cc *ClientConn) ApplyPlugins(plugins ...plugin.Plugin) {
	for _, pp := range plugins {
		cc.pioc.Add(pp)
	}
}
func (cc *ClientConn) SetHeaderArg(key string, value interface{}) {
	cc.args[key] = value
}

func (cc *ClientConn) Invoke(ctx context.Context, method string, args, reply interface{}, opts ...CallOption) error {
	handler := func() error {
		return invoke(ctx, method, args, reply, cc, opts...)
	}
	return handler()
}

func genStreamKey(network net.Network, addr string, method string) string {
	return fmt.Sprintf("%s://%s%s", network, addr, method)
}

func (cc *ClientConn) NewStream(ctx context.Context, desc *StreamDesc, method string, opts ...CallOption) (cs ClientStream, err error) {
	var stream net.Conn
	var ok bool
	streamKey := genStreamKey(cc.protocol, cc.session.RemoteAddr().String(), method)
	if cs, ok = cc.streamCache[streamKey]; ok {
		return
	}

	s, err := cc.session.OpenStream()
	if err != nil {
		return nil, err
	}
	stream = &streamConn{s}

	if err != nil {
		return nil, err
	}
	args := map[string]interface{}{
		"codec":      cc.dopts.codec,
		"compressor": cc.dopts.compressor,
	}
	for k, v := range cc.args {
		args[k] = v
	}
	header := &streamHeader{
		Cmd:        Init,
		FullMethod: method,
		Args:       args,
	}
	headerJson, err := json.Marshal(&header)
	if err != nil {
		return nil, err
	}
	hdr := msgHeader(headerJson, false)
	hdr[0] = byte(cmdHeader)
	if _, err = stream.Write(hdr); err != nil {
		return nil, err
	}
	if _, err = stream.Write(headerJson); err != nil {
		return nil, err
	}
	cs = &clientStream{
		stream: stream,
		header: header,
		codec:  encoding.GetCodec(cc.dopts.codec),
		cp:     encoding.GetCompressor(cc.dopts.compressor),
		pioc:   cc.pioc,
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
	for k, v := range cc.args {
		if vv, ok := v.(string); ok {
			ctx = SetCookie(ctx, k, vv)
		}
	}
	if err := cs.SendMsg(ctx, req); err != nil {
		return err
	}
	ctx, err = cs.RecvMsg(ctx, reply)
	return
}
