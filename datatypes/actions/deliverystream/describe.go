package deliverystream

import (
	"errors"
	"time"

	"github.com/taiyoh/toyhose/datatypes/s3"
)

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

type Destination struct {
	ID     string   `json:"DestinationId"`
	S3Conf *s3.Conf `json:"ExtendedS3DestinationDescription"`
}

type DescribeOutput struct {
	Created         time.Time      `json:"CreatedTimeStamp"`
	Updated         time.Time      `json:"UpdatedTimeStamp"`
	ARN             string         `json:"DeliveryStreamARN"`
	Name            string         `json:"DeliveryStreamName"`
	Status          string         `json:"DeliveryStreamStatus"`
	Type            string         `json:"DeliveryStreamType"`
	Destinations    []*Destination `json:"Destinations"`
	MoreDestination bool           `json:"HasMoreDestinations"`
	VersionID       string         `json:"VersionId"`
}
