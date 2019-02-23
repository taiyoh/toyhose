package actions

import (
	"context"

	"github.com/taiyoh/toyhose/datatypes/arn"
	"github.com/taiyoh/toyhose/datatypes/firehose"
)

type DeliveryStreamRepository interface {
	Save(context.Context, *firehose.DeliveryStream) error
	Find(context.Context, arn.DeliveryStream) *firehose.DeliveryStream
	FindMulti(ctx context.Context, start arn.DeliveryStream, limit uint) ([]*firehose.DeliveryStream, bool)
}

type DestinationRepository interface {
	DispenseID(context.Context) firehose.DestinationID
	Save(context.Context, *firehose.Destination) error
}
