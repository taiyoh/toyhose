package deliverystream

import (
	"encoding/json"

	"github.com/taiyoh/toyhose/datatypes/firehose"
	"github.com/taiyoh/toyhose/datatypes/s3"
	"github.com/taiyoh/toyhose/errors"
)

// CreateInput provides input resource for creating delivery stream usecase
type CreateInput struct {
	Name   string   `json:"DeliveryStreamName"`
	Type   string   `json:"DeliveryStreamType"`
	S3Conf *s3.Conf `json:"ExtendedS3DestinationConfiguration"`
}

// NewCreateInput provides constructor for CreateInput object
func NewCreateInput(arg []byte) (*CreateInput, errors.Raised) {
	input := CreateInput{}
	if err := json.Unmarshal(arg, &input); err != nil {
		return nil, errors.NewValidationError()
	}
	if err := input.validate(); err != nil {
		return nil, err
	}
	if s3conf := input.S3Conf; s3conf != nil {
		s3conf.FillDefaultValue()
	}
	return &input, nil
}

func (i CreateInput) validate() errors.Raised {
	if err := validateName(i.Name); err != nil {
		return err
	}
	if _, err := firehose.RestoreStreamType(i.Type); err != nil {
		return err
	}
	if s3conf := i.S3Conf; s3conf != nil {
		if err := s3conf.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// CreateOutput provides output resource for creating delivery stream usecase
type CreateOutput struct {
	ARN string `json:"DeliveryStreamARN"`
}
