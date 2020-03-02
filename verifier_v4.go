package toyhose

import (
	"errors"
	"io"
	"net/http"
	"time"

	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
)

var (
	errBuildingSignFailed = errors.New("failed to build sign")
	errInvalidSignature   = errors.New("invalid signature")
)

func verifyV4(req *http.Request, body io.ReadSeeker) error {
	refAuth := req.Header.Get("Authorization")
	ts, _ := time.Parse("20060102T150405Z", req.Header.Get("X-Amz-Date"))
	c := awsConfig()
	copiedReq := req.Clone(req.Context())
	copiedReq.Header.Del("Accept-Encoding") // anyone else?
	if _, err := v4.NewSigner(c.Credentials).Sign(copiedReq, body, "firehose", *c.Region, ts); err != nil {
		return errBuildingSignFailed
	}
	if a := copiedReq.Header.Get("Authorization"); a != refAuth {
		return errInvalidSignature
	}
	return nil
}
