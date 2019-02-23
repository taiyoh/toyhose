package firehose

type StreamType int

const (
	TypeInvalid StreamType = iota
	TypeDirectPut
)
