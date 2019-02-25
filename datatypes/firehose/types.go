package firehose

// StreamType provides type of delivery stream
type StreamType int

const (
	// TypeInvalid is unknown status of delivery stram
	TypeInvalid StreamType = iota
	// TypeDirectPut provides that Provider applications access the delivery stream directly.
	TypeDirectPut
)
