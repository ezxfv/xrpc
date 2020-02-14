

package chordpb

import (
    "context"
    "x.io/xrpc/pkg/codes"
    "x.io/xrpc"
    "fmt"
)

// ChordClient is the client API for Chord service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/x.io/xrpc#ClientConn.NewStream.
type ChordClient interface {
    Join(ctx context.Context, in_1 *Message) (out_1 *Message)
    Leave(ctx context.Context, in_1 *Message) (out_1 *Message)
    Lookup(ctx context.Context, in_1 *Message) (out_1 *Message)
    FindSuccessor(ctx context.Context, in_1 *Message) (out_1 *Message)
    Notify(ctx context.Context, in_1 *Message) (out_1 *Message)
    HeartBeat(ctx context.Context, in_1 *Message) (out_1 *Message)
    Set(ctx context.Context, in_1 *Message) (out_1 *Message)
    Get(ctx context.Context, in_1 *Message) (out_1 *Message)
    Del(ctx context.Context, in_1 *Message) (out_1 *Message)
}

type chordClient struct {
    cc *xrpc.ClientConn
    opts []xrpc.CallOption
}

func NewChordClient(cc *xrpc.ClientConn, opts ...xrpc.CallOption) ChordClient {
    return &chordClient{cc, opts}
}

func (c *chordClient) Join(ctx context.Context, in_1 *Message) (out_1 *Message) {
    var ins, outs []interface{}
    ins = append(ins, in_1)
    outs = append(outs, &out_1)
    err := c.cc.Invoke(ctx, "/chordpb.Chord/Join", ins, &outs, c.opts...)
    if err != nil { return out_1 }
    return out_1
}

func (c *chordClient) Leave(ctx context.Context, in_1 *Message) (out_1 *Message) {
    var ins, outs []interface{}
    ins = append(ins, in_1)
    outs = append(outs, &out_1)
    err := c.cc.Invoke(ctx, "/chordpb.Chord/Leave", ins, &outs, c.opts...)
    if err != nil { return out_1 }
    return out_1
}

func (c *chordClient) Lookup(ctx context.Context, in_1 *Message) (out_1 *Message) {
    var ins, outs []interface{}
    ins = append(ins, in_1)
    outs = append(outs, &out_1)
    err := c.cc.Invoke(ctx, "/chordpb.Chord/Lookup", ins, &outs, c.opts...)
    if err != nil { return out_1 }
    return out_1
}

func (c *chordClient) FindSuccessor(ctx context.Context, in_1 *Message) (out_1 *Message) {
    var ins, outs []interface{}
    ins = append(ins, in_1)
    outs = append(outs, &out_1)
    err := c.cc.Invoke(ctx, "/chordpb.Chord/FindSuccessor", ins, &outs, c.opts...)
    if err != nil { return out_1 }
    return out_1
}

func (c *chordClient) Notify(ctx context.Context, in_1 *Message) (out_1 *Message) {
    var ins, outs []interface{}
    ins = append(ins, in_1)
    outs = append(outs, &out_1)
    err := c.cc.Invoke(ctx, "/chordpb.Chord/Notify", ins, &outs, c.opts...)
    if err != nil { return out_1 }
    return out_1
}

func (c *chordClient) HeartBeat(ctx context.Context, in_1 *Message) (out_1 *Message) {
    var ins, outs []interface{}
    ins = append(ins, in_1)
    outs = append(outs, &out_1)
    err := c.cc.Invoke(ctx, "/chordpb.Chord/HeartBeat", ins, &outs, c.opts...)
    if err != nil { return out_1 }
    return out_1
}

func (c *chordClient) Set(ctx context.Context, in_1 *Message) (out_1 *Message) {
    var ins, outs []interface{}
    ins = append(ins, in_1)
    outs = append(outs, &out_1)
    err := c.cc.Invoke(ctx, "/chordpb.Chord/Set", ins, &outs, c.opts...)
    if err != nil { return out_1 }
    return out_1
}

func (c *chordClient) Get(ctx context.Context, in_1 *Message) (out_1 *Message) {
    var ins, outs []interface{}
    ins = append(ins, in_1)
    outs = append(outs, &out_1)
    err := c.cc.Invoke(ctx, "/chordpb.Chord/Get", ins, &outs, c.opts...)
    if err != nil { return out_1 }
    return out_1
}

