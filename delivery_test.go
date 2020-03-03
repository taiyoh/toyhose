package toyhose

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

func TestOperateDeliveryFromAPI(t *testing.T) {
	intervalSeconds := int(1)
	sizeInMBs := int(5)
	awsConf := awsConfig()
	endpoint := os.Getenv("S3_ENDPOINT_URL")
	s3cli := s3Client(awsConf, endpoint)

	d := NewDispatcher(&DispatcherConfig{
		AWSConf: awsConf,
		S3InjectedConf: S3InjectedConf{
			IntervalInSeconds: &intervalSeconds,
			SizeInMBs:         &sizeInMBs,
			EndPoint:          &endpoint,
		},
	})
	mux := http.ServeMux{}
	mux.HandleFunc("/", d.Dispatch)
	testserver := httptest.NewServer(&mux)
	defer testserver.Close()

	bucketName := "delivery-api-test-" + uuid.New().String()
	closer, err := setupS3(s3cli, bucketName)
	if err != nil {
		t.Fatal(err)
	}
	defer closer()

	streamName := "foobar"
	prefix := "aaa-prefix"
	fh := firehose.New(session.New(awsConfig().WithEndpoint(testserver.URL)))
	t.Run("create delivery_stream", func(t *testing.T) {
		out, err := fh.CreateDeliveryStream(&firehose.CreateDeliveryStreamInput{
			DeliveryStreamName: &streamName,
			DeliveryStreamType: aws.String("DirectPut"),
			S3DestinationConfiguration: &firehose.S3DestinationConfiguration{
				BucketARN: aws.String("arn:aws:s3:::" + bucketName),
				BufferingHints: &firehose.BufferingHints{
					SizeInMBs:         aws.Int64(32),
					IntervalInSeconds: aws.Int64(60),
				},
				Prefix:  &prefix,
				RoleARN: aws.String("foo"),
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if out.DeliveryStreamARN == nil {
			t.Error("deliveryStreamARN not found")
		}
	})

	t.Run("put record", func(t *testing.T) {
		out, err := fh.PutRecord(&firehose.PutRecordInput{
			DeliveryStreamName: &streamName,
			Record: &firehose.Record{
				Data: []byte(base64.StdEncoding.EncodeToString([]byte("1111111111\n"))),
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if out.RecordId == nil {
			t.Error("RecordId not found")
		}
	})

	t.Run("put_batch records", func(t *testing.T) {
		out, err := fh.PutRecordBatch(&firehose.PutRecordBatchInput{
			DeliveryStreamName: &streamName,
			Records: []*firehose.Record{
				{Data: []byte(base64.StdEncoding.EncodeToString([]byte("2222222222\n")))},
				{Data: []byte(base64.StdEncoding.EncodeToString([]byte("3333333333\n")))},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if out.FailedPutCount == nil || *out.FailedPutCount > 0 {
			t.Errorf("unexpected FailedPutCount: %v", out.FailedPutCount)
		}
		for idx, res := range out.RequestResponses {
			if res.RecordId == nil {
				t.Errorf("[%d] response failed: %#v", idx, res)
			}
		}
	})

	t.Run("receive objects", func(t *testing.T) {
		var contents []*s3.Object
		for i := 0; i < 100; i++ {
			out, err := s3cli.ListObjects(&s3.ListObjectsInput{
				Bucket: &bucketName,
				Prefix: &prefix,
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(out.Contents) > 0 {
				contents = out.Contents
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		if len(contents) == 0 {
			t.Fatal("contents not found")
		}
		byteSize := 0
		for _, content := range contents {
			obj, err := s3cli.GetObject(&s3.GetObjectInput{
				Bucket: &bucketName,
				Key:    content.Key,
			})
			if err != nil {
				t.Fatal(err)
			}
			body, err := ioutil.ReadAll(obj.Body)
			if err != nil {
				t.Fatal(err)
			}
			byteSize += len(body)
		}
		if byteSize != 33 {
			t.Errorf("wrong content received: %d", byteSize)
		}
	})

	t.Run("delete delivery_stream", func(t *testing.T) {
		if _, err := fh.DeleteDeliveryStream(&firehose.DeleteDeliveryStreamInput{
			DeliveryStreamName: &streamName,
		}); err != nil {
			t.Fatal(err)
		}
	})
}
