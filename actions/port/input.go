package port

import (
	"bytes"
	"context"
	"io"
)

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

func (i *Input) Ctx() context.Context {
	return i.ctx
}

func (i *Input) Arg() []byte {
	return i.arg
}
