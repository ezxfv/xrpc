package xrpc

import (
	"context"
	"encoding/json"
	"time"

	"x.io/xrpc/types"
)

func XBackground() *XContext {
	return &XContext{
		ctx:     context.Background(),
		values:  map[interface{}]interface{}{},
		cookies: map[string]string{},
	}
}

type XContext struct {
	ctx     context.Context
	values  map[interface{}]interface{}
	cookies map[string]string
}

func (xctx *XContext) Deadline() (deadline time.Time, ok bool) {
	return xctx.ctx.Deadline()
}

func (xctx *XContext) Done() <-chan struct{} {
	return xctx.ctx.Done()
}

func (xctx *XContext) Err() error {
	return xctx.ctx.Err()
}

func (xctx *XContext) Value(key interface{}) interface{} {
	if v, ok := xctx.values[key]; ok {
		return v
	}
	return xctx.ctx.Value(key)
}

func (xctx *XContext) Context() context.Context {
	ctx := context.Background()
	ctx = types.SetCookies(ctx, xctx.cookies)
	for k, v := range xctx.values {
		ctx = context.WithValue(ctx, k, v)
	}
	return ctx
}

func (xctx *XContext) SetCtx(ctx context.Context) {
	xctx.ctx = ctx
	xctx.cookies = types.FetchCookies(ctx)
}

func (xctx *XContext) WithValue(key, value interface{}) {
	xctx.values[key] = value
	xctx.ctx = context.WithValue(xctx.ctx, key, value)
}

func (xctx *XContext) SetCookie(key, value string) {
	xctx.cookies[key] = value
	xctx.ctx = types.SetCookie(xctx.ctx, key, value)
}

func (xctx *XContext) Cookie(key string) (cookie string) {
	if cookie, ok := xctx.cookies[key]; ok {
		return cookie
	}
	return types.GetCookie(xctx.ctx, key)
}

func (xctx *XContext) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"values":  xctx.values,
		"cookies": xctx.cookies,
	}
	return json.Marshal(m)
}

func (xctx *XContext) UnmarshalJSON(data []byte) (err error) {
	m := map[string]interface{}{}
	err = json.Unmarshal(data, &m)
	if cookies, ok := m["cookies"]; ok {
		xctx.cookies = cookies.(map[string]string)
		xctx.ctx = types.SetCookies(context.Background(), xctx.cookies)
	}
	if values, ok := m["values"]; ok {
		xctx.values = values.(map[interface{}]interface{})
		for k, v := range xctx.values {
			xctx.ctx = context.WithValue(xctx.ctx, k, v)
		}
	}
	return
}
