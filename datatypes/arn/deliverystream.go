package arn

import "fmt"

// arn:aws:firehose:region:account-id:deliverystream/delivery-stream-name

type DeliveryStream struct {
	region     string
	accountID  string
	streamName string
}

func NewDeliveryStream(region, accountID, name string) DeliveryStream {
	return DeliveryStream{region, accountID, name}
}

func (d DeliveryStream) Code() string {
	return fmt.Sprintf("arn:aws:firehose:%s:%s:deliverystream/%s", d.region, d.accountID, d.streamName)
}

func (d DeliveryStream) Name() string {
	return d.streamName
}

type CompareLevel int

const (
	CompareNotEqual CompareLevel = iota
	CompareEqualAccount
	CompareEqualRegionAccount
	CompareEqualAll
)

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
