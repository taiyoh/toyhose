package actions

import (
	"context"

	"github.com/taiyoh/toyhose/datatypes/firehose"
)

type DeliveryStreamRepository interface {
	Save(context.Context, *firehose.DeliveryStream) error
}

type DestinationRepository interface {
	DispenseID(context.Context) firehose.DestinationID
	Save(context.Context, *firehose.Destination) error
}
