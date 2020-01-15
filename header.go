package xrpc

import "encoding/binary"

type payloadFormat uint8

const (
	payloadLen = 1
	sizeLen    = 4
	headerLen  = payloadLen + sizeLen

	compressionNone payloadFormat = 0 // no compression
	compressionMade payloadFormat = 1 // compressed
)

type streamHeader struct {
	FullMethod string
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
