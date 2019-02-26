package port

import (
	"bytes"
	"context"
	"io"
)

// Input provides request wrapper for each reqest
type Input struct {
	ctx context.Context
	arg []byte
}

func newInput(body io.ReadCloser) *Input {
	b := bytes.NewBuffer([]byte{})
	b.ReadFrom(body)
	return &Input{
		ctx: context.Background(),
		arg: b.Bytes(),
	}
}

// Ctx returns context object
func (i *Input) Ctx() context.Context {
	return i.ctx
}

// Arg returns input bytes
func (i *Input) Arg() []byte {
	return i.arg
}
