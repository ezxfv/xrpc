package common

import (
	"errors"
	"reflect"
)

func Call(fn interface{}, params ...interface{}) (out []interface{}, err error) {
	f := reflect.ValueOf(fn)
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
