package xrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"x.io/xrpc/pkg/encoding"
	"x.io/xrpc/pkg/net"
	"x.io/xrpc/plugin"
	_ "x.io/xrpc/plugin/chord"
	"x.io/xrpc/types"

	"github.com/xtaci/smux"
)

type ClientConn struct {
	dopts       *dialOptions
	protocol    net.Network
	session     *smux.Session
	conn        net.Conn
	streamCache map[string]types.ClientStream

	args map[string]interface{}
	pioc plugin.Container
}

type CallOption struct {
}

func Dial(network net.Network, addr string, opts ...DialOption) (cc *ClientConn, err error) {
	conn, err := net.Dial(context.Background(), network, addr)
	if err != nil {
		return
	}
	session, err := smux.Client(conn, nil)
	if err != nil {
		return
	}
	n, err := conn.Write([]byte(types.Preface))
	if err != nil {
		return
	}
	if n != len(types.Preface) {
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
		streamCache: map[string]types.ClientStream{},
		args:        map[string]interface{}{},
		pioc:        plugin.NewPluginContainer(),
	}
	return
}

func NewRawClient(protocol, addr string, opts ...DialOption) (*RawClient, error) {
	conn, err := Dial(protocol, addr, opts...)
	if err != nil {
		return nil, err
	}
	return &RawClient{cc: conn}, nil
}

type RawClient struct {
	cc *ClientConn
}

func (rc *RawClient) Setup(s func(cc *ClientConn)) {
	s(rc.cc)
}

func (rc *RawClient) RawCall(ctx context.Context, method string, reply interface{}, args ...interface{}) (err error) {
	ctx = context.WithValue(ctx, "codec", "json")
	arr := strings.Split(method, ".")
	if len(arr) != 2 {
		return errors.New("method name should be: xxx.xxx")
	}
	fullMethod := fmt.Sprintf("/%s.%s/%s", customPrefix, arr[0], arr[1])
	cs, err := rc.cc.NewStream(ctx, types.RawRPC, nil, fullMethod)
	if err != nil {
		return
	}
	for k, v := range rc.cc.args {
		if vv, ok := v.(string); ok {
			ctx = types.SetCookie(ctx, k, vv)
		}
	}
	if err = cs.SendMsg(ctx, &args); err != nil {
		return
	}
	ctx, err = cs.RecvMsg(ctx, reply)
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

func (cc *ClientConn) NewStream(ctx context.Context, rpc types.Rpc, desc *types.StreamDesc, method string, opts ...CallOption) (cs types.ClientStream, err error) {
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

	codec := cc.dopts.codec
	compressor := cc.dopts.compressor
	if ctx.Value("codec") != nil {
		if v, ok := ctx.Value("codec").(string); ok {
			codec = v
		}
	}
	if ctx.Value("compressor") != nil {
		if v, ok := ctx.Value("compressor").(string); ok {
			compressor = v
		}
	}
	args := map[string]interface{}{
		"codec":      codec,
		"compressor": compressor,
	}
	for k, v := range cc.args {
		args[k] = v
	}
	header := &types.streamHeader{
		Cmd:        types.Init,
		FullMethod: method,
		RpcType:    rpc,
		Args:       args,
	}
	headerJson, err := json.Marshal(&header)
	if err != nil {
		return nil, err
	}
	hdr := types.msgHeader(headerJson, false)
	hdr[0] = byte(types.cmdHeader)
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

func (cc *ClientConn) Network() string {
	return cc.conn.RemoteAddr().Network()
}

func (cc *ClientConn) Addr() string {
	return cc.conn.RemoteAddr().String()
}

func invoke(ctx context.Context, method string, req, reply interface{}, cc *ClientConn, opts ...CallOption) (err error) {
	cs, err := cc.NewStream(ctx, types.XRPC, nil, method, opts...)
	if err != nil {
		return
	}
	for k, v := range cc.args {
		if vv, ok := v.(string); ok {
			ctx = types.SetCookie(ctx, k, vv)
		}
	}
	if err := cs.SendMsg(ctx, req); err != nil {
		return err
	}
	ctx, err = cs.RecvMsg(ctx, reply)
	return
}
