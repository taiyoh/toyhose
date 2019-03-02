package firehose

import (
	"time"

	"github.com/taiyoh/toyhose/datatypes/arn"
)

// DeliveryStream provides domain object of firehose deliverystream
type DeliveryStream struct {
	ARN     arn.DeliveryStream
	Created time.Time
	Updated time.Time
	Status  StreamStatus
	Type    StreamType
	Version uint
}

// NewDeliveryStream returns DeliveryStream object
func NewDeliveryStream(a arn.DeliveryStream, typ string) (*DeliveryStream, error) {
	t, err := RestoreStreamType(typ)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	return &DeliveryStream{
		ARN:     a,
		Created: now,
		Updated: now,
		Status:  StatusCreating,
		Type:    t,
		Version: 1,
	}, nil
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

// StreamName returns delivery stream name as string
func (d *DeliveryStream) StreamName() string {
	return d.ARN.Name()
}

// Active provides make status active and update Updated field
func (d *DeliveryStream) Active() *DeliveryStream {
	newDS := d.clone()
	newDS.Status = StatusActive
	newDS.Updated = time.Now()
	return newDS
}
