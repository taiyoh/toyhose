package firehose

import (
	"github.com/taiyoh/toyhose/datatypes/arn"
	"github.com/taiyoh/toyhose/datatypes/s3"
)

// DestinationID provides identification of Destination
type DestinationID string

// Destination provides destination for a delivery stream
type Destination struct {
	ID        DestinationID
	SourceARN arn.DeliveryStream
	S3Conf    *s3.Conf
}
