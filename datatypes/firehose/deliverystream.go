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
