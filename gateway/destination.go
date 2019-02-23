package gateway

import (
	"fmt"

	"github.com/taiyoh/toyhose/datatypes/firehose"
)

type destinationDriver interface {
	DispenceSequence() uint32
	Save(*firehose.Destination) error
}

type Destination struct {
	driver destinationDriver
}

func NewDestination(d destinationDriver) *Destination {
	return &Destination{d}
}

func (d *Destination) DispenseID() firehose.DestinationID {
	id := fmt.Sprintf("destinationId-%09d", d.driver.DispenceSequence())
	return firehose.DestinationID(id)
}

func (d *Destination) Save(dest *firehose.Destination) error {
	return d.driver.Save(dest)
}
