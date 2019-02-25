package firehose_test

import (
	"testing"

	"github.com/taiyoh/toyhose/datatypes/arn"
	"github.com/taiyoh/toyhose/datatypes/firehose"
)

func TestDeliveryStream(t *testing.T) {
	ds := firehose.NewDeliveryStream(arn.NewDeliveryStream("foo", "bar", "baz"))
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
