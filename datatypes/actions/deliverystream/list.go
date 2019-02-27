package deliverystream

import (
	"encoding/json"

	"github.com/taiyoh/toyhose/errors"
)

// ListInput provides input resource for listing delivery stream usecase
type ListInput struct {
	Type               *string `json:"DeliveryStreamType"`
	ExclusiveStartName *string `json:"ExclusiveStartDeliveryStreamName"`
	Limit              *uint   `json:"Limit"`
}

// NewListInput provides constructor for ListInput object
func NewListInput(arg []byte) (*ListInput, errors.Raised) {
	input := ListInput{}
	if err := json.Unmarshal(arg, &input); err != nil {
		return nil, errors.NewValidationError()
	}
	if err := input.validate(); err != nil {
		return nil, err
	}
	(&input).fillDefaultValue()
	return &input, nil
}

func (i ListInput) validateLimit() errors.Raised {
	lmPtr := i.Limit
	if lmPtr == nil {
		return nil
	}
	if *lmPtr < 1 || 10000 < *lmPtr {
		return errors.NewInvalidParameterValue("Limit")
	}
	return nil
}

func (i ListInput) validateExclusiveStartName() errors.Raised {
	exName := i.ExclusiveStartName
	if exName == nil {
		return nil
	}
	if len(*exName) > 64 {
		return errors.NewInvalidParameterValue("ExclusiveStartDeliveryStreamName")
	}
	if !nameRE.MatchString(*exName) {
		return errors.NewInvalidArgumentException("ExclusiveStartDeliveryStreamName")
	}
	return nil
}

func (i ListInput) validate() errors.Raised {
	if i.Type != nil {
		if err := validateType(*i.Type); err != nil {
			return err
		}
	}
	if err := i.validateLimit(); err != nil {
		return err
	}
	if err := i.validateExclusiveStartName(); err != nil {
		return err
	}
	return nil
}

var defaultListLimit uint = 10

func (i *ListInput) fillDefaultValue() {
	if i.Limit == nil {
		i.Limit = &defaultListLimit
	}
}

// ExclusiveStartDeliveryStreamName returns cursor for start position of listing
func (i ListInput) ExclusiveStartDeliveryStreamName() string {
	name := i.ExclusiveStartName
	if name == nil {
		return "*"
	}
	return *name
}

// ListOutput provides output resource for listing delivery stream usecase
type ListOutput struct {
	Names   []string `json:"DeliveryStreamNames"`
	HasNext bool     `json:"HasMoreDeliveryStreams"`
}
