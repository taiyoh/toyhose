package firehose

import (
	"time"

	"github.com/taiyoh/toyhose/datatypes/arn"
)

type DeliveryStream struct {
	ARN     arn.DeliveryStream
	Created time.Time
	Updated time.Time
	Status  StreamStatus
	Type    StreamType
	Version uint
}

func (d *DeliveryStream) clone() *DeliveryStream {
	return &DeliveryStream{
		ARN:     d.ARN,
		Created: d.Created,
		Updated: d.Updated,
		Status:  d.Status,
		Type:    d.Type,
		Version: d.Version,
	}
}

func (d *DeliveryStream) Active() *DeliveryStream {
	newDS := d.clone()
	newDS.Status = StatusActive
	newDS.Updated = time.Now()
	return newDS
}
