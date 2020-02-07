package xrpc

import (
	"time"

	"github.com/edenzhong7/xrpc/pkg/net"
)

type options struct {
	writeBufferSize   int
	readBufferSize    int
	connectionTimeout time.Duration
}

type ConnectOptions struct {
	dialer net.Dialer
}

// dialOptions configure a Dial call. dialOptions are set by the DialOption
// values passed to Dial.
type dialOptions struct {
	block       bool
	insecure    bool
	timeout     time.Duration
	copts       ConnectOptions
	callOptions []CallOption
	codec       string
	compressor  string
}

// A ServerOption sets options such as credentials, codec and keepalive parameters, etc.
type ServerOption func(opts *options)

type DialOption interface {
	apply(dopts *dialOptions)
}

type dialOption struct {
	f func(*dialOptions)
}

func (d *dialOption) apply(doyts *dialOptions) {
	d.f(doyts)
}

func WithJsonCodec() DialOption {
	return &dialOption{func(dopts *dialOptions) {
		dopts.codec = "json"
	}}
}

func WithSnappyCompressor() DialOption {
	return &dialOption{func(dopts *dialOptions) {
		dopts.compressor = "snappy"
	}}
}

// WithInsecure returns a DialOption which disables transport security for this
// ClientConn. Note that transport security is required unless WithInsecure is
// set.
func WithInsecure() DialOption {
	return newFuncDialOption(func(o *dialOptions) {
		o.insecure = true
	})
}

// funcDialOption wraps a function that modifies dialOptions into an
// implementation of the DialOption interface.
type funcDialOption struct {
	f func(*dialOptions)
}

func (fdo *funcDialOption) apply(do *dialOptions) {
	fdo.f(do)
}

func newFuncDialOption(f func(*dialOptions)) *funcDialOption {
	return &funcDialOption{
		f: f,
	}
}
