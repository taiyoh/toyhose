package errors

import (
	"fmt"
	"net/http"
)

type MissingParameter struct {
	field string
}

func NewMissingParameter(f string) *MissingParameter {
	return &MissingParameter{f}
}

func (p *MissingParameter) Error() string {
	return fmt.Sprintf("%s is required", p.field)
}

func (*MissingParameter) Code() int {
	return http.StatusBadRequest
}
