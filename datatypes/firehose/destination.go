package firehose

import (
	"github.com/taiyoh/toyhose/datatypes/arn"
	"github.com/taiyoh/toyhose/datatypes/s3"
)

type DestinationID string

type Destination struct {
	ID        DestinationID
	SourceARN arn.DeliveryStream
	S3Conf    *s3.Conf
}
