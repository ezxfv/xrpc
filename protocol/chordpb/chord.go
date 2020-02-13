package chordpb

import "x.io/xrpc"

type Chord interface {
	Join(ctx *xrpc.XContext, req *Message) (reply *Message)
	Leave(ctx *xrpc.XContext, req *Message) (reply *Message)
	Lookup(ctx *xrpc.XContext, req *Message) (reply *Message)
	FindSuccessor(ctx *xrpc.XContext, req *Message) (reply *Message)
	Notify(ctx *xrpc.XContext, req *Message) (reply *Message)
	HeartBeat(ctx *xrpc.XContext, req *Message) (reply *Message)

	Set(ctx *xrpc.XContext, req *Message) (reply *Message)
	Get(ctx *xrpc.XContext, req *Message) (reply *Message)
	Del(ctx *xrpc.XContext, req *Message) (reply *Message)
}
