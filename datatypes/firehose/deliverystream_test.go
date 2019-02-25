package firehose_test

import (
	"testing"

	"github.com/taiyoh/toyhose/datatypes/arn"
	"github.com/taiyoh/toyhose/datatypes/firehose"
)

func TestDeliveryStream(t *testing.T) {
	ds, err := firehose.NewDeliveryStream(arn.NewDeliveryStream("foo", "bar", "baz"), "hoge")
	if err == nil {
		t.Error("error should be returns")
	}
	if ds != nil {
		t.Error("delivery stream object should not be returns")
	}

	ds, err = firehose.NewDeliveryStream(arn.NewDeliveryStream("foo", "bar", "baz"), "DirectPut")
	if err != nil {
		t.Error("error returns:", err)
		return
	}
	newDS := ds.Active()
	if ds == newDS {
		t.Error("pointer address of returned value should be wrong")
	}
	if ds.Status == newDS.Status {
		t.Error("status should be changed")
	}
	if newDS.Status != firehose.StatusActive {
		t.Error("status should be ACTIVE")
	}
}
