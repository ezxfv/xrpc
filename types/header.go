package types

import (
	"encoding/binary"
	"strings"
)

type PayloadFormat uint8
type HeaderCmd string
type Rpc int

const (
	payloadLen = 1
	cookieLen  = 4
	sizeLen    = 4
	HeaderLen  = payloadLen + sizeLen

	compressionNone PayloadFormat = iota // no compression
	CompressionMade                      // compressed
	metaHeader
	CmdHeader

	Init    HeaderCmd = "init"
	Close   HeaderCmd = "close"
	Upgrade HeaderCmd = "upgrade"

	Preface = "xrpc/cheers"

	XRPC   Rpc = 0
	RawRPC Rpc = 1
)

type StreamHeader struct {
	FullMethod string
	Cmd        HeaderCmd
	RpcType    Rpc
	Args       map[string]interface{}
	Payload    []byte
}

func (sh *StreamHeader) SplitMethod() (service, method string) {
	arr := strings.Split(sh.FullMethod, "/")
	if len(arr) != 3 {
		return
	}
	service = arr[1]
	method = arr[2]
	return
}

func GetCodecArg(header *StreamHeader) string {
	c, ok := header.Args["codec"]
	if !ok {
		return "proto"
	}
	return c.(string)
}

func GetCompressorArg(header *StreamHeader) string {
	c, ok := header.Args["compressor"]
	if !ok {
		return "gzip"
	}
	return c.(string)
}

// msgHeader returns a 5-byte header for the message being transmitted and the
// payload, which is compData if non-nil or data otherwise.
func MsgHeader(data []byte, comp bool) (hdr []byte) {
	hdr = make([]byte, HeaderLen)
	if comp {
		hdr[0] = byte(CompressionMade)
	} else {
		hdr[0] = byte(compressionNone)
	}
	// Write length of payload into buf
	binary.BigEndian.PutUint32(hdr[payloadLen:], uint32(len(data)))
	return hdr
}
