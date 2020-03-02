package toyhose

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
)

func TestCreateDeliveryRequest(t *testing.T) {
	mux := http.ServeMux{}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		i := firehose.CreateDeliveryStreamInput{}
		b := bytes.NewBuffer([]byte{})
		if err := json.NewDecoder(io.TeeReader(r.Body, b)).Decode(&i); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// t.Logf("method: %s, host: %s, path: %s", r.Method, r.Host, r.URL.Path)
		// t.Logf("headers: %#v", r.Header)
		if err := verifyV4(r, bytes.NewReader(b.Bytes())); err != nil {
			t.Logf("verifyV4 failed: %v", err)
			w.WriteHeader(http.StatusForbidden)
			return
		}
	})
	testserver := httptest.NewServer(&mux)
	defer testserver.Close()

	fh := firehose.New(session.New(awsConfig().WithEndpoint(testserver.URL)))
	input := &firehose.CreateDeliveryStreamInput{
		DeliveryStreamName: aws.String("foobar"),
		DeliveryStreamType: aws.String("DirectPut"),
		S3DestinationConfiguration: &firehose.S3DestinationConfiguration{
			BucketARN: aws.String("arn:aws:s3:::testbucket"),
			BufferingHints: &firehose.BufferingHints{
				SizeInMBs:         aws.Int64(32),
				IntervalInSeconds: aws.Int64(60),
			},
			Prefix:  aws.String("aaa-prefix"),
			RoleARN: aws.String("foo"),
		},
	}
	out, err := fh.CreateDeliveryStream(input)
	if err != nil {
		t.Fatal(err)
	}
	if out.DeliveryStreamARN == nil {
		t.Error("deliveryStreamARN not found")
	}
}
