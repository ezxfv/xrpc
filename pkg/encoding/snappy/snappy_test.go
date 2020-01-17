package snappy

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/golang/snappy"
	"github.com/stretchr/testify/assert"
)

var testData = []byte("snappy")

func TestCompressor(t *testing.T) {
	c := &compressor{}
	c.poolCompressor.New = func() interface{} {
		return &writer{Writer: snappy.NewBufferedWriter(ioutil.Discard), pool: &c.poolCompressor}
	}
	w := bytes.NewBuffer([]byte{})
	z, err := c.Compress(w)
	assert.Equal(t, nil, err)
	n, err := z.Write(testData)
	assert.Equal(t, nil, err)
	assert.Equal(t, len(testData), n)
	assert.Equal(t, nil, z.Close())

	compData := w.Bytes()
	data := make([]byte, len(testData))
	r := bytes.NewBuffer(compData)
	rz, err := c.Decompress(r)
	assert.Equal(t, nil, err)
	n, err = rz.Read(data)
	assert.Equal(t, nil, err)
	assert.Equal(t, len(testData), n)
	assert.Equal(t, testData, data)
}
