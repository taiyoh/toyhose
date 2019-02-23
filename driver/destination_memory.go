package driver

import (
	"context"

	"github.com/taiyoh/toyhose/datatypes/firehose"
)

type DestinationMemory struct {
	sequence uint32
	list     []*firehose.Destination
}

func NewDestinationMemory(seq uint32) *DestinationMemory {
	return &DestinationMemory{sequence: seq, list: []*firehose.Destination{}}
}

func (d *DestinationMemory) DispenceSequence(ctx context.Context) uint32 {
	d.sequence++
	return d.sequence
}

func (d *DestinationMemory) Save(ctx context.Context, dest *firehose.Destination) error {
	d.list = append(d.list, dest)
	return nil
}
