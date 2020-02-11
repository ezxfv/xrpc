package crypto_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/edenzhong7/xrpc/pkg/crypto"
)

func TestAesEncryptor(t *testing.T) {
	ae := crypto.NewAesEncryptor()
	be := crypto.NewBlake2b()
	msg := []byte("hello")
	key := []byte("1234")
	hashKey := be.HashBytes(key)
	secret, err := ae.Encrypt(msg, hashKey)
	new_msg, err := ae.Decrypt(secret, hashKey)
	assert.Equal(t, nil, err)
	assert.Equal(t, msg, new_msg)
}
