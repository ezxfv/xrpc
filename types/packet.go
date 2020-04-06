package types

import (
	"context"
	"encoding/binary"
	"encoding/json"
)

const (
	CookieKey = "xcookies"
)

func NewPacket() *Packet {
	return &Packet{
		cookies:  map[string]string{},
		payload:  nil,
		encoding: "",
	}
}

type Packet struct {
	cookies  map[string]string
	payload  []byte
	encoding string
}

func (sp *Packet) Write(payload []byte) {
	sp.payload = append(sp.payload, payload...)
}

func (sp *Packet) Set(payload []byte) {
	sp.payload = payload
}

func (sp *Packet) Cookies() map[string]string {
	cs := make(map[string]string)
	for k, v := range sp.cookies {
		cs[k] = v
	}
	return cs
}

func (sp *Packet) Payload() []byte {
	return sp.payload
}

func (sp *Packet) Encoding() string {
	return sp.encoding
}

func (sp *Packet) ReSet() {
	sp.cookies = make(map[string]string)
	sp.payload = make([]byte, 1024)
	sp.encoding = ""
}

func (sp *Packet) Fork() *Packet {
	f := &Packet{}
	f.encoding = sp.encoding
	f.payload = make([]byte, 1024)
	f.cookies = sp.Cookies()
	return f
}

func (sp *Packet) Marshal() ([]byte, error) {
	ss := struct {
		Cookies  map[string]string
		Payload  []byte
		Encoding string
	}{sp.cookies, sp.payload, sp.encoding}
	data, err := json.Marshal(ss)
	return data, err
}

func (sp *Packet) Unmarshal(data []byte) error {
	ss := struct {
		Cookies  map[string]string
		Payload  []byte
		Encoding string
	}{}
	err := json.Unmarshal(data, &ss)
	sp.cookies = ss.Cookies
	sp.payload = ss.Payload
	sp.encoding = ss.Encoding
	return err
}

func ParseCookies(ctx context.Context, cookies []byte) context.Context {
	m := map[string]string{}
	err := json.Unmarshal(cookies, &m)
	if err != nil {
		return ctx
	}
	return SetCookies(ctx, m)
}

func ReadCookiesHeader(ctx context.Context, data []byte) (context.Context, int) {
	if len(data) < cookieLen {
		return ctx, 0
	}
	length := binary.BigEndian.Uint32(data[:cookieLen])
	if length == 0 {
		return ctx, 0
	}
	l := int(cookieLen + length)
	if len(data) < l {
		return ctx, 0
	}
	cookies := data[cookieLen:l]
	return ParseCookies(ctx, cookies), l
}

func CookiesHeader(ctx context.Context) []byte {
	hdr := make([]byte, cookieLen, cookieLen)
	cookies := FetchCookies(ctx)
	cookiesData, err := json.Marshal(cookies)
	if err != nil {
		return hdr
	}
	binary.BigEndian.PutUint32(hdr[:], uint32(len(cookiesData)))
	return append(hdr, cookiesData...)
}

func SetCookies(ctx context.Context, cookies map[string]string) context.Context {
	for k, v := range cookies {
		ctx = SetCookie(ctx, k, v)
	}
	return ctx
}

func SetCookie(ctx context.Context, k, v string) context.Context {
	var cs map[string]string
	var ok bool
	if cs, ok = ctx.Value(CookieKey).(map[string]string); !ok {
		cs = make(map[string]string)
	}
	cs[k] = v
	ctx = context.WithValue(ctx, CookieKey, cs)
	return ctx
}

func FetchCookies(ctx context.Context) map[string]string {
	cs := make(map[string]string)
	var ok bool
	if ctx.Value(CookieKey) == nil {
		return cs
	}
	if cs, ok = ctx.Value(CookieKey).(map[string]string); !ok {
		return cs
	}
	return cs
}

func GetCookie(ctx context.Context, key string) string {
	var cs map[string]string
	var ok bool
	if cs, ok = ctx.Value(CookieKey).(map[string]string); !ok {
		return ""
	}
	return cs[key]
}
