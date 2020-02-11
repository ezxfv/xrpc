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
  - 加密数据
  - 服务注册

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
        whitelistPlugin := whitelist.New(map[string]bool{"127.0.0.1": true}, nil)
        s.ApplyPlugins(promPlugin, logPlugin, tracePlugin, whitelistPlugin)
        if enableCrypto {
            cryptoPlugin := crypto.New()
            cryptoPlugin.SetKey(sessionID, sessionKey)
            s.ApplyPlugins(cryptoPlugin)
        }
        s.StartPlugins()
    }
    if enableAuth {
        admin := xrpc.NewAdminAuthenticator(user, pass)
        s.SetAuthenticator(admin)
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

    cryptoPlugin := crypto.New()
    cryptoPlugin.SetKey(sessionID, sessionKey)
    conn.ApplyPlugins(cryptoPlugin)

    // 设置auth和crypto参数
    conn.SetHeaderArg("user", user)
    conn.SetHeaderArg("pass", pass)
    conn.SetHeaderArg(crypto.Key, sessionID)
    
    client := pb.NewMathClient(conn)
    ctx :=context.Background()

    r := client.Add(ctx, a, b)
    assert.Equal(t, r, 5)
    log.Printf("3 + 2 = %d", r)
    sum, avg := client.Calc(ctx, a, b)
    assert.Equal(t, 5, sum)
    assert.Equal(t, 2.50, avg)
    log.Printf("sum = %d, avg = %.2f", sum, avg)
    val, n := client.Inc(ctx, n)
    assert.Equal(t, val, n.Val)
    assert.Equal(t, int32(20), val)
    log.Printf("new num: %d", n.Val)
}
```

### 直接调用服务
#### CustomServer
```go
var (
	sessionID  = "session_math_0"
	sessionKey = "1234"

	user = "admin"
	pass = "1234"
	ctx  = context.Background()
)

func setupConn(conn *xrpc.ClientConn) {
	cryptoPlugin := crypto.New()
	cryptoPlugin.SetKey(sessionID, sessionKey)
	conn.ApplyPlugins(cryptoPlugin)

	conn.SetHeaderArg("user", user)
	conn.SetHeaderArg("pass", pass)
	conn.SetHeaderArg(crypto.Key, sessionID)
}

func TestCustomServer(t *testing.T) {
	lis, s := newServer("tcp", "localhost:9898")
	s.RegisterCustomService("math", &Math{})
	s.RegisterFunction("default", "Double", func(a int) int {
		return a * 2
	})
	s.RegisterFunction("default", "RpcDouble", func(c *xrpc.XContext, a int) int {
		client, err := xrpc.NewRawClient("tcp", "localhost:9898", xrpc.WithJsonCodec())
		if err != nil {
			return 0
		}
		client.Setup(setupConn)

		var r int
		err = client.RawCall(c, "default.Double", &r, a)
		return r
	})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	s.Start()
}
```
#### Client
```go
func TestCustomClientTrace(t *testing.T) {
	client, err := xrpc.NewRawClient("tcp", "localhost:9898", xrpc.WithJsonCodec())
	assert.Equal(t, nil, err)
	client.Setup(setupConn)
	var c int
	xctx := xrpc.XBackground()
	err = client.RawCall(xctx, "default.RpcDouble", &c, 10)
	assert.Equal(t, nil, err)
	assert.Equal(t, 20, c)
}
```

RpcDouble中新建client调用了实际的Double计算，测试多级rpc调用时trace能正确记录
