

package math

import (
    "github.com/edenzhong7/xrpc"
    "fmt"
    "context"
    "github.com/edenzhong7/xrpc/pkg/codes"
)

// CounterClient is the client API for Counter service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/github.com/edenzhong7/xrpc#ClientConn.NewStream.
type CounterClient interface {
    Inc(ctx context.Context, in_1 *Num) (out_1 int32, out_2 *Num)
    Dec(ctx context.Context, in_1 Num) (out_1 *Num)
}

type counterClient struct {
    cc *xrpc.ClientConn
    opts []xrpc.CallOption
}

func NewCounterClient(cc *xrpc.ClientConn, opts ...xrpc.CallOption) CounterClient {
    return &counterClient{cc, opts}
}

func (c *counterClient) Inc(ctx context.Context, in_1 *Num) (out_1 int32, out_2 *Num) {
    var ins, outs []interface{}
    ins = append(ins, in_1)
    outs = append(outs, &out_1, &out_2)
    err := c.cc.Invoke(ctx, "/math.Counter/Inc", ins, &outs, c.opts...)
    if err != nil { return out_1, out_2 }
    return out_1, out_2
}

func (c *counterClient) Dec(ctx context.Context, in_1 Num) (out_1 *Num) {
    var ins, outs []interface{}
    ins = append(ins, in_1)
    outs = append(outs, &out_1)
    err := c.cc.Invoke(ctx, "/math.Counter/Dec", ins, &outs, c.opts...)
    if err != nil { return out_1 }
    return out_1
}

// MathClient is the client API for Math service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/github.com/edenzhong7/xrpc#ClientConn.NewStream.
type MathClient interface {
    Inc(ctx context.Context, in_1 *Num) (out_1 int32, out_2 *Num)
    Dec(ctx context.Context, in_1 Num) (out_1 *Num)
    XRpcAdd(ctx context.Context, in_1,in_2 int) (out_1 int)
    XRpcDouble(ctx context.Context, in_1 int) (out_1 int)
    Add(ctx context.Context, in_1,in_2 int) (out_1 int)
    Double(ctx context.Context, in_1 int) (out_1 int)
    Calc(ctx context.Context, in_1 ...int) (out_1 int, out_2 float64)
}

type mathClient struct {
    cc *xrpc.ClientConn
    opts []xrpc.CallOption
}

func NewMathClient(cc *xrpc.ClientConn, opts ...xrpc.CallOption) MathClient {
    return &mathClient{cc, opts}
}

func (c *mathClient) Inc(ctx context.Context, in_1 *Num) (out_1 int32, out_2 *Num) {
    var ins, outs []interface{}
    ins = append(ins, in_1)
    outs = append(outs, &out_1, &out_2)
    err := c.cc.Invoke(ctx, "/math.Math/Inc", ins, &outs, c.opts...)
    if err != nil { return out_1, out_2 }
    return out_1, out_2
}

func (c *mathClient) Dec(ctx context.Context, in_1 Num) (out_1 *Num) {
    var ins, outs []interface{}
    ins = append(ins, in_1)
    outs = append(outs, &out_1)
    err := c.cc.Invoke(ctx, "/math.Math/Dec", ins, &outs, c.opts...)
    if err != nil { return out_1 }
    return out_1
}

func (c *mathClient) XRpcAdd(ctx context.Context, in_1,in_2 int) (out_1 int) {
    var ins, outs []interface{}
    ins = append(ins, in_1, in_2)
    outs = append(outs, &out_1)
    err := c.cc.Invoke(ctx, "/math.Math/XRpcAdd", ins, &outs, c.opts...)
    if err != nil { return out_1 }
    return out_1
}

func (c *mathClient) XRpcDouble(ctx context.Context, in_1 int) (out_1 int) {
    var ins, outs []interface{}
    ins = append(ins, in_1)
    outs = append(outs, &out_1)
    err := c.cc.Invoke(ctx, "/math.Math/XRpcDouble", ins, &outs, c.opts...)
    if err != nil { return out_1 }
    return out_1
}

