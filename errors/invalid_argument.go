package errors

import (
	"fmt"
	"net/http"
)

type InvalidArgumentException struct {
	field string
}

func NewInvalidArgumentException(field string) *InvalidArgumentException {
	return &InvalidArgumentException{field}
}

func (e *InvalidArgumentException) Error() string {
	return fmt.Sprintf("%s is invalid", e.field)
}

func (*InvalidArgumentException) Code() int {
	return http.StatusBadRequest
}
