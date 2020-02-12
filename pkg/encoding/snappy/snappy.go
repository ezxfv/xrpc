package snappy

import (
	"io"
	"io/ioutil"
	"sync"

	"x.io/xrpc/pkg/encoding"

	"github.com/golang/snappy"
)

const Name = "snappy"

func init() {
	c := &compressor{}
	c.poolCompressor.New = func() interface{} {
		return &writer{Writer: snappy.NewBufferedWriter(ioutil.Discard), pool: &c.poolCompressor}
	}
	encoding.RegisterCompressor(c)
}

type compressor struct {
	poolCompressor   sync.Pool
	poolDecompressor sync.Pool
}

func (c *compressor) Name() string {
	return Name
}

func (c *compressor) Compress(w io.Writer) (io.WriteCloser, error) {
	z := c.poolCompressor.Get().(*writer)
	z.Writer.Reset(w)
	return z, nil
}

func (c *compressor) Decompress(r io.Reader) (io.Reader, error) {
	z, inPool := c.poolDecompressor.Get().(*reader)
	if !inPool {
		newZ := snappy.NewReader(r)
		return &reader{Reader: newZ, pool: &c.poolDecompressor}, nil
	}
	z.Reset(r)
	return z, nil
}

type writer struct {
	*snappy.Writer
	pool *sync.Pool
}

func (z *writer) Close() error {
	defer z.pool.Put(z)
	return z.Writer.Close()
}

type reader struct {
	*snappy.Reader
	pool *sync.Pool
}

func (z *reader) Read(p []byte) (n int, err error) {
	n, err = z.Reader.Read(p)
	if err == io.EOF {
		z.pool.Put(z)
	}
	return n, err
}
