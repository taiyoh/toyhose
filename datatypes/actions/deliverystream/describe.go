package deliverystream

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/taiyoh/toyhose/datatypes/s3"
	"github.com/taiyoh/toyhose/errors"
)

// DescribeInput provides input resource for output delivery stream definition.
type DescribeInput struct {
	Name                        string  `json:"DeliveryStreamName"`
	ExclusiveStartDestinationID *string `json:"ExclusiveStartDestinationId"`
	Limit                       *uint   `json:"Limit"`
}

// NewDescribeInput provides constructor for DescribeInput object.
func NewDescribeInput(arg []byte) (*DescribeInput, errors.Raised) {
	input := DescribeInput{}
	if err := json.Unmarshal(arg, &input); err != nil {
		return nil, errors.NewValidationError()
	}
	if err := input.validate(); err != nil {
		return nil, err
	}

	return &input, nil
}

func (i DescribeInput) validate() errors.Raised {
	if err := validateName(i.Name); err != nil {
		return err
	}
	if idPtr := i.ExclusiveStartDestinationID; idPtr != nil {
		if err := validateRange("ExclusiveStartSestinationID", len(*idPtr), 1, 100); err != nil {
			return err
		}
	}
	if limPtr := i.Limit; limPtr != nil {
		if err := validateRangeUInt("Limit", *limPtr, 1, 10000); err != nil {
			return nil
		}
	}
	return nil
}

type Destination struct {
	ID     string   `json:"DestinationId"`
	S3Conf *s3.Conf `json:"S3DestinationDescription"`
}

// DeliveryStreamTime represents timestamp with milliseconds.
type DeliveryStreamTime struct {
	time time.Time
}

func NewDeliveryStreamTime(t time.Time) DeliveryStreamTime {
	return DeliveryStreamTime{t}
}

// MarshalJSON provides transition from time.Time object to timestamp string with milliseconds.
func (t DeliveryStreamTime) MarshalJSON() ([]byte, error) {
	unixMilli := t.time.UnixNano() / 1000000
	unixSec := unixMilli / 1000
	milliDiff := unixMilli - (unixSec * 1000)
	return []byte(fmt.Sprintf("%d.%d", unixSec, milliDiff)), nil
}

// DescribeOutput provides output resource for describing delivery stream usecase.
type DescribeOutput struct {
	Created         DeliveryStreamTime `json:"CreatedTimeStamp"`
	Updated         DeliveryStreamTime `json:"UpdatedTimeStamp"`
	ARN             string             `json:"DeliveryStreamARN"`
	Name            string             `json:"DeliveryStreamName"`
	Status          string             `json:"DeliveryStreamStatus"`
	Type            string             `json:"DeliveryStreamType"`
	Destinations    []*Destination     `json:"Destinations"`
	MoreDestination bool               `json:"HasMoreDestinations"`
	VersionID       string             `json:"VersionId"`
}
