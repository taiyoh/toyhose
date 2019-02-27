package gateway

import (
	"context"
	"sync"

	"github.com/taiyoh/toyhose/datatypes/arn"
	"github.com/taiyoh/toyhose/datatypes/firehose"
)

// DeliveryStream provides delivery stream memory storage
type DeliveryStream struct {
	streams  []*firehose.DeliveryStream
	arnIndex map[arn.DeliveryStream]int
	mu       *sync.RWMutex
}

// NewDeliveryStream returns DeliveryStream object
func NewDeliveryStream() *DeliveryStream {
	return &DeliveryStream{
		streams:  []*firehose.DeliveryStream{},
		arnIndex: map[arn.DeliveryStream]int{},
		mu:       &sync.RWMutex{},
	}
}

// Save provides set delivery stream object to this instance
func (d *DeliveryStream) Save(ctx context.Context, ds *firehose.DeliveryStream) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if i, exists := d.arnIndex[ds.ARN]; exists {
		d.streams[i] = ds
		return nil
	}
	d.streams = append(d.streams, ds)
	d.arnIndex[ds.ARN] = len(d.streams) - 1
	return nil
}

// Find returns delivery stream object
func (d *DeliveryStream) Find(ctx context.Context, a arn.DeliveryStream) *firehose.DeliveryStream {
	d.mu.RLock()
	defer d.mu.RUnlock()
	i, exists := d.arnIndex[a]
	if !exists {
		return nil
	}
	return d.streams[i]
}

func (d *DeliveryStream) findIndex(a arn.DeliveryStream) int {
	if a.Name() == "*" {
		return 0
	}
	if i, exists := d.arnIndex[a]; exists {
		return i + 1
	}
	return -1
}

func (d *DeliveryStream) calcRetrieveCount(idx, lim int) (int, bool) {
	if l := len(d.streams); l-1 <= idx+lim {
		return (l - idx), false
	}
	return lim, true
}

// FindMulti returns delivery stream list by supplied ARN
func (d *DeliveryStream) FindMulti(ctx context.Context, a arn.DeliveryStream, limit uint) ([]*firehose.DeliveryStream, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	streams := []*firehose.DeliveryStream{}
	startIdx := d.findIndex(a)
	if startIdx == -1 {
		return streams, false
	}
	retrieveCount, hasNext := d.calcRetrieveCount(startIdx, int(limit))

	seek := startIdx
	for i := 0; i < retrieveCount; i++ {
		st := d.streams[seek]
		streams = append(streams, st)
		seek++
	}

	return streams, hasNext
}