func (c *chordClient) Del(ctx context.Context, in_1 *Message) (out_1 *Message) {
    var ins, outs []interface{}
    ins = append(ins, in_1)
    outs = append(outs, &out_1)
    err := c.cc.Invoke(ctx, "/chordpb.Chord/Del", ins, &outs, c.opts...)
    if err != nil { return out_1 }
    return out_1
}

// UnimplementedChord can be embedded to have forward compatible implementations.
type UnimplementedChord struct {
}

func (*UnimplementedChord) Join(ctx *xrpc.XContext, req *Message) (reply *Message) {
    panic(fmt.Sprint(codes.Unimplemented, "method Join not implemented"))
}

func (*UnimplementedChord) Leave(ctx *xrpc.XContext, req *Message) (reply *Message) {
    panic(fmt.Sprint(codes.Unimplemented, "method Leave not implemented"))
}

func (*UnimplementedChord) Lookup(ctx *xrpc.XContext, req *Message) (reply *Message) {
    panic(fmt.Sprint(codes.Unimplemented, "method Lookup not implemented"))
}

func (*UnimplementedChord) FindSuccessor(ctx *xrpc.XContext, req *Message) (reply *Message) {
    panic(fmt.Sprint(codes.Unimplemented, "method FindSuccessor not implemented"))
}

func (*UnimplementedChord) Notify(ctx *xrpc.XContext, req *Message) (reply *Message) {
    panic(fmt.Sprint(codes.Unimplemented, "method Notify not implemented"))
}

func (*UnimplementedChord) HeartBeat(ctx *xrpc.XContext, req *Message) (reply *Message) {
    panic(fmt.Sprint(codes.Unimplemented, "method HeartBeat not implemented"))
}

func (*UnimplementedChord) Set(ctx *xrpc.XContext, req *Message) (reply *Message) {
    panic(fmt.Sprint(codes.Unimplemented, "method Set not implemented"))
}

func (*UnimplementedChord) Get(ctx *xrpc.XContext, req *Message) (reply *Message) {
    panic(fmt.Sprint(codes.Unimplemented, "method Get not implemented"))
}

func (*UnimplementedChord) Del(ctx *xrpc.XContext, req *Message) (reply *Message) {
    panic(fmt.Sprint(codes.Unimplemented, "method Del not implemented"))
}

func RegisterChordServer(s *xrpc.Server, srv Chord) {
    s.RegisterService(&_Chord_serviceDesc, srv)
}

func _Chord_Join_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor xrpc.UnaryServerInterceptor) (interface{}, error) {
    var ins []interface{}
    var (
        xctx  = xrpc.XBackground()
        in_1 = new(Message)
    )
    xctx.SetCtx(ctx)
    ins = append(ins, in_1)
    var (
        out_1 = new(Message)
    )
    if err := dec(&ins); err != nil { return nil, err }
    if interceptor == nil {
        var results []interface{}
        out_1 = srv.(Chord).Join(xctx, in_1)
        results = append(results, out_1)
        return results, nil
    }
    info := &xrpc.UnaryServerInfo{
        Server: srv,
        FullMethod: "/chordpb.Chord/Join",
    }
    handler := func(ctx context.Context, req interface{}) (interface{}, error) {
        var results []interface{}
        out_1 = srv.(Chord).Join(xctx, in_1)
        results = append(results, out_1)
        return results, nil
    }
    return interceptor(ctx, ins, info, handler)
}

func _Chord_Leave_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor xrpc.UnaryServerInterceptor) (interface{}, error) {
    var ins []interface{}
    var (
        xctx  = xrpc.XBackground()
        in_1 = new(Message)
    )
    xctx.SetCtx(ctx)
    ins = append(ins, in_1)
    var (
        out_1 = new(Message)
    )
    if err := dec(&ins); err != nil { return nil, err }
    if interceptor == nil {
        var results []interface{}
        out_1 = srv.(Chord).Leave(xctx, in_1)
        results = append(results, out_1)
        return results, nil
    }
    info := &xrpc.UnaryServerInfo{
        Server: srv,
        FullMethod: "/chordpb.Chord/Leave",
    }
    handler := func(ctx context.Context, req interface{}) (interface{}, error) {
        var results []interface{}
        out_1 = srv.(Chord).Leave(xctx, in_1)
        results = append(results, out_1)
        return results, nil
    }
    return interceptor(ctx, ins, info, handler)
}

