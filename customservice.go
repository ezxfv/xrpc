package xrpc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"

	"x.io/xrpc/pkg/encoding"
	_ "x.io/xrpc/pkg/encoding/json"
)

const (
	customPrefix = "custom"
)

// ServiceInfo service info.
type ServiceInfo struct {
	Name    string
	PkgPath string
	Methods []*MethodInfo
}

type UnaryHandler func(ctx context.Context, req interface{}) (interface{}, error)

type StdHandler func(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor UnaryServerInterceptor) (interface{}, error)

type CustomHandler func(srv interface{}, ctx context.Context)

type MethodInfo struct {
	Name      string
	ReqName   string
	Req       string
	ReplyName string
	Reply     string
}

func Call(f reflect.Value, params ...interface{}) (out []interface{}, err error) {
	if len(params) != f.Type().NumIn() {
		err = errors.New("the number of params is not adapted")
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		v := reflect.ValueOf(param)
		if f.Type().In(k).Kind() == reflect.Ptr {
			in[k] = v
		} else {
			if v.Kind() == reflect.Ptr {
				in[k] = v.Elem()
			} else {
				in[k] = v
			}
		}
	}
	resValues := f.Call(in)
	for _, rv := range resValues {
		out = append(out, rv.Interface())
	}
	return
}

type customService struct {
	name string
	ss   interface{}
	md   map[string]reflect.Value
}

func NewCustomServer() *CustomServer {
	return &CustomServer{
		pool:  &argsPool{pools: &sync.Pool{}},
		m:     map[string]*customService{},
		codec: encoding.GetCodec("json"),
		mu:    &sync.Mutex{},
	}
}

type CustomServer struct {
	pool  *argsPool
	m     map[string]*customService
	codec encoding.Codec
	mu    *sync.Mutex
}

func (r *CustomServer) RegisterCustomService(serviceName string, ss interface{}) (err error) {
	sv := reflect.ValueOf(ss)
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.m[serviceName]; ok {
		log.Fatalf("Server.RegisterCustomService found duplicate service registration for %q", serviceName)
	}
	serviceName = customPrefix + "." + serviceName
	srv := &customService{
		name: serviceName,
		ss:   ss,
		md:   make(map[string]reflect.Value),
	}
	for i := 0; i < sv.Type().NumMethod(); i++ {
		method := sv.Type().Method(i)
		srv.md[method.Name] = sv.Method(i)
	}
	r.m[serviceName] = srv
	return
}

func (r *CustomServer) RegisterFunction(serviceName string, fname string, fn interface{}) {
	serviceName = customPrefix + "." + serviceName
	if _, ok := r.m[serviceName]; !ok {
		r.m[serviceName] = &customService{
			name: serviceName,
			ss:   nil,
			md:   map[string]reflect.Value{},
		}
	}
	r.m[serviceName].md[fname] = reflect.ValueOf(fn)
}

func (r *CustomServer) RpcCall(ctx context.Context, service, method string, dec func(interface{}) error, interceptor UnaryServerInterceptor) (reply interface{}, err error) {
	srv, knownService := r.m[service]
	if knownService {
		if md, ok := srv.md[method]; ok {
			ins, _, hasCtx, ok := r.pool.GenArgsForFunc(md)
			if !ok {
				return nil, errors.New("gen args failed")
			}
			if hasCtx {
				ins = ins[1:]
			}
			err = dec(&ins)
			if err != nil {
				return
			}
			if hasCtx {
				xctx := XBackground()
				xctx.SetCtx(ctx)
				ins = append([]interface{}{xctx}, ins[:]...)
			}
			if interceptor == nil {
				reply, err = Call(md, ins...)
			} else {
				info := &UnaryServerInfo{
					Server:     srv,
					FullMethod: fmt.Sprintf("/%s/%s", service, method),
				}
				handler := func(ctx context.Context, req interface{}) (interface{}, error) {
					return Call(md, ins...)
				}
				reply, err = interceptor(ctx, ins, info, handler)
			}
		}
	}
	return
}

func (r *CustomServer) DirectCall(method string, data []byte) (result interface{}, err error) {
	info := strings.Split(method, ".")
	if len(info) != 2 {
		err = errors.New("invalid method string")
		return
	}
	srv, knownService := r.m[info[0]]
	if knownService {
		if md, ok := srv.md[info[1]]; ok {
			ins, _, _, ok := r.pool.GenArgsForFunc(md)
			if !ok {
				println("gen args failed")
				return nil, nil
			}
			err = r.codec.Unmarshal(data, &ins)
			if err != nil {
				fmt.Println(err.Error())
			}
			result, err = Call(md, ins...)
			return
		}
	}
	return
}
