package model

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"

	"github.com/edenzhong7/xrpc/pkg/encoding"
	_ "github.com/edenzhong7/xrpc/pkg/encoding/json"

	"google.golang.org/grpc"
)

type (
	UnaryServerInfo        = grpc.UnaryServerInfo
	UnaryServerInterceptor = grpc.UnaryServerInterceptor
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

// MethodInfo method info
type MethodInfo struct {
	Name      string
	ReqName   string
	Req       string
	ReplyName string
	Reply     string
}

func _call(f reflect.Value, params ...interface{}) (out []interface{}, err error) {
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
			in[k] = v.Elem()
		}
	}
	resValues := f.Call(in)
	for _, rv := range resValues {
		out = append(out, rv.Interface())
	}
	return
}

type service struct {
	name string
	ss   interface{}
	md   map[string]reflect.Value
}

func NewCustomService() *CustomService {
	return &CustomService{
		pool:  &argsPool{pools: &sync.Pool{}},
		m:     map[string]*service{},
		codec: encoding.GetCodec("json"),
		mu:    &sync.Mutex{},
	}
}

type CustomService struct {
	pool  *argsPool
	m     map[string]*service
	codec encoding.Codec
	mu    *sync.Mutex
}

func (r *CustomService) RegisterService(serviceName string, ss interface{}) (err error) {
	sv := reflect.ValueOf(ss)
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.m[serviceName]; ok {
		log.Fatalf("Server.RegisterService found duplicate service registration for %q", serviceName)
	}
	srv := &service{
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

func (r *CustomService) RegisterFunction(serviceName string, fname string, fn interface{}) {
	if _, ok := r.m[serviceName]; !ok {
		r.m[serviceName] = &service{
			name: serviceName,
			ss:   nil,
			md:   map[string]reflect.Value{},
		}
	}
	r.m[serviceName].md[fname] = reflect.ValueOf(fn)
}

func (r *CustomService) Call(method string, data []byte) (result []interface{}, err error) {
	info := strings.Split(method, ".")
	if len(info) != 2 {
		err = errors.New("invalid method string")
		return
	}
	srv, knownService := r.m[info[0]]
	if knownService {
		if md, ok := srv.md[info[1]]; ok {
			ins, _, ok := r.pool.GenArgsForFunc(md)
			if !ok {
				println("gen args failed")
				return nil, nil
			}
			err = r.codec.Unmarshal(data, &ins)
			if err != nil {
				fmt.Println(err.Error())
			}
			result, err = _call(md, ins...)
			return
		}
	}
	return
}
