package deliverystream

import (
	"errors"
	"regexp"
)

var nameRE = regexp.MustCompile("[a-zA-Z0-9_.-]+")

func validateName(name string) error {
	if name == "" {
		return errors.New("DeliveryStreamName is required")
	}
	if len(name) > 64 {
		return errors.New("DeliveryStreamName length is over")
	}
	if !nameRE.MatchString(name) {
		return errors.New("DeliveryStreamName pattern unmatched")
	}
	return nil
}
