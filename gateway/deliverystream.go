package gateway

import (
	"context"

	"github.com/taiyoh/toyhose/datatypes/arn"
	"github.com/taiyoh/toyhose/datatypes/firehose"
)

type streamDriver interface {
	Save(context.Context, *firehose.DeliveryStream) error
	Find(context.Context, arn.DeliveryStream) *firehose.DeliveryStream
	FindMulti(context.Context, arn.DeliveryStream, uint) ([]*firehose.DeliveryStream, bool)
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

func (d *DeliveryStream) Find(ctx context.Context, arn arn.DeliveryStream) *firehose.DeliveryStream {
	return d.driver.Find(ctx, arn)
}

func (d *DeliveryStream) FindMulti(ctx context.Context, arn arn.DeliveryStream, limit uint) ([]*firehose.DeliveryStream, bool) {
	return d.driver.FindMulti(ctx, arn, limit)
}
