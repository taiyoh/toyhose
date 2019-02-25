package deliverystream

import (
	"encoding/json"

	"github.com/taiyoh/toyhose/datatypes/arn"
	"github.com/taiyoh/toyhose/datatypes/firehose"
	"github.com/taiyoh/toyhose/datatypes/s3"
	"github.com/taiyoh/toyhose/exception"
)

type CreateInput struct {
	region    string
	accountID string
	Name      string   `json:"DeliveryStreamName"`
	Type      string   `json:"DeliveryStreamType"`
	S3Conf    *s3.Conf `json:"ExtendedS3DestinationConfiguration"`
}

func NewCreateInput(region, accountID string, arg []byte) (*CreateInput, exception.Raised) {
	input := CreateInput{region: region, accountID: accountID}
	if err := json.Unmarshal(arg, &input); err != nil {
		return nil, exception.NewInvalidArgument("input")
	}
	if err := input.Validate(); err != nil {
		return nil, err
	}
	if s3conf := input.S3Conf; s3conf != nil {
		s3conf.FillDefaultValue()
	}
	return &input, nil
}

func (i CreateInput) Validate() exception.Raised {
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

type CreateOutput struct {
	ARN string `json:"DeliveryStreamARN"`
}

func (i CreateInput) ARN() arn.DeliveryStream {
	return arn.NewDeliveryStream(i.region, i.accountID, i.Name)
}
