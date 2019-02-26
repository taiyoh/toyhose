package errors

import (
	"fmt"
	"net/http"
)

type InvalidParameterValue struct {
	field string
}

func NewInvalidParameterValue(f string) *InvalidParameterValue {
	return &InvalidParameterValue{f}
}

func (i *InvalidParameterValue) Error() string {
	return fmt.Sprintf("%s value is out of range", i.field)
}

func (*InvalidParameterValue) Code() int {
	return http.StatusBadRequest
}
