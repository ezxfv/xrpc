package xrpc

import (
	"errors"
	"reflect"
)

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
