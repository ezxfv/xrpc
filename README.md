# XRPC
A Simple RPC Framework

## Requirements
- 支持proto3(定制proto-gen-go)和go interface(基于go ast解析)生成桩代码，或者直接使用函数地址Call(reflect实现)，入参和返回值都是[]byte
- 传输类型(tcp, kcp, ws, quic) x (tls, multiple stream)
- 插件系统
  - jaeger分布式链路追踪
  - prometheus监控上报
  - 特定日志
  - 连接黑白名单
  - 连接认证

## TODO
- p2p服务发现(基于chord协议)
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
