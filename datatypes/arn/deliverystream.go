package arn

import "fmt"

// arn:aws:firehose:region:account-id:deliverystream/delivery-stream-name

type DeliveryStream struct {
	region     string
	accountID  string
	streamName string
}

func (d DeliveryStream) Code() string {
	return fmt.Sprintf("arn:aws:firehose:%s:%s:deliverystream/%s", d.region, d.accountID, d.streamName)
}
