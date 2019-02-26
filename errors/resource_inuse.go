package errors

import (
	"fmt"
	"net/http"
)

type ResourceInUseException struct {
	name string
}

func NewResourceInUse(name string) *ResourceInUseException {
	return &ResourceInUseException{name}
}

func (r *ResourceInUseException) Error() string {
	return fmt.Sprintf("ResouceName: %s is already in use", r.name)
}

func (*ResourceInUseException) Code() int {
	return http.StatusBadRequest
}
