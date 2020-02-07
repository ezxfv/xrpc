# XRPC
A Simple RPC Framework

## Requirements
- 支持protobuf + struct + function三种方式注册服务
- 传输类型(tcp, kcp, ws, quic) + (tls, multiple stream)
- 上层中间件(调用对应的handler function之前执行)
- 底层插件(插件初始化/销毁， 连接建立/断开(如交换秘钥,协商压缩算法)+消息收/发后处理(加解密，解压缩))
- 基于protoc-gen-go, 可直接使用proto3文件
- 完整的单元测试+性能测试
- jaeger分布式链路追踪
- prometheus监控上报
- 服务发现，负载均衡，连接认证
- p2p
- 多语言客户端(本地tcp连client agent->server)，主要是要修改对应的proto-gen-xxx

## Example
### 直接使用接口定义MathService
- 编写math.go定义服务接口
```go
package math

type Num struct {
	Val int32
	S   Step
}

type Step struct {
	S int32
}

type Counter interface {
	Inc(n *Num) (int32, *Num)
	Dec(n Num) *Num
}

type Math interface {
	Counter
	Add(a, b int) int
	Calc(ints ...int) (int, float64)
}
```

- 使用parser生成[math.stub.go](protocol/math/math.stub.go), 实现一个MathServer
```go
type MathImpl struct {
	math_pb.UnimplementedMath
}

func (m *MathImpl) Inc(n *math_pb.Num) (int32, *math_pb.Num) {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
	n.Val += n.S.S
	return n.Val, n
}

func (m *MathImpl) Add(a, b int) int {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
	return a + b
}

func (m *MathImpl) Calc(ns ...int) (int, float64) {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
	var sum int
	for _, i := range ns {
		sum += i
	}
	return sum, float64(sum) / float64(len(ns))
}
```

- 启动MathServer
```go
func StartMathServer(protocol, addr string) {
	lis, err := net.Listen(context.Background(), protocol, addr)
	if err != nil {
		log.Fatal(err)
	}
	s := xrpc.NewServer()
	if enablePlugin {
		promPlugin := prom.New()
		logPlugin := logp.New()
		tracePlugin := trace.New()
		s.ApplyPlugins(promPlugin, logPlugin, tracePlugin)
	}
	math_pb.RegisterMathServer(s, &MathImpl{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	s.Start()
}
```

- 创建MathClient
```go
func RunMathClient() {
    conn, err := xrpc.Dial("tcp", "localhost:9898", xrpc.WithInsecure(), xrpc.WithJsonCodec())
    if err != nil {
        log.Fatalf("did not connect: %v", err)
    }
    defer conn.Close()
    client := pb.NewMathClient(conn)

    r := client.Add(context.Background(), a, b)
    assert.Equal(t, r, 5)
    log.Printf("3 + 2 = %d", r)
    sum, avg := client.Calc(context.Background(), a, b)
    assert.Equal(t, 5, sum)
    assert.Equal(t, 2.50, avg)
    log.Printf("sum = %d, avg = %.2f", sum, avg)
    val, n := client.Inc(context.Background(), n)
    assert.Equal(t, val, n.Val)
    assert.Equal(t, int32(20), val)
    log.Printf("new num: %d", n.Val)
}
```