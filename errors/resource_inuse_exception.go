package errors

import (
	"net/http"
)

type ResourceInUseException struct {
	msg string
}

func NewResourceInUse(msg string) *ResourceInUseException {
	return &ResourceInUseException{msg}
}

func (r *ResourceInUseException) Error() string {
	return r.msg
}

func (*ResourceInUseException) code() int {
	return http.StatusBadRequest
}

func (r *ResourceInUseException) Output() (int, []byte) {
	return r.code(), marshalOutput("ResourceInUseException", r.msg)
}
