package firehose

type StreamStatus int

const (
	_ StreamStatus = iota
	StatusCreating
	StatusDeleting
	StatusActive
)

func (s StreamStatus) String() string {
	switch s {
	case StatusCreating:
		return "CREATING"
	case StatusDeleting:
		return "DELETING"
	case StatusActive:
		return "ACTIVE"
	default:
		panic("invalid status")
	}
}
