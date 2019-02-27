package firehose_test

import (
	"testing"

	"github.com/taiyoh/toyhose/datatypes/firehose"
)

func TestStreamType(t *testing.T) {
	t.Run("required", func(t *testing.T) {
		_, err := firehose.RestoreStreamType("")
		if err == nil {
			t.Error("restoreStreamType should be failed")
			return
		}
		e, ok := err.(firehose.Error)
		if !ok {
			t.Error("type cast failed")
			return
		}
		if e != firehose.ErrRequired {
			t.Error("wrong error captured")
		}
	})

	t.Run("invalid", func(t *testing.T) {
		_, err := firehose.RestoreStreamType("foo")
		if err == nil {
			t.Error("restoreStreamType should be failed")
			return
		}
		e, ok := err.(firehose.Error)
		if !ok {
			t.Error("type cast failed")
			return
		}
		if e != firehose.ErrNotFound {
			t.Error("wrong error captured")
		}
	})

	t.Run("valid", func(t *testing.T) {
		typ, err := firehose.RestoreStreamType("DirectPut")
		if err != nil {
			t.Error("restoreStreamType should be succeed:", err)
		}
		if typ != firehose.TypeDirectPut {
			t.Error("returned type is wrong")
		}
	})
}
