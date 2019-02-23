package port

import "io"

func New(body io.ReadCloser) (*Input, *Output) {
	i := newInput(body)
	o := &Output{}
	return i, o
}
