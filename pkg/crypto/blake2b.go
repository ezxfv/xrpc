package crypto

import (
	blake2blib "github.com/minio/blake2b-simd"
)

// Blake2b represents the BLAKE2 cryptographic hash algorithm.
type Blake2b struct{}

var (
	_ HashPolicy = (*Blake2b)(nil)
)

// New returns a BLAKE2 hash policy.
func NewBlake2b() *Blake2b {
	return &Blake2b{}
}

// HashBytes hashes the given bytes using the BLAKE2 hash algorithm.
func (p *Blake2b) HashBytes(bytes []byte) []byte {
	result := blake2blib.Sum256(bytes)
	return result[:]
}

func (p *Blake2b) Size() int {
	return blake2blib.Size
}
