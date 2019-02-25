package arn_test

import (
	"testing"

	"github.com/taiyoh/toyhose/datatypes/arn"
)

func TestDeliveryStream(t *testing.T) {
	a := arn.NewDeliveryStream("foo", "bar", "baz")
	t.Run("ARN test", func(t *testing.T) {
		if n := a.Name(); n != "baz" {
			t.Error("wrong name returns: ", n)
		}
		if c := a.Code(); c != "arn:aws:firehose:foo:bar:deliverystream/baz" {
			t.Error("wrong code returns: ", c)
		}
	})
	t.Run("compare test", func(t *testing.T) {
		for _, tt := range []struct {
			label    string
			arn      arn.DeliveryStream
			expected arn.CompareLevel
		}{
			{
				"all wrong",
				arn.NewDeliveryStream("hoge", "fuga", "piyo"),
				arn.CompareNotEqual,
			},
			{
				"same region but not equal",
				arn.NewDeliveryStream("foo", "fuga", "piyo"),
				arn.CompareNotEqual,
			},
			{
				"only same account",
				arn.NewDeliveryStream("hoge", "bar", "piyo"),
				arn.CompareEqualAccount,
			},
			{
				"same account and region",
				arn.NewDeliveryStream("foo", "bar", "piyo"),
				arn.CompareEqualRegionAccount,
			},
			{
				"all data are same",
				arn.NewDeliveryStream("foo", "bar", "baz"),
				arn.CompareEqualAll,
			},
		} {
			res := a.Compare(tt.arn)
			if res != tt.expected {
				t.Errorf(`label="%s" expected="%v" actual="%v"`, tt.label, tt.expected, res)
			}
		}
	})
}
