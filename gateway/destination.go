package gateway

import (
	"context"
	"fmt"

	"github.com/taiyoh/toyhose/datatypes/firehose"
)

type destinationDriver interface {
	DispenceSequence(context.Context) uint32
	Save(context.Context, *firehose.Destination) error
}

type Destination struct {
	driver destinationDriver
}

func NewDestination(d destinationDriver) *Destination {
	return &Destination{d}
}

func (d *Destination) DispenseID(ctx context.Context) firehose.DestinationID {
	id := fmt.Sprintf("destinationId-%09d", d.driver.DispenceSequence(ctx))
	return firehose.DestinationID(id)
}

func (d *Destination) Save(ctx context.Context, dest *firehose.Destination) error {
	return d.driver.Save(ctx, dest)
}
