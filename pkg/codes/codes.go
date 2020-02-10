package codes

import (
	"strings"
)

type Code uint

const (
	Unimplemented Code = iota

	Ok
	ServerError
	Unknown
)

func ErrorClass(err error) string {
	if err == nil {
		return "ok"
	}

	return "unknown"
}

func ErrorCode(err error) Code {
	if err == nil {
		return Ok
	}
	if strings.Contains(err.Error(), "server") {
		return ServerError
	}
	return Unknown
}
