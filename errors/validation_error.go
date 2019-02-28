package errors

import "net/http"

type ValidationError struct{}

func NewValidationError() *ValidationError {
	return &ValidationError{}
}

func (*ValidationError) Error() string {
	return "invalid input"
}

func (*ValidationError) code() int {
	return http.StatusBadRequest
}

func (v *ValidationError) Output() (int, []byte) {
	return v.code(), marshalOutput("ValidationError", v.Error())
}
