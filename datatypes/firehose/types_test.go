package firehose

import "testing"

func TestStreamType(t *testing.T) {
	if _, err := restoreStreamType("foo"); err == nil {
		t.Error("restoreStreamType should be failed")
	}
	typ, err := restoreStreamType("DirectPut")
	if err != nil {
		t.Error("restoreStreamType should be succeed:", err)
	}
	if typ != TypeDirectPut {
		t.Error("returned type is wrong")
	}
}
