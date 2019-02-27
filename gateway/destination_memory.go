package gateway

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/taiyoh/toyhose/datatypes/firehose"
)

// Destination provides destination memory storage
type Destination struct {
	sequence uint64
	list     []*firehose.Destination
	idmap    map[firehose.DestinationID]int
	mu       *sync.RWMutex
}

// NewDestination returns DestinationMemory object
func NewDestination() *Destination {
	return &Destination{
		sequence: 0,
		list:     []*firehose.Destination{},
		idmap:    map[firehose.DestinationID]int{},
		mu:       &sync.RWMutex{},
	}
}

func (d *Destination) dispenseSequence() uint64 {
	atomic.AddUint64(&d.sequence, 1)
	return d.sequence
}

// DispenseID returns generated id for destination
func (d *Destination) DispenseID(ctx context.Context) firehose.DestinationID {
	id := fmt.Sprintf("destinationId-%09d", d.dispenseSequence())
	return firehose.DestinationID(id)
}

// Save provides set Destination object to this instance
func (d *Destination) Save(ctx context.Context, dest *firehose.Destination) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if i, exists := d.idmap[dest.ID]; exists {
		d.list[i] = dest
		return nil
	}
	d.list = append(d.list, dest)
	d.idmap[dest.ID] = len(d.list) - 1
	return nil
}
