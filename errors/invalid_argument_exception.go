package errors

import (
	"net/http"
)

type InvalidArgumentException struct {
	msg string
}

func NewInvalidArgumentException(msg string) *InvalidArgumentException {
	return &InvalidArgumentException{msg}
}

func (e *InvalidArgumentException) Error() string {
	return e.msg
}

func (*InvalidArgumentException) code() int {
	return http.StatusBadRequest
}

func (e *InvalidArgumentException) Output() (int, []byte) {
	return e.code(), marshalOutput("InvalidArgumentException", e.msg)
}
