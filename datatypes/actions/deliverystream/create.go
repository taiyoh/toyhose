package deliverystream

import (
	"encoding/json"

	"github.com/taiyoh/toyhose/datatypes/s3"
	"github.com/taiyoh/toyhose/exception"
)

type CreateInput struct {
	Name   string   `json:"DeliveryStreamName"`
	Type   string   `json:"DeliveryStreamType"`
	S3Conf *s3.Conf `json:"ExtendedS3DestinationConfiguration"`
}

func NewCreateInput(arg []byte) (*CreateInput, error) {
	input := CreateInput{}
	if err := json.Unmarshal(arg, &input); err != nil {
		return nil, err
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}
	return &input, nil
}

func (i CreateInput) validateType() error {
	if i.Type == "" {
		return nil
	}
	if i.Type != "DirectPut" {
		return exception.NewInvalidArgument("DeliveryStreamType")
	}
	return nil
}

func (i CreateInput) Validate() error {
	if err := validateName(i.Name); err != nil {
		return err
	}
	if err := i.validateType(); err != nil {
		return err
	}
	if s3conf := i.S3Conf; s3conf != nil {
		if err := s3conf.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type CreateOutput struct {
	arn string `json:"DeliveryStreamARN"`
}
