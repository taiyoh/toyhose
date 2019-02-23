package driver

import (
	"context"

	"github.com/taiyoh/toyhose/datatypes/arn"
	"github.com/taiyoh/toyhose/datatypes/firehose"
)

type DeliveryStreamMemory struct {
	streams  []*firehose.DeliveryStream
	arnIndex map[arn.DeliveryStream]int
}

func NewDeliveryStreamMemory() *DeliveryStreamMemory {
	return &DeliveryStreamMemory{
		streams:  []*firehose.DeliveryStream{},
		arnIndex: map[arn.DeliveryStream]int{},
	}
}

func (d *DeliveryStreamMemory) Save(ctx context.Context, ds *firehose.DeliveryStream) error {
	if i, exists := d.arnIndex[ds.ARN]; exists {
		d.streams[i] = ds
		return nil
	}
	d.streams = append(d.streams, ds)
	d.arnIndex[ds.ARN] = len(d.streams) - 1
	return nil
}
