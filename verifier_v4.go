package toyhose

import (
	"errors"
	"io"
	"net/http"
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
	ts, _ := time.Parse("20060102T150405Z", req.Header.Get("X-Amz-Date"))
	copiedReq := req.Clone(req.Context())
	copiedReq.Header.Del("Accept-Encoding") // anyone else?
	copiedReq.Header.Del("Authorization")
	if _, err := v4.NewSigner(c.Credentials).Sign(copiedReq, body, "firehose", *c.Region, ts); err != nil {
		return awserr.New("IncompleteSignature", "failed to build sign", err)
	}
	if a := copiedReq.Header.Get("Authorization"); a != refAuth {
		return awserr.New("AccessDeniedException", "signature mismatched", errInvalidSignature)
	}
	return nil
}
