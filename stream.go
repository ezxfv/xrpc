package xrpc

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/edenzhong7/xrpc/pkg/encoding"
	"github.com/edenzhong7/xrpc/pkg/net"
	"github.com/edenzhong7/xrpc/plugin"

	"github.com/xtaci/smux"
	"google.golang.org/grpc/metadata"
)

// Stream defines the common interface a client or server stream has to satisfy.
type Stream interface {
	// Context returns the context for this stream.
	Context() context.Context
	// SendMsg blocks until it sends m, the stream is done or the stream
	// breaks.
	// On error, it aborts the stream and returns an RPC status on client
	// side. On server side, it simply returns the error to the caller.
	// SendMsg is called by generated code. Also Users can call SendMsg
	// directly when it is really needed in their use cases.
	// It's safe to have a goroutine calling SendMsg and another goroutine calling
	// recvMsg on the same stream at the same time.
	// But it is not safe to call SendMsg on the same stream in different goroutines.
	SendMsg(ctx context.Context, m interface{}) error
	// RecvMsg blocks until it receives a message or the stream is
	// done. On client side, it returns io.EOF when the stream is done. On
	// any other error, it aborts the stream and returns an RPC status. On
	// server side, it simply returns the error to the caller.
	// It's safe to have a goroutine calling SendMsg and another goroutine calling
	// recvMsg on the same stream at the same time.
	// But it is not safe to call RecvMsg on the same stream in different goroutines.
	RecvMsg(ctx context.Context, m interface{}) (context.Context, error)

	Close() error
}

type ClientStream interface {
	// Stream.SendMsg() may return a non-nil error when something wrong happens sending
	// the request. The returned error indicates the status of this sending, not the final
	// status of the RPC.
	// Always call Stream.RecvMsg() to get the final status if you care about the status of
	// the RPC.
	Stream
	// Header returns the header metadata received from the server if there
	// is any. It blocks if the metadata is not ready to read.
	Header() (metadata.MD, error)
	// Trailer returns the trailer metadata from the server, if there is any.
	// It must only be called after stream.CloseAndRecv has returned, or
	// stream.Recv has returned a non-nil error (including io.EOF).
	Trailer() metadata.MD
	// CloseSend closes the send direction of the stream. It closes the stream
	// when non-nil error is met.
	CloseSend() error
}

type ServerStream interface {
	Stream
	// SetHeader sets the header metadata. It may be called multiple times.
	// When call multiple times, all the provided metadata will be merged.
	// All the metadata will be sent out when one of the following happens:
	//  - ServerStream.SendHeader() is called;
	//  - The first response is sent out;
	//  - An RPC status is sent out (error or success).
	SetHeader(metadata.MD) error
	// SendHeader sends the header metadata.
	// The provided md and headers set by SetHeader() will be sent.
	// It fails if called multiple times.
	SendHeader(metadata.MD) error
	// SetTrailer sets the trailer metadata which will be sent with the RPC status.
	// When called more than once, all the provided metadata will be merged.
	SetTrailer(metadata.MD)
}

type clientStream struct {
	ctx context.Context

	stream net.Conn
	header *streamHeader
	codec  encoding.Codec
	cp     encoding.Compressor
}

func (cs *clientStream) Close() error {
	return cs.stream.Close()
}

func (cs *clientStream) Context() context.Context {
	return cs.ctx
}

func (cs *clientStream) SendMsg(ctx context.Context, m interface{}) error {
	data, err := cs.codec.Marshal(m)
	if err != nil {
		return err
	}
	cookies := CookiesHeader(ctx)
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
	hdr, payload := msgHeader(data, compData)
	if _, err = cs.stream.Write(hdr); err != nil {
		return err
	}
	if _, err = cs.stream.Write(payload); err != nil {
		return err
	}
	return nil
}

func recv(conn io.Reader) (pf payloadFormat, msg []byte, err error) {
	header := make([]byte, headerLen, headerLen)
	if _, err := conn.Read(header[:]); err != nil {
		return 0, nil, err
	}

	pf = payloadFormat(header[0])
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
	pf, msg, err := recv(cs.stream)
	if err != nil {
		return ctx, err
	}
	var data []byte
	if pf == compressionMade {
		dc, _ := cs.cp.Decompress(bytes.NewReader(msg))
		data, err = ioutil.ReadAll(dc)
		if err != nil {
			return ctx, err
		}
	} else {
		data = msg
	}
	ctx, l := ReadCookiesHeader(ctx, data)
	if err = cs.codec.Unmarshal(data[l:], m); err != nil {
		return ctx, errors.New(fmt.Sprintf("grpc: failed to unmarshal the received message %v", err))
	}
	return ctx, nil
}

func (cs *clientStream) Header() (metadata.MD, error) {
	panic("implement me")
}

func (cs *clientStream) Trailer() metadata.MD {
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
	header *streamHeader

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
	// TODO DoPreWriteResponse
	if err = ss.sc.DoPreWriteResponse(ss.ctx, nil, m); err != nil {
		return err
	}
	defer func() {
		// TODO DoPostWriteResponse
		err = ss.sc.DoPostWriteResponse(ss.ctx, nil, m, err)
	}()
	data, err := ss.codec.Marshal(m)
	if err != nil {
		return err
	}
	cookies := CookiesHeader(ctx)
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
	hdr, payload := msgHeader(data, compData)
	if _, err = ss.stream.Write(hdr); err != nil {
		return err
	}
	if _, err = ss.stream.Write(payload); err != nil {
		return err
	}
	return err
}

func (ss *serverStream) RecvMsg(ctx context.Context, m interface{}) (context.Context, error) {
	// TODO DoPreReadRequest
	if err := ss.sc.DoPreReadRequest(ss.ctx); err != nil {
		return ctx, err
	}
	pf, msg, err := recv(ss.stream)
	if err != nil {
		return ctx, err
	}
	var data []byte
	if pf == compressionMade {
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
	ctx, l := ReadCookiesHeader(ctx, data)
	// TODO 分解服务端byte数据
	if err = ss.codec.Unmarshal(data[l:], m); err != nil {
		err = errors.New(fmt.Sprintf("xrpc: failed to unmarshal the received message for %v", err))
	}
	// TODO DoPostReadRequest
	err = ss.sc.DoPostReadRequest(ss.ctx, m, err)
	return ctx, err
}

func (ss *serverStream) SetHeader(metadata.MD) error {
	panic("implement me")
}

func (ss *serverStream) SendHeader(metadata.MD) error {
	panic("implement me")
}

func (ss *serverStream) SetTrailer(metadata.MD) {
	panic("implement me")
}
