package firehose

// Error provides domain layer's error definitions
type Error string

// Error returns error message
func (e Error) Error() string {
	return string(e)
}

var (
	// ErrRequired provides supplied field is required
	ErrRequired = Error("required")
	// ErrNotFound provides supplied field is out of collection
	ErrNotFound = Error("not found")
)