func (c *mathClient) Add(ctx context.Context, in_1,in_2 int) (out_1 int) {
    var ins, outs []interface{}
    ins = append(ins, in_1, in_2)
    outs = append(outs, &out_1)
    err := c.cc.Invoke(ctx, "/math.Math/Add", ins, &outs, c.opts...)
    if err != nil { return out_1 }
    return out_1
}

func (c *mathClient) Double(ctx context.Context, in_1 int) (out_1 int) {
    var ins, outs []interface{}
    ins = append(ins, in_1)
    outs = append(outs, &out_1)
    err := c.cc.Invoke(ctx, "/math.Math/Double", ins, &outs, c.opts...)
    if err != nil { return out_1 }
    return out_1
}

func (c *mathClient) Calc(ctx context.Context, in_1 ...int) (out_1 int, out_2 float64) {
    var ins, outs []interface{}
    ins = append(ins, in_1)
    outs = append(outs, &out_1, &out_2)
    err := c.cc.Invoke(ctx, "/math.Math/Calc", ins, &outs, c.opts...)
    if err != nil { return out_1, out_2 }
    return out_1, out_2
}

// UnimplementedCounter can be embedded to have forward compatible implementations.
type UnimplementedCounter struct {
}

func (*UnimplementedCounter) Inc(n *Num) (int32, *Num) {
    panic(fmt.Sprint(codes.Unimplemented, "method Inc not implemented"))
}

func (*UnimplementedCounter) Dec(n Num) *Num {
    panic(fmt.Sprint(codes.Unimplemented, "method Dec not implemented"))
}

func RegisterCounterServer(s *xrpc.Server, srv Counter) {
    s.RegisterService(&_Counter_serviceDesc, srv)
}

func _Counter_Inc_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor xrpc.UnaryServerInterceptor) (interface{}, error) {
    var ins []interface{}
    var (
        in_1 = new(Num)
    )
    ins = append(ins, in_1)
    var (
        out_1 int32
        out_2 = new(Num)
    )
    if err := dec(&ins); err != nil { return nil, err }
    if interceptor == nil {
        var results []interface{}
        out_1, out_2 = srv.(Counter).Inc(in_1)
        results = append(results, out_1, out_2)
        return results, nil
    }
    info := &xrpc.UnaryServerInfo{
        Server: srv,
        FullMethod: "/math.Counter/Inc",
    }
    handler := func(ctx context.Context, req interface{}) (interface{}, error) {
        var results []interface{}
        out_1, out_2 = srv.(Counter).Inc(in_1)
        results = append(results, out_1, out_2)
        return results, nil
    }
    return interceptor(ctx, ins, info, handler)
}

func _Counter_Dec_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor xrpc.UnaryServerInterceptor) (interface{}, error) {
    var ins []interface{}
    var (
        in_1 Num
    )
    ins = append(ins, &in_1)
    var (
        out_1 = new(Num)
    )
    if err := dec(&ins); err != nil { return nil, err }
    if interceptor == nil {
        var results []interface{}
        out_1 = srv.(Counter).Dec(in_1)
        results = append(results, out_1)
        return results, nil
    }
    info := &xrpc.UnaryServerInfo{
        Server: srv,
        FullMethod: "/math.Counter/Dec",
    }
    handler := func(ctx context.Context, req interface{}) (interface{}, error) {
        var results []interface{}
        out_1 = srv.(Counter).Dec(in_1)
        results = append(results, out_1)
        return results, nil
    }
    return interceptor(ctx, ins, info, handler)
}

var _Counter_serviceDesc = xrpc.ServiceDesc {
    ServiceName: "math.Counter",
    HandlerType: (*Counter)(nil),
    Methods: []xrpc.MethodDesc{
        {
            MethodName: "Inc",
            Handler: _Counter_Inc_Handler,
        },
        {
            MethodName: "Dec",
            Handler: _Counter_Dec_Handler,
        },
    },
    Streams: []xrpc.StreamDesc{},
    Metadata: "math",
}

// UnimplementedMath can be embedded to have forward compatible implementations.
type UnimplementedMath struct {
}

func (*UnimplementedMath) Inc(n *Num) (int32, *Num) {
    panic(fmt.Sprint(codes.Unimplemented, "method Inc not implemented"))
}

func (*UnimplementedMath) Dec(n Num) *Num {
    panic(fmt.Sprint(codes.Unimplemented, "method Dec not implemented"))
}

