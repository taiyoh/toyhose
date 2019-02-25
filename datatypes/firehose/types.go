package firehose

import "errors"

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

func restoreStreamType(typ string) (StreamType, error) {
	t, exists := streamTypeMap[typ]
	if !exists {
		return TypeInvalid, errors.New("no stream type found")
	}
	return t, nil
}
