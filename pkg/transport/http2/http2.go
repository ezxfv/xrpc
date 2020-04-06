package transport

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"x.io/xrpc/pkg/encoding"
	"x.io/xrpc/types"
)

type http2ServerTransport struct {
	tls bool
	w   http.ResponseWriter
	r   *http.Request

	buf    chan *http.Response
	header *types.StreamHeader
	codec  encoding.Codec
	cp     encoding.Compressor
}

func (t *http2ServerTransport) Protocol() string {
	if t.tls {
		return "http2tls"
	}
	return "http2"
}

func (t *http2ServerTransport) SendMsg(ctx context.Context, m interface{}) error {
	n, err := t.w.Write(m.([]byte))
	if n != len(m.([]byte)) {
		return errors.New("write content length isn't match")
	}
	return err
}

func (t *http2ServerTransport) RecvMsg(ctx context.Context, m interface{}) (context.Context, error) {
	if t.r.Body == nil {
		return nil, errors.New("req body is nil")
	}
	data, err := ioutil.ReadAll(t.r.Body)
	if err != nil {
		return nil, err
	}
	ctx, l := types.ReadCookiesHeader(ctx, data)
	if err = t.codec.Unmarshal(data[l:], m); err != nil {
		return ctx, errors.New(fmt.Sprintf("xrpc: failed to unmarshal the received message %v", err))
	}
	return ctx, nil
}

func (t *http2ServerTransport) Close() error {
	panic("implement me")
}

type http2ClientTransport struct {
	tls    bool
	c      *http.Client
	buf    chan *http.Response
	header *types.StreamHeader
	codec  encoding.Codec
	cp     encoding.Compressor
}

func (t *http2ClientTransport) Protocol() string {
	if t.tls {
		return "http2tls"
	}
	return "http2"
}

func (t *http2ClientTransport) SendMsg(ctx context.Context, m interface{}) error {
	r := bytes.NewReader(m.([]byte))
	resp, err := t.c.Post("http://", "application/raw-data", r)
	t.buf <- resp
	return err
}

func (t *http2ClientTransport) RecvMsg(ctx context.Context, m interface{}) (context.Context, error) {
	res, ok := <-t.buf
	if !ok {
		return nil, errors.New("client buf is closed")
	}
	if res.Body == nil {
		return nil, errors.New("resp body is nil")
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	ctx, l := types.ReadCookiesHeader(ctx, data)
	if err = t.codec.Unmarshal(data[l:], m); err != nil {
		return ctx, errors.New(fmt.Sprintf("xrpc: failed to unmarshal the received message %v", err))
	}
	return ctx, nil
}

func (t *http2ClientTransport) Close() error {
	t.c.CloseIdleConnections()
	return nil
}
