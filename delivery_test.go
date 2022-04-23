package toyhose

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
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
	s3cli := s3Client(awsConf, s3EndpointURL)

	d := NewDispatcher(&DispatcherConfig{
		AWSConf: awsConf,
		S3InjectedConf: S3InjectedConf{
			IntervalInSeconds: &intervalSeconds,
			SizeInMBs:         &sizeInMBs,
			EndPoint:          &s3EndpointURL,
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
	fh := firehose.New(session.Must(session.NewSession(
		awsConf.Copy().WithEndpoint(testserver.URL))))
	for _, bucket := range []string{"", "foobarbaz"} {
		t.Run(fmt.Sprintf("bucket: [%s]", bucket), func(t *testing.T) {
			out, err := fh.CreateDeliveryStream(&firehose.CreateDeliveryStreamInput{
				DeliveryStreamName: &streamName,
				DeliveryStreamType: aws.String("DirectPut"),
				S3DestinationConfiguration: &firehose.S3DestinationConfiguration{
					BucketARN: aws.String("arn:aws:s3:::" + bucket),
					BufferingHints: &firehose.BufferingHints{
						SizeInMBs:         aws.Int64(32),
						IntervalInSeconds: aws.Int64(60),
					},
					Prefix:  &prefix,
					RoleARN: aws.String("foo"),
				},
			})
			if err == nil {
				t.Error("error should exists")
			}
			if out.DeliveryStreamARN != nil {
				t.Errorf("unexpected CreateDeliveryStreamOutput received: %#v", out)
			}
		})
	}

	t.Run("create and describe delivery_stream", func(t *testing.T) {
		bufferingHints := &firehose.BufferingHints{
			SizeInMBs:         aws.Int64(32),
			IntervalInSeconds: aws.Int64(60),
		}
		cout, err := fh.CreateDeliveryStream(&firehose.CreateDeliveryStreamInput{
			DeliveryStreamName: &streamName,
			DeliveryStreamType: aws.String("DirectPut"),
			S3DestinationConfiguration: &firehose.S3DestinationConfiguration{
				BucketARN:      aws.String("arn:aws:s3:::" + bucketName),
				BufferingHints: bufferingHints,
				Prefix:         &prefix,
				RoleARN:        aws.String("foo"),
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if cout.DeliveryStreamARN == nil {
			t.Error("deliveryStreamARN not found")
		}

		dout, err := fh.DescribeDeliveryStream(&firehose.DescribeDeliveryStreamInput{
			DeliveryStreamName: &streamName,
		})
		if err != nil {
			t.Fatal(err)
		}
		if dout.DeliveryStreamDescription.Source != nil {
			t.Error("streamType is DirectPut")
		}
		if len(dout.DeliveryStreamDescription.Destinations) != 1 {
			t.Errorf("unexpected destination_description included %#v", dout.DeliveryStreamDescription.Destinations)
		}
		desc := dout.DeliveryStreamDescription.Destinations[0]
		if desc.S3DestinationDescription == nil {
			t.Fatal("S3DestinationDescription is missing")
		}
		s3Dest := desc.S3DestinationDescription
		if *s3Dest.BucketARN != "arn:aws:s3:::"+bucketName {
			t.Errorf("unexpected BucketARN: %s", *s3Dest.BucketARN)
		}
		if !reflect.DeepEqual(*s3Dest.BufferingHints, *bufferingHints) {
			t.Errorf("unexpected bufferingHints: %#v", s3Dest.BufferingHints)
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

	t.Run("list deliveryStreams", func(t *testing.T) {
		streamNameBase := "stream-for-listing-%d"
		input := &firehose.CreateDeliveryStreamInput{
			DeliveryStreamType: aws.String("DirectPut"),
			S3DestinationConfiguration: &firehose.S3DestinationConfiguration{
				BucketARN: aws.String("arn:aws:s3:::" + bucketName),
				BufferingHints: &firehose.BufferingHints{
					SizeInMBs:         aws.Int64(1),
					IntervalInSeconds: aws.Int64(60),
				},
				Prefix:  &prefix,
				RoleARN: aws.String("foo"),
			},
		}
		for i := 1; i <= 15; i++ {
			input.DeliveryStreamName = aws.String(fmt.Sprintf(streamNameBase, i))
			out, err := fh.CreateDeliveryStream(input)
			if err != nil {
				t.Fatal(err)
			}
			if out.DeliveryStreamARN == nil {
				t.Fatal("deliveryStreamARN not found")
			}
			time.Sleep(100 * time.Millisecond)
		}
		out, err := fh.ListDeliveryStreams(&firehose.ListDeliveryStreamsInput{
			DeliveryStreamType: aws.String("DirectPut"),
			Limit:              aws.Int64(10),
		})
		if err != nil {
			t.Fatal(err)
		}
		if !*out.HasMoreDeliveryStreams {
			t.Error("HasMoreDeliveryStreams should be true")
		}
		if l := len(out.DeliveryStreamNames); l != 10 {
			t.Errorf("unexpected StreamNames returns: %d", l)
		}
		for i := 0; i < 10; i++ {
			expected := fmt.Sprintf(streamNameBase, i+1)
			if *out.DeliveryStreamNames[i] != expected {
				t.Errorf("index:%d expected:%s, actual:%s", i, expected, *out.DeliveryStreamNames[i])
			}
		}
		out, err = fh.ListDeliveryStreams(&firehose.ListDeliveryStreamsInput{
			DeliveryStreamType:               aws.String("DirectPut"),
			ExclusiveStartDeliveryStreamName: aws.String(fmt.Sprintf(streamNameBase, 10)),
		})
		if err != nil {
			t.Fatal(err)
		}
		if *out.HasMoreDeliveryStreams {
			t.Error("HasMoreDeliveryStreams should be false")
		}
		if l := len(out.DeliveryStreamNames); l != 5 {
			t.Errorf("unexpected StreamNames returns: %d", l)
		}
		for i := 0; i < 5; i++ {
			expected := fmt.Sprintf(streamNameBase, i+11)
			if *out.DeliveryStreamNames[i] != expected {
				t.Errorf("index:%d expected:%s, actual:%s", i, expected, *out.DeliveryStreamNames[i])
			}
		}
		out, err = fh.ListDeliveryStreams(&firehose.ListDeliveryStreamsInput{
			DeliveryStreamType: aws.String("KinesisStreamAsSource"),
		})
		if err != nil {
			t.Fatal(err)
		}
		if *out.HasMoreDeliveryStreams {
			t.Error("HasMoreDeliveryStreams should be false")
		}
		if l := len(out.DeliveryStreamNames); l != 0 {
			t.Errorf("unexpected StreamNames returns: %d", l)
		}
	})
}
