package arn_test

import (
	"testing"

	"github.com/taiyoh/toyhose/datatypes/arn"
)

func TestIAMRole(t *testing.T) {
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
			"invalid semicolon counts",
			"arn:aws:iam::account-id:role:role-name",
			true,
		},
		{
			"no role prefix",
			"arn:aws:iam::account-id:role-name",
			true,
		},
		{
			"correct",
			"arn:aws:iam::account-id:role/role-name",
			false,
		},
	} {
		a, err := arn.RestoreIAMRoleFromRaw(tt.raw)
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
