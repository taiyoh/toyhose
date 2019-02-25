package firehose

// StreamStatus provides status collection for a delivery stream
type StreamStatus int

const (
	// StatusInvalid is unknown delivrey stram status
	StatusInvalid StreamStatus = iota
	// StatusCreating is initial status of delivery stram
	StatusCreating
	// StatusDeleting is destructing status of delivery stram
	StatusDeleting
	// StatusActive provides that delivery stream can send to destination
	StatusActive
)

// String returns status as string
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
