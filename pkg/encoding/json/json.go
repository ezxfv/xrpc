package json

import (
	"x.io/xrpc/pkg/encoding"

	jsoniter "github.com/json-iterator/go"
)

const Name = "json"

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func init() {
	encoding.RegisterCodec(&codec{})
}

type codec struct {
}

func (c *codec) Name() string {
	return Name
}

func (c *codec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (c *codec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
