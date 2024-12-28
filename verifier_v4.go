package toyhose

import (
	"context"
	"crypto/sha256"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/smithy-go"
)

var (
	errInvalidSignature = errors.New("invalid signature")
)

func verifyV4(c *aws.Config, req *http.Request, body []byte) error {
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
	cred, err := c.Credentials.Retrieve(context.Background())
	if err != nil {
		return &smithy.GenericAPIError{
			Message: "failed to retrieve credentials",
			Code:    "IncompleteSignature",
			Fault:   smithy.FaultServer,
		}
	}
	payloadHash := sha256.Sum256(body)
	if err := v4.NewSigner().SignHTTP(context.Background(), cred, copiedReq, string(payloadHash[:]), "firehose", c.Region, ts); err != nil {
		return &smithy.GenericAPIError{
			Message: "failed to build sign",
			Code:    "IncompleteSignature",
			Fault:   smithy.FaultClient,
		}
	}
	if a := copiedReq.Header.Get("Authorization"); a != refAuth {
		return &smithy.GenericAPIError{
			Message: "signature mismatched",
			Code:    "IncompleteSignature",
			Fault:   smithy.FaultClient,
		}
	}
	return nil
}
