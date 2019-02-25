package firehose_test

import (
	"testing"

	"github.com/taiyoh/toyhose/datatypes/firehose"
)

func TestStreamType(t *testing.T) {
	if _, err := firehose.RestoreStreamType("foo"); err == nil {
		t.Error("restoreStreamType should be failed")
	}
	typ, err := firehose.RestoreStreamType("DirectPut")
	if err != nil {
		t.Error("restoreStreamType should be succeed:", err)
	}
	if typ != firehose.TypeDirectPut {
		t.Error("returned type is wrong")
	}
}
