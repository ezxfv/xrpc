package xrpc

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// Stream defines the common interface a client or server stream has to satisfy.
type Stream interface {
	// Context returns the context for this stream.
	Context() context.Context
	// SendMsg blocks until it sends m, the stream is done or the stream
	// breaks.
	// On error, it aborts the stream and returns an RPC status on client
	// side. On server side, it simply returns the error to the caller.
	// SendMsg is called by generated code. Also Users can call SendMsg
	// directly when it is really needed in their use cases.
	// It's safe to have a goroutine calling SendMsg and another goroutine calling
	// recvMsg on the same stream at the same time.
	// But it is not safe to call SendMsg on the same stream in different goroutines.
	SendMsg(m interface{}) error
	// RecvMsg blocks until it receives a message or the stream is
	// done. On client side, it returns io.EOF when the stream is done. On
	// any other error, it aborts the stream and returns an RPC status. On
	// server side, it simply returns the error to the caller.
	// It's safe to have a goroutine calling SendMsg and another goroutine calling
	// recvMsg on the same stream at the same time.
	// But it is not safe to call RecvMsg on the same stream in different goroutines.
	RecvMsg(m interface{}) error
}

type ClientStream interface {
	// Stream.SendMsg() may return a non-nil error when something wrong happens sending
	// the request. The returned error indicates the status of this sending, not the final
	// status of the RPC.
	// Always call Stream.RecvMsg() to get the final status if you care about the status of
	// the RPC.
	Stream
	// Header returns the header metadata received from the server if there
	// is any. It blocks if the metadata is not ready to read.
	Header() (metadata.MD, error)
	// Trailer returns the trailer metadata from the server, if there is any.
	// It must only be called after stream.CloseAndRecv has returned, or
	// stream.Recv has returned a non-nil error (including io.EOF).
	Trailer() metadata.MD
	// CloseSend closes the send direction of the stream. It closes the stream
	// when non-nil error is met.
	CloseSend() error
}

type ServerStream interface {
	Stream
	// SetHeader sets the header metadata. It may be called multiple times.
	// When call multiple times, all the provided metadata will be merged.
	// All the metadata will be sent out when one of the following happens:
	//  - ServerStream.SendHeader() is called;
	//  - The first response is sent out;
	//  - An RPC status is sent out (error or success).
	SetHeader(metadata.MD) error
	// SendHeader sends the header metadata.
	// The provided md and headers set by SetHeader() will be sent.
	// It fails if called multiple times.
	SendHeader(metadata.MD) error
	// SetTrailer sets the trailer metadata which will be sent with the RPC status.
	// When called more than once, all the provided metadata will be merged.
	SetTrailer(metadata.MD)
}
