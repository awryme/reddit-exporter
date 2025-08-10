package render

import (
	"errors"
	"net/http"
)

var DefaultErrCode = http.StatusInternalServerError

type errorWithCode struct {
	err  error
	code int
}

func (e errorWithCode) Error() string {
	return e.err.Error()
}

func (e errorWithCode) Unwrap() error {
	return e.err
}

func ErrorWithCode(err error, code int) error {
	return errorWithCode{err, code}
}

func GetCode(err error) int {
	var e errorWithCode
	if errors.As(err, &e) {
		return e.code
	}
	return DefaultErrCode
}
