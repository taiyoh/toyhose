package deliverystream

import "errors"

type DescribeInput struct {
	Name                        string  `json:"DeliveryStreamName"`
	ExclusiveStartDestinationID *string `json:"ExclusiveStartDestinationId"`
	Limit                       *uint   `json:"Limit"`
}

func (i DescribeInput) Validate() error {
	if err := validateName(i.Name); err != nil {
		return err
	}
	if idPtr := i.ExclusiveStartDestinationID; idPtr != nil {
		if l := len(*idPtr); l < 1 || 100 < l {
			return errors.New("ExclusiveStartDestinationID is invalid")
		}
	}
	if limPtr := i.Limit; limPtr != nil {
		l := *limPtr
		if l < 1 || 10000 < l {
			return errors.New("Limit is invalid")
		}
	}
	return nil
}

type DescribeOutput struct {
}
