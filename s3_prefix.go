package toyhose

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vjeantet/jodaTime"
)

type firehoseErrorType string

const (
	notFailed        firehoseErrorType = "____"
	processingFailed firehoseErrorType = "processing-failed"
)

var (
	fhErrOutputTypeRE = regexp.MustCompile("\\!\\{firehose:error-output-type\\}")
	fhRandStrRE       = regexp.MustCompile("\\!\\{firehose:random-string\\}")
	fhTimeStampRE     = regexp.MustCompile("\\!\\{timestamp:(.+?)\\}")
)

func extractNamespace(b []byte, ts time.Time) ([]byte, bool) {
	processed := false
	if fhRandStrRE.Match(b) {
		b = fhRandStrRE.ReplaceAllFunc(b, func(s []byte) []byte {
			return []byte(strings.ReplaceAll(uuid.New().String(), "-", "")[:11])
		})
		processed = true
	}
	if fhTimeStampRE.Match(b) {
		b = fhTimeStampRE.ReplaceAllFunc(b, func(s []byte) []byte {
			df := fhTimeStampRE.ReplaceAllString(string(s), "$1")
			return []byte(jodaTime.Format(df, ts))
		})
		processed = true
	}
	return b, processed
}

func keyPrefix(pref string, ts time.Time) string {
	if pref == "" {
		return ts.Format("2006/01/02/")
	}
	b, processed := extractNamespace([]byte(pref), ts)
	if processed {
		return string(b)
	}
	return fmt.Sprintf("%s/%s", string(b), ts.Format("2006/01/02/"))
}

func keyErrPrefix(pref string, ts time.Time, errType firehoseErrorType) string {
	if pref == "" {
		return ""
	}
	b := []byte(pref)
	if fhErrOutputTypeRE.Match(b) {
		b = fhErrOutputTypeRE.ReplaceAll(b, []byte(errType))
	}
	b, _ = extractNamespace(b, ts)
	return string(b)
}
