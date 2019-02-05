package common

import "errors"

// https://docs.aws.amazon.com/ja_jp/firehose/latest/APIReference/API_BufferingHints.html

type BufferingHints struct {
	IntervalInSeconds uint `json:"IntervalInSeconds"` // default: 300
	SizeInMBs         uint `json:"SizeInMBs"`         // delault: 5
}

func (h *BufferingHints) validateIntervalSeconds() error {
	if h.IntervalInSeconds < 60 || 900 < h.IntervalInSeconds {
		return errors.New("IntervalSeconds is invalid")
	}
	return nil
}

func (h *BufferingHints) validateSizeInMBs() error {
	if h.SizeInMBs < 1 || 128 < h.SizeInMBs {
		return errors.New("SizeInMBs is invalid")
	}
	return nil
}

func (h *BufferingHints) FillDefaultValue() {
	if h.IntervalInSeconds == 0 {
		h.IntervalInSeconds = 300
	}
	if h.SizeInMBs == 0 {
		h.SizeInMBs = 5
	}
}

func (h *BufferingHints) Validate() error {
	return nil
}
