package gateway

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/taiyoh/toyhose/datatypes/arn"
	"github.com/taiyoh/toyhose/datatypes/firehose"
)

// Destination provides destination memory storage
type Destination struct {
	sequence  uint64
	list      []*firehose.Destination
	idmap     map[firehose.DestinationID]int
	sourcemap map[arn.DeliveryStream][]int
	mu        *sync.RWMutex
}

// NewDestination returns DestinationMemory object
func NewDestination() *Destination {
	return &Destination{
		sequence:  0,
		list:      []*firehose.Destination{},
		idmap:     map[firehose.DestinationID]int{},
		sourcemap: map[arn.DeliveryStream][]int{},
		mu:        &sync.RWMutex{},
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

func (d *Destination) addToList(dst *firehose.Destination) int {
	d.list = append(d.list, dst)
	idx := len(d.list) - 1
	d.idmap[dst.ID] = idx
	return idx
}

func (d *Destination) addToSource(s arn.DeliveryStream, idx int) {
	if lst, ok := d.sourcemap[s]; ok {
		d.sourcemap[s] = append(lst, idx)
		return
	}
	d.sourcemap[s] = []int{idx}
}

// Save provides set Destination object to this instance
func (d *Destination) Save(ctx context.Context, dest *firehose.Destination) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if i, exists := d.idmap[dest.ID]; exists {
		d.list[i] = dest
		return nil
	}
	idx := d.addToList(dest)
	d.addToSource(dest.SourceARN, idx)

	return nil
}

// FindBySource returns destination list by supplied source
func (d *Destination) FindBySource(ctx context.Context, source arn.DeliveryStream) []*firehose.Destination {
	dests := []*firehose.Destination{}
	lst, ok := d.sourcemap[source]
	if !ok {
		return dests
	}
	for i := range lst {
		dests = append(dests, d.list[i])
	}
	return dests
}
