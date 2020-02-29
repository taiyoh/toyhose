package toyhose

import (
	"encoding/json"
	"io/ioutil"
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
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		t.Logf("method: %s, path: %s", r.Method, r.URL.Path)
		t.Logf("headers: %v", r.Header)
		t.Logf("body: %s", string(body))
		i := firehose.CreateDeliveryStreamInput{}
		if err := json.Unmarshal(body, &i); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		t.Logf("unmarshalled: %v", i)
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
