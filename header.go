package xrpc

import (
	"encoding/binary"
	"strings"
)

type payloadFormat uint8
type HeaderCmd string

const (
	payloadLen = 1
	sizeLen    = 4
	headerLen  = payloadLen + sizeLen

	compressionNone payloadFormat = 0 // no compression
	compressionMade payloadFormat = 1 // compressed
	MetaHeader      payloadFormat = 2
	CmdHeader       payloadFormat = 3

	Init    HeaderCmd = "init"
	Close   HeaderCmd = "close"
	Upgrade HeaderCmd = "upgrade"

	Preface = "xrpc/cheers"
)

type streamHeader struct {
	FullMethod string
	Cmd        HeaderCmd
	Args       map[string]interface{}
	Payload    []byte
}

func (sh *streamHeader) splitMethod() (service, method string) {
	arr := strings.Split(sh.FullMethod, "/")
	if len(arr) != 3 {
		return
	}
	service = arr[1]
	method = arr[2]
	return
}

func getCodecArg(header *streamHeader) string {
	c, ok := header.Args["codec"]
	if !ok {
		return "proto"
	}
	return c.(string)
}

func getCompressorArg(header *streamHeader) string {
	c, ok := header.Args["compressor"]
	if !ok {
		return "gzip"
	}
	return c.(string)
}

// msgHeader returns a 5-byte header for the message being transmitted and the
// payload, which is compData if non-nil or data otherwise.
func msgHeader(data, compData []byte) (hdr []byte, payload []byte) {
	hdr = make([]byte, headerLen)
	if compData != nil {
		hdr[0] = byte(compressionMade)
		data = compData
	} else {
		hdr[0] = byte(compressionNone)
	}

	// Write length of payload into buf
	binary.BigEndian.PutUint32(hdr[payloadLen:], uint32(len(data)))
	return hdr, data
}
