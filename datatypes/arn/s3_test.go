package arn_test

import (
	"testing"

	"github.com/taiyoh/toyhose/datatypes/arn"
)

func TestS3(t *testing.T) {
	for _, tt := range []struct {
		label    string
		raw      string
		hasError bool
	}{
		{
			"another arn",
			"arn:aws:firehose:foo:bar:deliverystream/baz",
			true,
		},
		{
			"wrong bucket_name",
			"arn:aws:s3:::",
			true,
		},
		{
			"bucket_name",
			"arn:aws:s3:::foo",
			false,
		},
		{
			"bucket_name and key_name",
			"arn:aws:s3:::foo/bar",
			false,
		},
	} {
		a, err := arn.RestoreS3FromRaw(tt.raw)
		if tt.hasError {
			if err == nil {
				t.Errorf(`label="%s" msg="error not found"`, tt.label)
			}
			continue
		}
		if c := a.Code(); c != tt.raw {
			t.Errorf(`label="%s" msg="wrong ARN built" expected="%s" actual="%s"`, tt.label, tt.raw, c)
		}
	}
}
