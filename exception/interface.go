package exception

type Raised interface {
	Error() string
	Code() int
}
