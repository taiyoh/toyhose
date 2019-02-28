package errors

type Raised interface {
	Error() string
	Output() (int, []byte)
}
