package deliverystream

import (
	"encoding/json"

	"github.com/taiyoh/toyhose/datatypes/arn"
	"github.com/taiyoh/toyhose/datatypes/firehose"
	"github.com/taiyoh/toyhose/errors"
)

type ListInput struct {
	region             string
	accountID          string
	Type               string `json:"DeliveryStreamType"`
	ExclusiveStartName string `json:"ExclusiveStartDeliveryStreamName"`
	Limit              *uint  `json:"Limit"`
}

func NewListInput(region, accountID string, arg []byte) (*ListInput, errors.Raised) {
	input := ListInput{region: region, accountID: accountID}
	if err := json.Unmarshal(arg, &input); err != nil {
		return nil, errors.NewValidationError()
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}
	(&input).FillDefaultValue()
	return &input, nil
}

func (i ListInput) validateLimit() errors.Raised {
	lmPtr := i.Limit
	if lmPtr == nil {
		return nil
	}
	if *lmPtr > 10000 {
		return errors.NewInvalidParameterValue("Limit")
	}
	return nil
}

func (i ListInput) validateExclusiveStartName() errors.Raised {
	if i.ExclusiveStartName == "" {
		return nil
	}
	if len(i.ExclusiveStartName) > 64 {
		return errors.NewInvalidParameterValue("ExclusiveStartDeliveryStreamName")
	}
	if !nameRE.MatchString(i.ExclusiveStartName) {
		return errors.NewInvalidArgumentException("ExclusiveStartDeliveryStreamName")
	}
	return nil
}

func (i ListInput) Validate() errors.Raised {
	if _, err := firehose.RestoreStreamType(i.Type); err != nil {
		return err
	}
	if err := i.validateLimit(); err != nil {
		return err
	}
	if err := i.validateExclusiveStartName(); err != nil {
		return err
	}
	return nil
}

func (i *ListInput) FillDefaultValue() {
	if i.Limit == nil {
		ten := uint(10)
		i.Limit = &ten
	}
}

func (i ListInput) ARN() arn.DeliveryStream {
	name := i.ExclusiveStartName
	if name == "" {
		name = "*"
	}
	return arn.NewDeliveryStream(i.region, i.accountID, name)
}

type ListOutput struct {
	Names   []string `json:"DeliveryStreamNames"`
	HasNext bool     `json:"HasMoreDeliveryStreams"`
}
