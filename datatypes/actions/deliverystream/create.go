package deliverystream

import (
	"errors"
	"regexp"

	"github.com/taiyoh/toyhose/datatypes/s3"
)

type CreateInput struct {
	Name   string   `json:"DeliveryStreamName"`
	Type   string   `json:"DeliveryStreamType"`
	S3Conf *s3.Conf `json:"ExtendedS3DestinationConfiguration"`
}

var nameRE = regexp.MustCompile("[a-zA-Z0-9_.-]+")

func (i CreateInput) validateName() error {
	name := i.Name
	if name == "" {
		return errors.New("DeliveryStreamName is required")
	}
	if len(name) > 64 {
		return errors.New("DeliveryStreamName length is over")
	}
	if !nameRE.MatchString(name) {
		return errors.New("DeliveryStreamName pattern unmatched")
	}
	return nil
}

func (i CreateInput) validateType() error {
	if i.Type == "" {
		return nil
	}
	if i.Type != "DirectPut" && i.Type != "KinesisStreamAsSource" {
		return errors.New("DeliveryStreamType is invalid")
	}
	return nil
}

func (i CreateInput) Validate() error {
	if err := i.validateName(); err != nil {
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
