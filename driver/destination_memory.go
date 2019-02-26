package driver

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/taiyoh/toyhose/datatypes/firehose"
)

// DestinationMemory provides destination memory storage
type DestinationMemory struct {
	sequence uint64
	list     []*firehose.Destination
	idmap    map[firehose.DestinationID]int
	mu       *sync.RWMutex
}

// NewDestinationMemory returns DestinationMemory object
func NewDestinationMemory() *DestinationMemory {
	return &DestinationMemory{
		sequence: 0,
		list:     []*firehose.Destination{},
		idmap:    map[firehose.DestinationID]int{},
		mu:       &sync.RWMutex{},
	}
}

// DispenceSequence returns sequence number for generating id
func (d *DestinationMemory) DispenceSequence(ctx context.Context) uint64 {
	atomic.AddUint64(&d.sequence, 1)
	return d.sequence
}

// Save provides set Destination object to this instance
func (d *DestinationMemory) Save(ctx context.Context, dest *firehose.Destination) error {
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
