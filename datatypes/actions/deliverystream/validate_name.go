package deliverystream

import (
	"regexp"

	"github.com/taiyoh/toyhose/errors"
)

var nameRE = regexp.MustCompile("^[a-zA-Z0-9_.-]+$")

func validateName(name string) errors.Raised {
	if name == "" {
		return errors.NewMissingParameter("DeliveryStreamName")
	}
	if len(name) > 64 {
		return errors.NewInvalidParameterValue("DeliveryStreamName")
	}
	if !nameRE.MatchString(name) {
		return errors.NewInvalidArgumentException("DeliveryStreamName")
	}
	return nil
}
