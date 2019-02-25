package firehose

import "github.com/taiyoh/toyhose/exception"

// StreamType provides type of delivery stream
type StreamType int

const (
	// TypeInvalid is unknown status of delivery stram
	TypeInvalid StreamType = iota
	// TypeDirectPut provides that Provider applications access the delivery stream directly.
	TypeDirectPut
)

var streamTypeMap = map[string]StreamType{
	"DirectPut": TypeDirectPut,
}

// RestoreStreamType returns detected StreamType, also returns exception if it is invalid
func RestoreStreamType(typ string) (StreamType, exception.Raised) {
	t, exists := streamTypeMap[typ]
	if !exists {
		return TypeInvalid, exception.NewInvalidArgument("DeliveryStreamType")
	}
	return t, nil
}
