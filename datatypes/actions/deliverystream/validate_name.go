package deliverystream

import (
	"regexp"

	"github.com/taiyoh/toyhose/errors"
)

var nameRE = regexp.MustCompile("^[a-zA-Z0-9_.-]+$")

func validateName(name string) errors.Raised {
	if name == "" {
		return errors.NewInvalidArgumentException("DeliveryStreamName is required")
	}
	if len(name) > 64 {
		return errors.NewInvalidArgumentException("DeliveryStreamName value length is over")
	}
	if !nameRE.MatchString(name) {
		return errors.NewInvalidArgumentException("DeliveryStreamName value is invalid format")
	}
	return nil
}
