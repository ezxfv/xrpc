package trace_test

import (
	"fmt"
	"io"
	"testing"
	"time"

	"context"

	"github.com/edenzhong7/xrpc"
	"github.com/edenzhong7/xrpc/plugin/trace"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	jaeger "github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

func initJaeger(service string) (opentracing.Tracer, io.Closer) {
	cfg := &config.Configuration{
		ServiceName: service,
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: "127.0.0.1:6831",
		},
	}
	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	return tracer, closer
}

func foo3(req string, ctx context.Context) (reply string) {
	//1.创建子span
	span, _ := opentracing.StartSpanFromContext(ctx, "span_foo3")
	defer func() {
		//4.接口调用完，在tag中设置request和reply
		span.SetTag("request", req)
		span.SetTag("reply", reply)
		span.Finish()
	}()

	println(req)
	//2.模拟处理耗时
	time.Sleep(time.Second / 2)
	//3.返回reply
	reply = "foo3Reply"
	return
}

//跟foo3一样逻辑
func foo4(req string, ctx context.Context) (reply string) {
	span, _ := opentracing.StartSpanFromContext(ctx, "span_foo4")
	defer func() {
		span.SetTag("request", req)
		span.SetTag("reply", reply)
		span.Finish()
	}()

	println(req)
	time.Sleep(time.Second / 2)
	reply = "foo4Reply"
	return
}

func TestJaeger(t *testing.T) {
	tracer, closer := initJaeger("jaeger-demo")
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer) //StartspanFromContext创建新span时会用到

	span := tracer.StartSpan("span_root")
	ctx := opentracing.ContextWithSpan(context.Background(), span)
	r1 := foo3("Hello foo3", ctx)
	r2 := foo4("Hello foo4", ctx)
	fmt.Println(r1, r2)
	span.Finish()
}

func TestJaegerTwoStep(t *testing.T) {
	tracer, closer := initJaeger("jaeger-demo")
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer) //StartspanFromContext创建新span时会用到

	span1 := tracer.StartSpan("span_root")
	ctx1 := opentracing.ContextWithSpan(context.Background(), span1)
	foo3("Hi foo3", ctx1)
	w := trace.NewSpanCtxWriter()
	tracer.Inject(span1.Context(), opentracing.TextMap, w)
	span1.Finish()

	ctx2 := context.Background()
	for k, v := range w.M {
		ctx2 = context.WithValue(ctx2, k, v)
	}
	r := trace.NewSpanCtxReader(ctx2)
	spanCtx, _ := tracer.Extract(opentracing.TextMap, r)
	span2 := tracer.StartSpan(
		"foo4",
		ext.RPCServerOption(spanCtx),
	)
	span3 := tracer.StartSpan(
		"foo4.1",
		ext.RPCServerOption(span2.Context()),
	)
	ctx2 = opentracing.ContextWithSpan(context.Background(), span3)
	foo4("Hi foo4", ctx2)
	span2.Finish()
	span3.Finish()
}

func TestTracePlugin(t *testing.T) {
	p := trace.New()
	tracer := opentracing.GlobalTracer()
	//defer p.Stop()
	for i := 0; i < 1; i++ {
		span := tracer.StartSpan(fmt.Sprintf("trace_root_%d", i))
		ctx := opentracing.ContextWithSpan(context.Background(), span)

		w := trace.NewSpanCtxWriter()
		tracer.Inject(span.Context(), opentracing.TextMap, w)
		for k, v := range w.M {
			ctx = context.WithValue(ctx, k, v)
		}

		info := &xrpc.UnaryServerInfo{FullMethod: "test_trace_plugin"}
		ctx, err := p.PreHandle(ctx, nil, info)
		time.Sleep(time.Millisecond * 200)
		p.PostHandle(ctx, nil, nil, info, err)
		span.Finish()
	}
	// 需要delay刷新数据，或者手动Close刷掉缓存
	time.Sleep(time.Second)
}
