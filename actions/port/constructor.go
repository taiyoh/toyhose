package port

import "io"

// New returns input and output port object
func New(body io.ReadCloser) (*Input, *Output) {
	i := newInput(body)
	o := &Output{}
	return i, o
}