func _Chord_Lookup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor xrpc.UnaryServerInterceptor) (interface{}, error) {
    var ins []interface{}
    var (
        xctx  = xrpc.XBackground()
        in_1 = new(Message)
    )
    xctx.SetCtx(ctx)
    ins = append(ins, in_1)
    var (
        out_1 = new(Message)
    )
    if err := dec(&ins); err != nil { return nil, err }
    if interceptor == nil {
        var results []interface{}
        out_1 = srv.(Chord).Lookup(xctx, in_1)
        results = append(results, out_1)
        return results, nil
    }
    info := &xrpc.UnaryServerInfo{
        Server: srv,
        FullMethod: "/chordpb.Chord/Lookup",
    }
    handler := func(ctx context.Context, req interface{}) (interface{}, error) {
        var results []interface{}
        out_1 = srv.(Chord).Lookup(xctx, in_1)
        results = append(results, out_1)
        return results, nil
    }
    return interceptor(ctx, ins, info, handler)
}

func _Chord_FindSuccessor_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor xrpc.UnaryServerInterceptor) (interface{}, error) {
    var ins []interface{}
    var (
        xctx  = xrpc.XBackground()
        in_1 = new(Message)
    )
    xctx.SetCtx(ctx)
    ins = append(ins, in_1)
    var (
        out_1 = new(Message)
    )
    if err := dec(&ins); err != nil { return nil, err }
    if interceptor == nil {
        var results []interface{}
        out_1 = srv.(Chord).FindSuccessor(xctx, in_1)
        results = append(results, out_1)
        return results, nil
    }
    info := &xrpc.UnaryServerInfo{
        Server: srv,
        FullMethod: "/chordpb.Chord/FindSuccessor",
    }
    handler := func(ctx context.Context, req interface{}) (interface{}, error) {
        var results []interface{}
        out_1 = srv.(Chord).FindSuccessor(xctx, in_1)
        results = append(results, out_1)
        return results, nil
    }
    return interceptor(ctx, ins, info, handler)
}

func _Chord_Notify_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor xrpc.UnaryServerInterceptor) (interface{}, error) {
    var ins []interface{}
    var (
        xctx  = xrpc.XBackground()
        in_1 = new(Message)
    )
    xctx.SetCtx(ctx)
    ins = append(ins, in_1)
    var (
        out_1 = new(Message)
    )
    if err := dec(&ins); err != nil { return nil, err }
    if interceptor == nil {
        var results []interface{}
        out_1 = srv.(Chord).Notify(xctx, in_1)
        results = append(results, out_1)
        return results, nil
    }
    info := &xrpc.UnaryServerInfo{
        Server: srv,
        FullMethod: "/chordpb.Chord/Notify",
    }
    handler := func(ctx context.Context, req interface{}) (interface{}, error) {
        var results []interface{}
        out_1 = srv.(Chord).Notify(xctx, in_1)
        results = append(results, out_1)
        return results, nil
    }
    return interceptor(ctx, ins, info, handler)
}

func _Chord_HeartBeat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor xrpc.UnaryServerInterceptor) (interface{}, error) {
    var ins []interface{}
    var (
        xctx  = xrpc.XBackground()
        in_1 = new(Message)
    )
    xctx.SetCtx(ctx)
    ins = append(ins, in_1)
    var (
        out_1 = new(Message)
    )
    if err := dec(&ins); err != nil { return nil, err }
    if interceptor == nil {
        var results []interface{}
        out_1 = srv.(Chord).HeartBeat(xctx, in_1)
        results = append(results, out_1)
        return results, nil
    }
    info := &xrpc.UnaryServerInfo{
        Server: srv,
        FullMethod: "/chordpb.Chord/HeartBeat",
    }
    handler := func(ctx context.Context, req interface{}) (interface{}, error) {
        var results []interface{}
        out_1 = srv.(Chord).HeartBeat(xctx, in_1)
        results = append(results, out_1)
        return results, nil
    }
    return interceptor(ctx, ins, info, handler)
}

