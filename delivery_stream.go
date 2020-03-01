package toyhose

import (
	"context"
	"sync"
	"time"
)

type deliveryStream struct {
	arn       string
	source    chan []byte
	closer    context.CancelFunc
	s3Dest    *s3Destination
	createdAt time.Time
}

func (d *deliveryStream) Close() {
	d.closer()
	close(d.source)
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