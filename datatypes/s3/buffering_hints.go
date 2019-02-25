package s3

import (
	"github.com/taiyoh/toyhose/exception"
)

// https://docs.aws.amazon.com/ja_jp/firehose/latest/APIReference/API_BufferingHints.html

// BufferingHints describes hints for the buffering to perform before delivering data to the destination
type BufferingHints struct {
	IntervalInSeconds *uint `json:"IntervalInSeconds"` // default: 300
	SizeInMBs         *uint `json:"SizeInMBs"`         // delault: 5
}

func (h *BufferingHints) validateIntervalSeconds() exception.Raised {
	if sPtr := h.IntervalInSeconds; sPtr != nil && (*sPtr < 60 || 900 < *sPtr) {
		return exception.NewInvalidArgument("IntervalSeconds")
	}
	return nil
}

func (h *BufferingHints) validateSizeInMBs() exception.Raised {
	if bPtr := h.SizeInMBs; bPtr != nil && (*bPtr < 1 || 128 < *bPtr) {
		return exception.NewInvalidArgument("SizeInMBs")
	}
	return nil
}

var (
	defaultIntervalInSeconds uint = 300
	defaultSizeInMBs         uint = 5
)

// FillDefaultValue provides filling default value if not assigned
func (h *BufferingHints) FillDefaultValue() {
	if h.IntervalInSeconds == nil {
		h.IntervalInSeconds = &defaultIntervalInSeconds
	}
	if h.SizeInMBs == nil {
		h.SizeInMBs = &defaultSizeInMBs
	}
}

// Validate returns exception if each assigned value is invalid
func (h *BufferingHints) Validate() exception.Raised {
	if err := h.validateIntervalSeconds(); err != nil {
		return err
	}
	if err := h.validateSizeInMBs(); err != nil {
		return err
	}
	return nil
}
