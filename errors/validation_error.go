package errors

import "net/http"

type ValidationError struct{}

func NewValidationError() *ValidationError {
	return &ValidationError{}
}

func (*ValidationError) Error() string {
	return "invalid input"
}

func (*ValidationError) Code() int {
	return http.StatusBadRequest
}
