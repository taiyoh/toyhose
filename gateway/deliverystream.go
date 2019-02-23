package gateway

import (
	"context"

	"github.com/taiyoh/toyhose/datatypes/firehose"
)

type streamDriver interface {
	Save(context.Context, *firehose.DeliveryStream) error
}

type DeliveryStream struct {
	driver streamDriver
}

func NewDeliveryStream(driver streamDriver) *DeliveryStream {
	return &DeliveryStream{driver}
}

func (d *DeliveryStream) Save(ctx context.Context, ds *firehose.DeliveryStream) error {
	return d.driver.Save(ctx, ds)
}
