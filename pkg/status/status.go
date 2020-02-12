package status

import (
	"errors"
	"fmt"

	"x.io/xrpc/pkg/codes"
)

func Error(c codes.Code, err string) error {
	return errors.New(fmt.Sprintf("[%v]%s", err))
}

// Errorf returns Error(c, fmt.Sprintf(format, a...)).
func Errorf(c codes.Code, format string, a ...interface{}) error {
	return Error(c, fmt.Sprintf(format, a...))
}
