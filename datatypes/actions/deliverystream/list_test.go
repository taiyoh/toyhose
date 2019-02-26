package deliverystream_test

import (
	"testing"

	"github.com/taiyoh/toyhose/datatypes/actions/deliverystream"
)

func TestListInput(t *testing.T) {
	for _, tt := range []struct {
		label    string
		rawjson  string
		hasError bool
	}{
		{
			"broken input",
			"{}[",
			true,
		},
		{
			"invalid type",
			`{"DeliveryStreamType":"Foo"}`,
			true,
		},
		{
			"out-of-range Limit",
			`{"Limit":10001}`,
			true,
		},
		{
			"out-of-range ExclusiveStartName",
			`{"ExclusiveStartDeliveryStreamName":"12345678901234567890123456789012345678901234567890123456789012345"}`,
			true,
		},
		{
			"invalid ExclusiveStartName",
			`{"ExclusiveStartDeliveryStreamName":"foobar**!!@#"}`,
			true,
		},
		{
			"valid",
			`{"Limit":10,"ExclusiveStartDeliveryStreamName":"foobar","DeliveryStreamType":"DirectPut"}`,
			false,
		},
	} {
		li, err := deliverystream.NewListInput([]byte(tt.rawjson))
		if (li != nil) == tt.hasError {
			t.Errorf(`label="%s" msg="input resource returns"`, tt.label)
		}
		if (err != nil) != tt.hasError {
			t.Errorf(`label="%s" msg="NewListInput error should returns"`, tt.label)
		}
		if li != nil && li.ExclusiveStartDeliveryStreamName() != "foobar" {
			t.Errorf(`label="%s" msg="ExclusiveStartDeliveryStartName is wrong" expected="foobar" actual="%s"`, tt.label, li.ExclusiveStartDeliveryStreamName())
		}
	}
}

func TestListInputDefault(t *testing.T) {
	li, err := deliverystream.NewListInput([]byte("{}"))
	if err != nil {
		t.Error("error found:", err)
	}
	if li.ExclusiveStartName != nil {
		t.Error("ExclusiveStartName is not assigned")
	}
	if li.Type != nil {
		t.Error("Type is not assigned")
	}
	if li.Limit == nil {
		t.Error("Limit should be assigned")
		return
	}
	if *li.Limit != uint(10) {
		t.Error("default Limit value should be uint(10)")
	}
	if li.ExclusiveStartDeliveryStreamName() != "*" {
		t.Error("default ExclusiveStartDeliveryStreamName is `*`")
	}
}
