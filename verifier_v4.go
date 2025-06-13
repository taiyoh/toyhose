package toyhose

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

var (
	errInvalidSignature = errors.New("invalid signature")
)

func verifyV4(ctx context.Context, c aws.Config, req *http.Request, body []byte) error {
	refAuth := req.Header.Get("Authorization")
	ts, _ := time.Parse("20060102T150405Z", req.Header.Get("X-Amz-Date"))

	copiedReq := req.Clone(ctx)
	copiedReq.Body = io.NopCloser(bytes.NewReader(body))

	// v1の挙動を模倣: 署名に含まれるヘッダーのみをリクエストに含める
	if authParts := strings.Split(refAuth, ", "); len(authParts) > 1 {
		if sigParts := strings.Split(authParts[1], "="); len(sigParts) > 1 {
			signedHeaders := strings.Split(sigParts[1], ";")
			newHeader := http.Header{}
			for _, h := range signedHeaders {
				if val := req.Header.Get(h); val != "" {
					newHeader.Set(h, val)
				}
			}
			// 必須ヘッダーを追加
			newHeader.Set("X-Amz-Date", req.Header.Get("X-Amz-Date"))
			copiedReq.Header = newHeader
		}
	}

	hasher := sha256.New()
	hasher.Write(body)
	payloadHash := hex.EncodeToString(hasher.Sum(nil))

	signer := v4.NewSigner()
	creds, err := c.Credentials.Retrieve(ctx)
	if err != nil {
		return err
	}

	if err := signer.SignHTTP(ctx, creds, copiedReq, payloadHash, "firehose", c.Region, ts); err != nil {
		return err
	}

	if a := copiedReq.Header.Get("Authorization"); a != refAuth {
		return errInvalidSignature
	}
	return nil
}