func (*UnimplementedMath) XRpcAdd(ctx *xrpc.XContext, a,b int) int {
    panic(fmt.Sprint(codes.Unimplemented, "method XRpcAdd not implemented"))
}

func (*UnimplementedMath) XRpcDouble(ctx *xrpc.XContext, a int) int {
    panic(fmt.Sprint(codes.Unimplemented, "method XRpcDouble not implemented"))
}

func (*UnimplementedMath) Add(a,b int) int {
    panic(fmt.Sprint(codes.Unimplemented, "method Add not implemented"))
}

func (*UnimplementedMath) Double(a int) int {
    panic(fmt.Sprint(codes.Unimplemented, "method Double not implemented"))
}

func (*UnimplementedMath) Calc(ints ...int) (int, float64) {
    panic(fmt.Sprint(codes.Unimplemented, "method Calc not implemented"))
}

func RegisterMathServer(s *xrpc.Server, srv Math) {
    s.RegisterService(&_Math_serviceDesc, srv)
}

func _Math_Inc_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor xrpc.UnaryServerInterceptor) (interface{}, error) {
    var ins []interface{}
    var (
        in_1 = new(Num)
    )
    ins = append(ins, in_1)
    var (
        out_1 int32
        out_2 = new(Num)
    )
    if err := dec(&ins); err != nil { return nil, err }
    if interceptor == nil {
        var results []interface{}
        out_1, out_2 = srv.(Math).Inc(in_1)
        results = append(results, out_1, out_2)
        return results, nil
    }
    info := &xrpc.UnaryServerInfo{
        Server: srv,
        FullMethod: "/math.Math/Inc",
    }
    handler := func(ctx context.Context, req interface{}) (interface{}, error) {
        var results []interface{}
        out_1, out_2 = srv.(Math).Inc(in_1)
        results = append(results, out_1, out_2)
        return results, nil
    }
    return interceptor(ctx, ins, info, handler)
}

func _Math_Dec_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor xrpc.UnaryServerInterceptor) (interface{}, error) {
    var ins []interface{}
    var (
        in_1 Num
    )
    ins = append(ins, &in_1)
    var (
        out_1 = new(Num)
    )
    if err := dec(&ins); err != nil { return nil, err }
    if interceptor == nil {
        var results []interface{}
        out_1 = srv.(Math).Dec(in_1)
        results = append(results, out_1)
        return results, nil
    }
    info := &xrpc.UnaryServerInfo{
        Server: srv,
        FullMethod: "/math.Math/Dec",
    }
    handler := func(ctx context.Context, req interface{}) (interface{}, error) {
        var results []interface{}
        out_1 = srv.(Math).Dec(in_1)
        results = append(results, out_1)
        return results, nil
    }
    return interceptor(ctx, ins, info, handler)
}

func _Math_XRpcAdd_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor xrpc.UnaryServerInterceptor) (interface{}, error) {
    var ins []interface{}
    var (
        xctx  = xrpc.XBackground()
        in_1 int
        in_2 int
    )
    xctx.SetCtx(ctx)
    ins = append(ins, &in_1, &in_2)
    var (
        out_1 int
    )
    if err := dec(&ins); err != nil { return nil, err }
    if interceptor == nil {
        var results []interface{}
        out_1 = srv.(Math).XRpcAdd(xctx, in_1, in_2)
        results = append(results, out_1)
        return results, nil
    }
    info := &xrpc.UnaryServerInfo{
        Server: srv,
        FullMethod: "/math.Math/XRpcAdd",
    }
    handler := func(ctx context.Context, req interface{}) (interface{}, error) {
        var results []interface{}
        out_1 = srv.(Math).XRpcAdd(xctx, in_1, in_2)
        results = append(results, out_1)
        return results, nil
    }
    return interceptor(ctx, ins, info, handler)
}

