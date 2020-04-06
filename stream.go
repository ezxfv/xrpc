package xrpc

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"x.io/xrpc/pkg/encoding"
	"x.io/xrpc/pkg/net"
	"x.io/xrpc/pkg/transport"
	"x.io/xrpc/plugin"
	"x.io/xrpc/types"

	"github.com/xtaci/smux"
)

type clientStream struct {
	ctx context.Context

	stream net.Conn
	t      transport.Transport
	header *types.StreamHeader
	codec  encoding.Codec
	cp     encoding.Compressor

	pioc plugin.IOContainer
}

func (cs *clientStream) Close() error {
	return cs.stream.Close()
}

func (cs *clientStream) Context() context.Context {
	return cs.ctx
}

func (cs *clientStream) SendMsg(ctx context.Context, m interface{}) error {
	if cs.t != nil {
		// 通过transporter发送
		return cs.t.SendMsg(ctx, m)
	}

	data, err := cs.codec.Marshal(m)
	if err != nil {
		return err
	}
	cookies := types.CookiesHeader(ctx)
	data = append(cookies, data...)

	var compData []byte = nil
	cbuf := &bytes.Buffer{}
	z, err := cs.cp.Compress(cbuf)
	if z != nil {
		if _, err = z.Write(data); err != nil {
			return err
		}
		if err = z.Close(); err != nil {
			return err
		}
	}
	compData = cbuf.Bytes()
	comp := false
	if len(compData) > 0 {
		data = compData
		comp = true
	}
	if data, err = cs.pioc.DoPreWriteResponse(ctx, data); err != nil {
		return err
	}
	hdr := types.MsgHeader(data, comp)

	if _, err = cs.stream.Write(hdr); err != nil {
		return err
	}
	if _, err = cs.stream.Write(data); err != nil {
		return err
	}
	return nil
}

func recv(conn io.Reader) (pf types.PayloadFormat, msg []byte, err error) {
	header := make([]byte, types.HeaderLen, types.HeaderLen)
	if _, err := conn.Read(header[:]); err != nil {
		return 0, nil, err
	}

	pf = types.PayloadFormat(header[0])
	length := binary.BigEndian.Uint32(header[1:])

	if length == 0 {
		return pf, nil, nil
	}
	msg = make([]byte, int(length))
	if _, err := conn.Read(msg); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return 0, nil, err
	}
	return pf, msg, nil
}

func (cs *clientStream) RecvMsg(ctx context.Context, m interface{}) (context.Context, error) {
	if cs.t != nil {
		// 通过transporter接收
		return cs.t.RecvMsg(ctx, m)
	}
	pf, msg, err := recv(cs.stream)
	if err != nil {
		return ctx, err
	}
	if msg, err = cs.pioc.DoPreReadRequest(ctx, msg); err != nil {
		return ctx, err
	}
	var data []byte
	if pf == types.CompressionMade {
		dc, _ := cs.cp.Decompress(bytes.NewReader(msg))
		data, err = ioutil.ReadAll(dc)
		if err != nil {
			return ctx, err
		}
	} else {
		data = msg
	}
	ctx, l := types.ReadCookiesHeader(ctx, data)
	if err = cs.codec.Unmarshal(data[l:], m); err != nil {
		return ctx, errors.New(fmt.Sprintf("xrpc: failed to unmarshal the received message %v", err))
	}
	return ctx, nil
}

func (cs *clientStream) Header() (types.MD, error) {
	panic("implement me")
}

func (cs *clientStream) Trailer() types.MD {
	panic("implement me")
}

func (cs *clientStream) CloseSend() error {
	panic("implement me")
}

type streamConn struct {
	*smux.Stream
}

func (sc *streamConn) SupportMux() bool {
	return false
}

type serverStream struct {
	ctx context.Context

	stream net.Conn
	t      transport.Transport
	header *types.StreamHeader

	codec encoding.Codec
	cp    encoding.Compressor

	sc plugin.IOContainer
}

func (ss *serverStream) Close() error {
	return ss.stream.Close()
}

func (ss *serverStream) Context() context.Context {
	return ss.ctx
}

func (ss *serverStream) SendMsg(ctx context.Context, m interface{}) (err error) {
	defer func() {
		// DoPostWriteResponse
		err = ss.sc.DoPostWriteResponse(ctx, nil, m, err)
	}()
	if ss.t != nil {
		// 通过transporter发送
		return ss.t.SendMsg(ctx, m)
	}
	data, err := ss.codec.Marshal(m)
	if err != nil {
		return err
	}
	cookies := types.CookiesHeader(ctx)
	data = append(cookies, data...)
	var compData []byte = nil
	cbuf := &bytes.Buffer{}
	z, err := ss.cp.Compress(cbuf)
	if z != nil {
		if _, err = z.Write(data); err != nil {
			return err
		}
		if err = z.Close(); err != nil {
			return err
		}
	}
	compData = cbuf.Bytes()
	comp := false
	if len(compData) != 0 {
		data = compData
		comp = true
	}
	// DoPreWriteResponse
	if data, err = ss.sc.DoPreWriteResponse(ctx, data); err != nil {
		return err
	}
	hdr := types.MsgHeader(data, comp)
	if _, err = ss.stream.Write(hdr); err != nil {
		return err
	}
	if _, err = ss.stream.Write(data); err != nil {
		return err
	}
	return err
}

func (ss *serverStream) RecvMsg(ctx context.Context, m interface{}) (context.Context, error) {
	if ss.t != nil {
		// 通过transporter接收
		return ss.t.RecvMsg(ctx, m)
	}
	pf, msg, err := recv(ss.stream)
	if err != nil {
		return ctx, err
	}
	// DoPreReadRequest
	if msg, err = ss.sc.DoPreReadRequest(ctx, msg); err != nil {
		return ctx, err
	}
	var data []byte
	if pf == types.CompressionMade {
		dc, _ := ss.cp.Decompress(bytes.NewReader(msg))
		if dc == nil {
			return ctx, errors.New("decompress failed")
		}
		data, err = ioutil.ReadAll(dc)
		if err != nil {
			return ctx, err
		}
	} else {
		data = msg
	}
	ctx, l := types.ReadCookiesHeader(ctx, data)
	if err = ss.codec.Unmarshal(data[l:], m); err != nil {
		err = errors.New(fmt.Sprintf("xrpc: failed to unmarshal the received message for %v", err))
	}
	// DoPostReadRequest
	err = ss.sc.DoPostReadRequest(ctx, m, err)
	return ctx, err
}

func (ss *serverStream) SetHeader(types.MD) error {
	panic("implement me")
}

func (ss *serverStream) SendHeader(types.MD) error {
	panic("implement me")
}

func (ss *serverStream) SetTrailer(types.MD) {
	panic("implement me")
}