func _Chord_Set_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor xrpc.UnaryServerInterceptor) (interface{}, error) {
    var ins []interface{}
    var (
        xctx  = xrpc.XBackground()
        in_1 = new(Message)
    )
    xctx.SetCtx(ctx)
    ins = append(ins, in_1)
    var (
        out_1 = new(Message)
    )
    if err := dec(&ins); err != nil { return nil, err }
    if interceptor == nil {
        var results []interface{}
        out_1 = srv.(Chord).Set(xctx, in_1)
        results = append(results, out_1)
        return results, nil
    }
    info := &xrpc.UnaryServerInfo{
        Server: srv,
        FullMethod: "/chordpb.Chord/Set",
    }
    handler := func(ctx context.Context, req interface{}) (interface{}, error) {
        var results []interface{}
        out_1 = srv.(Chord).Set(xctx, in_1)
        results = append(results, out_1)
        return results, nil
    }
    return interceptor(ctx, ins, info, handler)
}

func _Chord_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor xrpc.UnaryServerInterceptor) (interface{}, error) {
    var ins []interface{}
    var (
        xctx  = xrpc.XBackground()
        in_1 = new(Message)
    )
    xctx.SetCtx(ctx)
    ins = append(ins, in_1)
    var (
        out_1 = new(Message)
    )
    if err := dec(&ins); err != nil { return nil, err }
    if interceptor == nil {
        var results []interface{}
        out_1 = srv.(Chord).Get(xctx, in_1)
        results = append(results, out_1)
        return results, nil
    }
    info := &xrpc.UnaryServerInfo{
        Server: srv,
        FullMethod: "/chordpb.Chord/Get",
    }
    handler := func(ctx context.Context, req interface{}) (interface{}, error) {
        var results []interface{}
        out_1 = srv.(Chord).Get(xctx, in_1)
        results = append(results, out_1)
        return results, nil
    }
    return interceptor(ctx, ins, info, handler)
}

func _Chord_Del_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor xrpc.UnaryServerInterceptor) (interface{}, error) {
    var ins []interface{}
    var (
        xctx  = xrpc.XBackground()
        in_1 = new(Message)
    )
    xctx.SetCtx(ctx)
    ins = append(ins, in_1)
    var (
        out_1 = new(Message)
    )
    if err := dec(&ins); err != nil { return nil, err }
    if interceptor == nil {
        var results []interface{}
        out_1 = srv.(Chord).Del(xctx, in_1)
        results = append(results, out_1)
        return results, nil
    }
    info := &xrpc.UnaryServerInfo{
        Server: srv,
        FullMethod: "/chordpb.Chord/Del",
    }
    handler := func(ctx context.Context, req interface{}) (interface{}, error) {
        var results []interface{}
        out_1 = srv.(Chord).Del(xctx, in_1)
        results = append(results, out_1)
        return results, nil
    }
    return interceptor(ctx, ins, info, handler)
}

var _Chord_serviceDesc = xrpc.ServiceDesc {
    ServiceName: "chordpb.Chord",
    HandlerType: (*Chord)(nil),
    Methods: []xrpc.MethodDesc{
        {
            MethodName: "Join",
            Handler: _Chord_Join_Handler,
        },
        {
            MethodName: "Leave",
            Handler: _Chord_Leave_Handler,
        },
        {
            MethodName: "Lookup",
            Handler: _Chord_Lookup_Handler,
        },
        {
            MethodName: "FindSuccessor",
            Handler: _Chord_FindSuccessor_Handler,
        },
        {
            MethodName: "Notify",
            Handler: _Chord_Notify_Handler,
        },
        {
            MethodName: "HeartBeat",
            Handler: _Chord_HeartBeat_Handler,
        },
        {
            MethodName: "Set",
            Handler: _Chord_Set_Handler,
        },
        {
            MethodName: "Get",
            Handler: _Chord_Get_Handler,
        },
        {
            MethodName: "Del",
            Handler: _Chord_Del_Handler,
        },
    },
    Streams: []xrpc.StreamDesc{},
    Metadata: "chordpb",
}

