package deliverystream

import "github.com/taiyoh/toyhose/exception"

func validateType(t string) exception.Raised {
	if t == "" {
		return nil
	}
	if t != "DirectPut" {
		return exception.NewInvalidArgument("DeliveryStreamType")
	}
	return nil
}
