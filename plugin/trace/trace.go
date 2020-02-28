package trace

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"x.io/xrpc"
	"x.io/xrpc/pkg/codes"
	"x.io/xrpc/types"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
)

const (
	Name    = "trace"
	SpanKey = "server_span"
)

var (
	SpanKeys = []string{"uber-trace-id"}
	c        io.Closer
)

func init() {
	c = registerJaeger()
}

func New() *tracePlugin {
	t := &tracePlugin{
		tracer:     opentracing.GlobalTracer(),
		otgrpcOpts: newOptions(),
	}
	return t
}

type tracePlugin struct {
	tracer     opentracing.Tracer
	otgrpcOpts *options
}

func (t *tracePlugin) PreHandle(ctx context.Context, req interface{}, info *types.UnaryServerInfo) (context.Context, error) {
	spanContext, err := t.tracer.Extract(opentracing.TextMap, NewSpanCtxReader(ctx))
	var span opentracing.Span
	if err != nil {
		span = t.tracer.StartSpan(
			info.FullMethod,
		)
	} else {
		span = t.tracer.StartSpan(
			info.FullMethod,
			ext.RPCServerOption(spanContext),
		)
	}
	if t.otgrpcOpts.inclusionFunc != nil &&
		!t.otgrpcOpts.inclusionFunc(spanContext, info.FullMethod, req, nil) {
		return ctx, nil
	}

	span.SetTag("xrpc", "dev")
	ctx = opentracing.ContextWithSpan(ctx, span)
	if t.otgrpcOpts.logPayloads {
		span.LogFields(log.Object("xrpc request", req))
	}
	ctx = context.WithValue(ctx, SpanKey, span)
	w := NewSpanCtxWriter()
	err = t.tracer.Inject(span.Context(), opentracing.TextMap, w)
	if err != nil {
		return ctx, err
	}
	for k, v := range w.M {
		ctx = xrpc.SetCookie(ctx, k, v)
	}
	return ctx, nil
}

func (t *tracePlugin) PostHandle(ctx context.Context, req interface{}, resp interface{}, info *types.UnaryServerInfo, err error) (context.Context, error) {
	serverSpan, ok := ctx.Value(SpanKey).(opentracing.Span)
	if !ok {
		return ctx, errors.New("trace plugin get server_span failed")
	}
	defer serverSpan.Finish()
	SetSpanTags(serverSpan, err, false)
	if err == nil {
		if t.otgrpcOpts.logPayloads {
			serverSpan.LogFields(log.Object("xrpc response", resp))
		}
	} else {
		serverSpan.LogFields(log.String("event", "error"), log.String("message", err.Error()))
	}
	if t.otgrpcOpts.decorator != nil {
		t.otgrpcOpts.decorator(serverSpan, info.FullMethod, req, resp, err)
	}
	return ctx, nil
}

func (t *tracePlugin) Stop() error {
	if c == nil {
		return nil
	}
	return c.Close()
}

func NewSpanCtxReader(ctx context.Context) opentracing.TextMapReader {
	return &SpanCtxReader{ctx}
}

type SpanCtxReader struct {
	ctx context.Context
}

func (m *SpanCtxReader) ForeachKey(handler func(key, val string) error) (err error) {
	for _, k := range SpanKeys {
		v := m.ctx.Value(k)
		vv := ""
		if v != nil {
			if sv, ok := v.(string); ok {
				vv = sv
			}
		} else {
			vv = xrpc.GetCookie(m.ctx, k)
		}
		err = handler(k, vv)
		if err != nil {
			return err
		}
	}
	return
}

func NewSpanCtxWriter() *SpanCtxWriter {
	return &SpanCtxWriter{M: map[string]string{}}
}

type SpanCtxWriter struct {
	M map[string]string
}

func (m *SpanCtxWriter) Set(key, val string) {
	m.M[key] = val
}

func (m *SpanCtxWriter) String() string {
	d, _ := json.Marshal(m.M)
	return string(d)
}

// SetSpanTags sets one or more tags on the given span according to the error.
func SetSpanTags(span opentracing.Span, err error, client bool) {
	code := codes.ErrorCode(err)
	span.SetTag("response_code", code)
	if err == nil {
		return
	}
	if client || code == codes.ServerError {
		ext.Error.Set(span, true)
	}
}
