package deliverystream

import (
	"fmt"

	"github.com/taiyoh/toyhose/errors"
)

func validateRange(key string, val, min, max int) errors.Raised {
	if val < min || max < val {
		return errors.NewInvalidArgumentException(fmt.Sprintf("%s is invalid", key))
	}
	return nil
}

func validateRangeUInt(key string, val, min, max uint) errors.Raised {
	if val < min || max < val {
		return errors.NewInvalidArgumentException(fmt.Sprintf("%s is invalid", key))
	}
	return nil
}
