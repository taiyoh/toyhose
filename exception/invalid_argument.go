package exception

import "fmt"

type InvalidArgument struct {
	field string
}

func NewInvalidArgument(field string) *InvalidArgument {
	return &InvalidArgument{field}
}

func (e *InvalidArgument) Error() string {
	return fmt.Sprintf("%s is invalid", e.field)
}
