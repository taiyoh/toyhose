package deliverystream

import (
	"github.com/taiyoh/toyhose/datatypes/firehose"
	"github.com/taiyoh/toyhose/errors"
)

func validateType(typ string) errors.Raised {
	_, err := firehose.RestoreStreamType(typ)
	if err == nil {
		return nil
	}
	switch err.(firehose.Error) {
	case firehose.ErrRequired:
		return errors.NewInvalidArgumentException("DeliveryStreamType is required")
	default:
		return errors.NewInvalidArgumentException("DeliveryStreamType value is invalid format")
	}
}
