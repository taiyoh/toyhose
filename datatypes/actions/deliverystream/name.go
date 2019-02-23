package deliverystream

import (
	"regexp"

	"github.com/taiyoh/toyhose/exception"
)

var nameRE = regexp.MustCompile("[a-zA-Z0-9_.-]+")

func validateName(name string) exception.Raised {
	if name == "" {
		return exception.NewInvalidArgument("DeliveryStreamName")
	}
	if len(name) > 64 {
		return exception.NewInvalidArgument("DeliveryStreamName")
	}
	if !nameRE.MatchString(name) {
		return exception.NewInvalidArgument("DeliveryStreamName")
	}
	return nil
}
