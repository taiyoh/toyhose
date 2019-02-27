package deliverystream_test

import (
	"testing"

	"github.com/taiyoh/toyhose/datatypes/actions/deliverystream"
)

func TestCreateInput(t *testing.T) {
	for _, tt := range []struct {
		label    string
		rawjson  string
		hasError bool
	}{
		{
			"broken json",
			`{}[`,
			true,
		},
		{
			"empty name",
			`{}`,
			true,
		},
		{
			"too long name",
			`{"DeliveryStreamName":"12345678901234567890123456789012345678901234567890123456789012345"}`,
			true,
		},
		{
			"invalid charactor is in name",
			`{"DeliveryStreamName":"foobar[]***$"}`,
			true,
		},
		{
			"missing StreamType",
			`{"DeliveryStreamName":"foobar"}`,
			true,
		},
		{
			"invalid StreamType",
			`{"DeliveryStreamName":"foobar","DeliveryStreamType":"hoge"}`,
			true,
		},
		{
			"invalid S3Conf",
			`{"DeliveryStreamName":"foobar","DeliveryStreamType":"DirectPut","S3DestinationConfiguration":{}}`,
			true,
		},
		{
			"valid",
			`{"DeliveryStreamName":"foobar","DeliveryStreamType":"DirectPut","S3DestinationConfiguration":{"BucketARN":"arn:aws:s3:::bucket_name","RoleARN":"arn:aws:iam::accoun_id:role/role_name"}}`,
			false,
		},
	} {
		ci, err := deliverystream.NewCreateInput([]byte(tt.rawjson))
		if (ci == nil) != tt.hasError {
			t.Errorf(`label="%s" msg="CreateInput object returns"`, tt.label)
		}
		if (err != nil) != tt.hasError {
			t.Errorf(`label="%s" msg="error should returns"`, tt.label)
		}
	}
}
