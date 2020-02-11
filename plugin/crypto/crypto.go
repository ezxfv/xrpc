package crypto

import (
	"context"
	"errors"

	"github.com/edenzhong7/xrpc"

	"github.com/edenzhong7/xrpc/pkg/crypto"
)

const (
	Key string = "session_id"
)

func New() *cryptoPlugin {
	return &cryptoPlugin{
		aes:     crypto.NewAesEncryptor(),
		blake2b: crypto.NewBlake2b(),
		keys:    map[string][]byte{},
	}
}

type cryptoPlugin struct {
	aes     *crypto.AesEncryptor
	blake2b *crypto.Blake2b

	keys map[string][]byte
}

func (c *cryptoPlugin) SetKey(key, pass string) {
	c.keys[key] = c.blake2b.HashBytes([]byte(pass))
}

func (c *cryptoPlugin) PreReadRequest(ctx context.Context, data []byte) ([]byte, error) {
	key := xrpc.GetCookie(ctx, Key)
	pass, ok := c.keys[key]
	if !ok {
		return data, errors.New("check session key failed")
	}
	data, err := c.aes.Decrypt(data, pass)
	return data, err
}

func (c *cryptoPlugin) PreWriteResponse(ctx context.Context, data []byte) ([]byte, error) {
	key := xrpc.GetCookie(ctx, Key)
	pass, ok := c.keys[key]
	if !ok {
		return data, errors.New("read session key failed")
	}
	data, err := c.aes.Encrypt(data, pass)
	return data, err
}
