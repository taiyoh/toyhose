package errors

type Raised interface {
	Error() string
	Code() int
}
