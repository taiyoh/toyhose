package toyhose

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/google/uuid"
)

type deliveryStream struct {
	arn                string
	deliveryStreamName string
	deliveryStreamType string
	recordCh           chan *deliveryRecord
	closer             context.CancelFunc
	destDesc           *firehose.DestinationDescription
	sourceDesc         *firehose.SourceDescription
	createdAt          time.Time
}

func (d *deliveryStream) Close() {
	d.closer()
	close(d.recordCh)
}

type deliveryRecord struct {
	id   string
	data []byte
}

func newDeliveryRecord(data []byte) *deliveryRecord {
	return &deliveryRecord{
		id:   uuid.New().String(),
		data: data,
	}
}

type deliveryStreamPool struct {
	mutex sync.RWMutex
	pool  map[string]*deliveryStream
}

func (p *deliveryStreamPool) Add(d *deliveryStream) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.pool[d.arn] = d
}

func (p *deliveryStreamPool) Find(arn string) *deliveryStream {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	if ds, ok := p.pool[arn]; ok {
		return ds
	}
	return nil
}

func (p *deliveryStreamPool) Delete(arn string) *deliveryStream {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if ds, ok := p.pool[arn]; ok {
		delete(p.pool, arn)
		return ds
	}
	return nil
}

func (p *deliveryStreamPool) FindAllBySource(streamType string, from *string, limit *int64) ([]*deliveryStream, bool) {
	pickups := make([]*deliveryStream, 0, len(p.pool))
	for _, ds := range p.pool {
		if ds.deliveryStreamType != streamType {
			continue
		}
		pickups = append(pickups, ds)
	}
	sort.Slice(pickups, func(i, j int) bool {
		return pickups[i].createdAt.Before(pickups[j].createdAt)
	})

	if from != nil {
		idx := 0
		for i, v := range pickups {
			if v.deliveryStreamName == *from {
				idx = i + 1
				break
			}
		}
		pickups = pickups[idx:]
	}

	hasNext := false

	if limit != nil {
		if lim := int(*limit); len(pickups) > lim {
			hasNext = true
			pickups = pickups[:lim]
		}
	}
	return pickups, hasNext
}
