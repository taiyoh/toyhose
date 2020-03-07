package toyhose

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
)

var (
	errInvalidSignature = errors.New("invalid signature")
)

func verifyV4(c *aws.Config, req *http.Request, body io.ReadSeeker) error {
	refAuth := req.Header.Get("Authorization")
	sigHeaders := strings.Split(refAuth, ", ")[1]
	sigHeaderVal := strings.Split(sigHeaders, "=")[1]
	ts, _ := time.Parse("20060102T150405Z", req.Header.Get("X-Amz-Date"))
	copiedReq := req.Clone(req.Context())
	remainHeaders := http.Header{}
	for _, key := range strings.Split(sigHeaderVal, ";") {
		if val := req.Header.Get(key); val != "" {
			remainHeaders.Set(key, val)
		}
	}
	copiedReq.Header = remainHeaders
	if _, err := v4.NewSigner(c.Credentials).Sign(copiedReq, body, "firehose", *c.Region, ts); err != nil {
		return awserr.New("IncompleteSignature", "failed to build sign", err)
	}
	if a := copiedReq.Header.Get("Authorization"); a != refAuth {
		return awserr.New("AccessDeniedException", "signature mismatched", errInvalidSignature)
	}
	return nil
}
