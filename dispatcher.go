package toyhose

import (
	"bytes"
	"io"

	"github.com/taiyoh/toyhose/actions"
)

type Dispatcher struct {
	repo   actions.DeliveryStreamRepository
	reader io.ReadCloser
}

type UseCaseFn func([]byte)

func NewDispatcher(repo actions.DeliveryStreamRepository, argreader io.ReadCloser) *Dispatcher {
	return &Dispatcher{repo, argreader}
}

func (d *Dispatcher) Dispatch(target string) UseCaseFn {
	a := actions.NewDeliveryStream(d.repo)
	switch FindType(target) {
	case CreateDeliveryStream:
		return a.Create
	case DeleteDeliveryStream:
		return a.Delete
	case DescribeDeliveryStream:
		return a.Describe
	case ListDeliveryStreams:
		return a.List
	case PutRecord:
		return a.PutRecord
	case PutRecordBatch:
		return a.PutRecordBatch
	}
	return nil
}

func (d *Dispatcher) Arg() []byte {
	b := bytes.NewBuffer([]byte{})
	b.ReadFrom(d.reader)
	return b.Bytes()
}
