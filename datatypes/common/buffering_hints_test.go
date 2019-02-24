package common_test

import (
	"encoding/json"
	"testing"

	"github.com/taiyoh/toyhose/datatypes/common"
)

func TestBufferingHints(t *testing.T) {
	defaultSec := uint(300)
	defaultMBs := uint(5)
	setSec := uint(65)
	setMBs := uint(10)
	for _, tt := range []struct {
		label     string
		json      string
		errFound  bool
		beforeSec *uint
		beforeMBs *uint
		afterSec  *uint
		afterMBs  *uint
	}{
		{"no object", `{}`, false, nil, nil, &defaultSec, &defaultMBs},
		{"invalid IntervalInSeconds", `{"IntervalInSeconds":901,"SizeInMBs":1}`, true, nil, nil, nil, nil},
		{"invalid SizeInMBs", `{"IntervalInSeconds":60,"SizeInMBs":129}`, true, nil, nil, nil, nil},
		{"not filling default values", `{"IntervalInSeconds":65,"SizeInMBs":10}`, false, &setSec, &setMBs, &setSec, &setMBs},
	} {
		bh := common.BufferingHints{}
		if err := json.Unmarshal([]byte(tt.json), &bh); err != nil {
			t.Errorf(`label="%s" msg="unmarshal error found: %s"`, tt.label, err)
			break
		}
		if err := bh.Validate(); (err != nil) != tt.errFound {
			t.Errorf(`label="%s" msg="validation error found: %s"`, tt.label, err)
			break
		}
		if tt.errFound {
			continue
		}
		if bh.IntervalInSeconds == nil && bh.IntervalInSeconds != tt.beforeSec {
			t.Errorf(`label="%s" field="IntervalInSeconds" state="before" expected="%d" actual="%v"`, tt.label, *tt.beforeSec, *bh.IntervalInSeconds)
		}
		if bh.IntervalInSeconds != nil && *bh.IntervalInSeconds != *tt.beforeSec {
			t.Errorf(`label="%s" field="IntervalInSeconds" state="before" expected="%d" actual="%v"`, tt.label, *tt.beforeSec, *bh.IntervalInSeconds)
		}
		if bh.SizeInMBs == nil && bh.SizeInMBs != tt.beforeMBs {
			t.Errorf(`label="%s" field="SizeInMBs" state="before" expected="%d" actual="%v"`, tt.label, *tt.beforeMBs, *bh.SizeInMBs)
		}
		if bh.SizeInMBs != nil && *bh.SizeInMBs != *tt.beforeMBs {
			t.Errorf(`label="%s" field="SizeInMBs" state="before" expected="%d" actual="%v"`, tt.label, *tt.beforeMBs, *bh.SizeInMBs)
		}
		bh.FillDefaultValue()
		if *bh.IntervalInSeconds != *tt.afterSec {
			t.Errorf(`label="%s" field="IntervalInSeconds" state="after" expected="%d" actual="%v"`, tt.label, *tt.afterSec, *bh.IntervalInSeconds)
		}
		if *bh.SizeInMBs != *tt.afterMBs {
			t.Errorf(`label="%s" field="SizeInMBs" state="after" expected="%d" actual="%v"`, tt.label, *tt.afterMBs, *bh.SizeInMBs)
		}
	}
}
