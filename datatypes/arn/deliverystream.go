package arn

import "fmt"

// arn:aws:firehose:region:account-id:deliverystream/delivery-stream-name

// DeliveryStream provides ARN builder for firehose deliverystream
type DeliveryStream struct {
	region     string
	accountID  string
	streamName string
}

// NewDeliveryStream returns arn object
func NewDeliveryStream(region, accountID, name string) DeliveryStream {
	return DeliveryStream{region, accountID, name}
}

// Code returns arn as string
func (d DeliveryStream) Code() string {
	return fmt.Sprintf("arn:aws:firehose:%s:%s:deliverystream/%s", d.region, d.accountID, d.streamName)
}

// Name returns streamname
func (d DeliveryStream) Name() string {
	return d.streamName
}

// CompareLevel provides compare status for arn comparison
type CompareLevel int

const (
	// CompareNotEqual provides that each ARNs have no same elements
	CompareNotEqual CompareLevel = iota
	// CompareEqualAccount provides that each ARNs are only same account
	CompareEqualAccount
	// CompareEqualRegionAccount provides that each ARNs are same region and account
	CompareEqualRegionAccount
	// CompareEqualAll provides that all elements are same
	CompareEqualAll
)

// Compare returns CompareLevel from each ARNs
func (d DeliveryStream) Compare(s DeliveryStream) CompareLevel {
	if d.accountID != s.accountID {
		return CompareNotEqual
	}
	if d.region != s.region {
		return CompareEqualAccount
	}
	if d.streamName != s.streamName {
		return CompareEqualRegionAccount
	}
	return CompareEqualAll
}
