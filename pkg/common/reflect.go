package common

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

type Context struct {
	Service string
	Method  string
	Values  map[string]interface{}
}

type service struct {
	name string
	ss   interface{}
	md   map[string]reflect.Value
}

func NewReflectDemo() *ReflectDemo {
	return &ReflectDemo{
		m:   map[string]*service{},
		tag: "demo",
		mu:  &sync.Mutex{},
	}
}

type ReflectDemo struct {
	m   map[string]*service
	tag string
	mu  *sync.Mutex
}

func (r *ReflectDemo) RegisterService(serviceName string, ss interface{}) (err error) {
	sv := reflect.ValueOf(ss)
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.m[serviceName]; ok {
		log.Fatalf("Server.RegisterCustomService found duplicate service registration for %q", serviceName)
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

func (r *ReflectDemo) Call(method string, params ...interface{}) (result []interface{}, err error) {
	info := strings.Split(method, ".")
	if len(info) != 2 {
		err = errors.New("invalid method string")
		return
	}
	srv, knownService := r.m[info[0]]
	if knownService {
		if md, ok := srv.md[info[1]]; ok {
			result, err = _call(md, params...)
			return
		}
	}
	return
}

func _call(f reflect.Value, params ...interface{}) (out []interface{}, err error) {
	if len(params) != f.Type().NumIn() {
		err = errors.New("the number of params is not adapted")
		return
	}

	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	resValues := f.Call(in)
	for _, rv := range resValues {
		out = append(out, rv.Interface())
	}
	return
}

func (r *ReflectDemo) Marshal(i interface{}) (bs []byte, err error) {
	v := reflect.ValueOf(i)
	val := v.Elem()
	w := bytes.NewBuffer([]byte{})
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		tag, ok := typeField.Tag.Lookup(r.tag)
		if !ok {
			err = errors.New("can't find tag key:" + r.tag)
			return
		}
		str := fmt.Sprintf("%s:`%v`", tag, valueField.Interface())
		w.Write([]byte(str + ","))
	}
	bs = w.Bytes()
	bs = bs[:len(bs)-1]
	return
}

func (r *ReflectDemo) Unmarshal(bs []byte, i interface{}) (err error) {
	bs = bs[:len(bs)-1]
	fs := strings.Split(string(bs), "`,")
	ps := reflect.ValueOf(i)
	s := ps.Elem()

	var ok bool
	var tagVal string
	if s.Kind() == reflect.Struct {
		for _, f := range fs {
			ff := strings.Split(f, ":`")
			if len(ff) != 2 {
				err = errors.New("parse bs failed")
				return
			}
			name := ff[0]
			val := ff[1]
			for j := 0; j < s.NumField(); j++ {
				tf := s.Type().Field(j)
				vf := s.Field(j)
				tagVal, ok = tf.Tag.Lookup(r.tag)
				if !ok {
					err = errors.New("can't find tag key:" + r.tag)
					return
				}
				if tagVal == name {
					if !vf.CanSet() {
						err = errors.New("can't set filed:" + tf.Name)
						return
					}
					switch vf.Kind() {
					case reflect.String:
						vf.SetString(val)
					case reflect.Int:
						intVal, _ := strconv.Atoi(val)
						if !vf.OverflowInt(int64(intVal)) {
							vf.SetInt(int64(intVal))
						}
					}
				}
			}
		}
	}
	return
}
