package toyhose

import (
	"regexp"
	"testing"
	"time"
)

func TestS3Prefix(t *testing.T) {
	ts, err := time.Parse(time.RFC3339, "2018-08-27T10:30:00+00:00")
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		label     string
		prefix    string
		errPrefix string
		expected  [2]string
	}{
		{
			label:     "case:1",
			errPrefix: "myFirehoseFailures/!{firehose:error-output-type}/",
			expected: [2]string{
				"^2018/08/27/10/$",
				"^myFirehoseFailures/processing-failed/$",
			},
		},
		// case:2 is skipped
		{
			label:     "case:3",
			prefix:    "myFirehose/DeliveredYear=!{timestamp:yyyy}/anyMonth/rand=!{firehose:random-string}",
			errPrefix: "myFirehoseFailures/!{firehose:error-output-type}/!{timestamp:yyyy}/anyMonth/!{timestamp:dd}",
			expected: [2]string{
				"^myFirehose/DeliveredYear=2018/anyMonth/rand=([0-9a-f]{11})$",
				"^myFirehoseFailures/processing-failed/2018/anyMonth/27$",
			},
		},
		{
			label:     "case:4",
			prefix:    "myPrefix/year=!{timestamp:yyyy}/month=!{timestamp:MM}/day=!{timestamp:dd}/hour=!{timestamp:HH}/",
			errPrefix: "myErrorPrefix/year=!{timestamp:yyyy}/month=!{timestamp:MM}/day=!{timestamp:dd}/hour=!{timestamp:HH}/!{firehose:error-output-type}",
			expected: [2]string{
				"^myPrefix/year=2018/month=08/day=27/hour=10/$",
				"^myErrorPrefix/year=2018/month=08/day=27/hour=10/processing-failed$",
			},
		},
		{
			label:  "case:5",
			prefix: "myFirehosePrefix",
			expected: [2]string{
				"^myFirehosePrefix/2018/08/27/10/$",
				"^$",
			},
		},
	} {
		t.Run(tt.label, func(t *testing.T) {
			pref := keyPrefix(tt.prefix, ts)
			errPref := keyErrPrefix(tt.errPrefix, ts, processingFailed)
			if m, _ := regexp.Match(tt.expected[0], []byte(pref)); !m {
				t.Errorf("wrong prefix captured. expected:%s, actual:%s", tt.expected[0], pref)
			}
			if m, _ := regexp.Match(tt.expected[1], []byte(errPref)); !m {
				t.Errorf("wrong error prefix captured. expected:%s, actual:%s", tt.expected[1], errPref)
			}
		})
	}
}
