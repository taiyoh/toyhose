package firehose

import (
	"github.com/taiyoh/toyhose/datatypes/s3"
)

type DestinationID string

type Destination struct {
	ID     DestinationID
	S3Conf *s3.Conf
}
