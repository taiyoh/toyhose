package exception

import (
	"fmt"
	"net/http"
)

type ResourceInUse struct {
	name string
}

func NewResourceInUse(name string) *ResourceInUse {
	return &ResourceInUse{name}
}

func (r *ResourceInUse) Error() string {
	return fmt.Sprintf("ResouceName: %s is already in use", r.name)
}

func (r *ResourceInUse) Code() int {
	return http.StatusBadRequest
}
