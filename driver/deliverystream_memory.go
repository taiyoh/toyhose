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

func (d *DeliveryStreamMemory) Find(ctx context.Context, a arn.DeliveryStream) *firehose.DeliveryStream {
	i, exists := d.arnIndex[a]
	if !exists {
		return nil
	}
	return d.streams[i]
}

type searchable struct {
	arn     arn.DeliveryStream
	enabled bool
}

func (s *searchable) Check(a arn.DeliveryStream) bool {
	res := s.arn.Compare(a)
	if res != arn.CompareEqualAll && res != arn.CompareEqualRegionAccount {
		return false
	}
	if !s.enabled && res == arn.CompareEqualAll {
		s.enabled = true
	}
	return s.enabled
}

func (d *DeliveryStreamMemory) FindMulti(ctx context.Context, a arn.DeliveryStream, limit uint) ([]*firehose.DeliveryStream, bool) {
	streams := []*firehose.DeliveryStream{}
	search := &searchable{a, a.Name() == "*"}
	hasNext := false
	for _, st := range d.streams {
		if !search.Check(st.ARN) {
			continue
		}
		if uint(len(streams)) >= limit {
			hasNext = true
			break
		}
		streams = append(streams, st)
	}
	return streams, hasNext
}
