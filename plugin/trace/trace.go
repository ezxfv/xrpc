package trace

import (
	"context"
	"encoding/json"
	"io"

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

func New(isServer ...bool) *tracePlugin {
	server := true
	if len(isServer) > 0 {
		server = isServer[0]
	}
	t := &tracePlugin{
		server:     server,
		tracer:     opentracing.GlobalTracer(),
		otgrpcOpts: newOptions(),
	}
	return t
}

type tracePlugin struct {
	server bool

	tracer     opentracing.Tracer
	otgrpcOpts *options
}

func (t *tracePlugin) Intercept(ctx context.Context, req interface{}, info *types.UnaryServerInfo, handler types.UnaryHandler) (resp interface{}, err error) {
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
		ctx = types.SetCookie(ctx, k, v)
	}

	resp, err = handler(ctx, req)

	SetSpanTags(span, err, t.server)
	if err == nil {
		if t.otgrpcOpts.logPayloads {
			span.LogFields(log.Object("xrpc response", resp))
		}
	} else {
		span.LogFields(log.String("event", "error"), log.String("message", err.Error()))
	}
	if t.otgrpcOpts.decorator != nil {
		t.otgrpcOpts.decorator(span, info.FullMethod, req, resp, err)
	}
	return resp, err
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
			vv = types.GetCookie(m.ctx, k)
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
func SetSpanTags(span opentracing.Span, err error, server bool) {
	code := codes.ErrorCode(err)
	span.SetTag("response_code", code)
	if err == nil {
		return
	}
	if !server || code == codes.ServerError {
		ext.Error.Set(span, true)
	}
}