func _Math_XRpcDouble_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor xrpc.UnaryServerInterceptor) (interface{}, error) {
    var ins []interface{}
    var (
        xctx  = xrpc.XBackground()
        in_1 int
    )
    xctx.SetCtx(ctx)
    ins = append(ins, &in_1)
    var (
        out_1 int
    )
    if err := dec(&ins); err != nil { return nil, err }
    if interceptor == nil {
        var results []interface{}
        out_1 = srv.(Math).XRpcDouble(xctx, in_1)
        results = append(results, out_1)
        return results, nil
    }
    info := &xrpc.UnaryServerInfo{
        Server: srv,
        FullMethod: "/math.Math/XRpcDouble",
    }
    handler := func(ctx context.Context, req interface{}) (interface{}, error) {
        var results []interface{}
        out_1 = srv.(Math).XRpcDouble(xctx, in_1)
        results = append(results, out_1)
        return results, nil
    }
    return interceptor(ctx, ins, info, handler)
}

func _Math_Add_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor xrpc.UnaryServerInterceptor) (interface{}, error) {
    var ins []interface{}
    var (
        in_1 int
        in_2 int
    )
    ins = append(ins, &in_1, &in_2)
    var (
        out_1 int
    )
    if err := dec(&ins); err != nil { return nil, err }
    if interceptor == nil {
        var results []interface{}
        out_1 = srv.(Math).Add(in_1, in_2)
        results = append(results, out_1)
        return results, nil
    }
    info := &xrpc.UnaryServerInfo{
        Server: srv,
        FullMethod: "/math.Math/Add",
    }
    handler := func(ctx context.Context, req interface{}) (interface{}, error) {
        var results []interface{}
        out_1 = srv.(Math).Add(in_1, in_2)
        results = append(results, out_1)
        return results, nil
    }
    return interceptor(ctx, ins, info, handler)
}

func _Math_Double_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor xrpc.UnaryServerInterceptor) (interface{}, error) {
    var ins []interface{}
    var (
        in_1 int
    )
    ins = append(ins, &in_1)
    var (
        out_1 int
    )
    if err := dec(&ins); err != nil { return nil, err }
    if interceptor == nil {
        var results []interface{}
        out_1 = srv.(Math).Double(in_1)
        results = append(results, out_1)
        return results, nil
    }
    info := &xrpc.UnaryServerInfo{
        Server: srv,
        FullMethod: "/math.Math/Double",
    }
    handler := func(ctx context.Context, req interface{}) (interface{}, error) {
        var results []interface{}
        out_1 = srv.(Math).Double(in_1)
        results = append(results, out_1)
        return results, nil
    }
    return interceptor(ctx, ins, info, handler)
}

func _Math_Calc_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor xrpc.UnaryServerInterceptor) (interface{}, error) {
    var ins []interface{}
    var (
        in_1 []int
    )
    ins = append(ins, &in_1)
    var (
        out_1 int
        out_2 float64
    )
    if err := dec(&ins); err != nil { return nil, err }
    if interceptor == nil {
        var results []interface{}
        out_1, out_2 = srv.(Math).Calc(in_1...)
        results = append(results, out_1, out_2)
        return results, nil
    }
    info := &xrpc.UnaryServerInfo{
        Server: srv,
        FullMethod: "/math.Math/Calc",
    }
    handler := func(ctx context.Context, req interface{}) (interface{}, error) {
        var results []interface{}
        out_1, out_2 = srv.(Math).Calc(in_1...)
        results = append(results, out_1, out_2)
        return results, nil
    }
    return interceptor(ctx, ins, info, handler)
}

var _Math_serviceDesc = xrpc.ServiceDesc {
    ServiceName: "math.Math",
    HandlerType: (*Math)(nil),
    Methods: []xrpc.MethodDesc{
        {
            MethodName: "Inc",
            Handler: _Math_Inc_Handler,
        },
        {
            MethodName: "Dec",
            Handler: _Math_Dec_Handler,
        },
        {
            MethodName: "XRpcAdd",
            Handler: _Math_XRpcAdd_Handler,
        },
        {
            MethodName: "XRpcDouble",
            Handler: _Math_XRpcDouble_Handler,
        },
        {
            MethodName: "Add",
            Handler: _Math_Add_Handler,
        },
        {
            MethodName: "Double",
            Handler: _Math_Double_Handler,
        },
        {
            MethodName: "Calc",
            Handler: _Math_Calc_Handler,
        },
    },
    Streams: []xrpc.StreamDesc{},
    Metadata: "math",
}

