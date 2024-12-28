package toyhose

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/firehose"
	fhtypes "github.com/aws/aws-sdk-go-v2/service/firehose/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
)

func TestOperateDeliveryFromAPI(t *testing.T) {
	intervalSeconds := int(1)
	sizeInMBs := int(5)
	s3cli := s3Client(awsConf, s3EndpointURL)

	d := NewDispatcher(context.Background(), &DispatcherConfig{
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
	if err := setupS3(t, s3cli, bucketName); err != nil {
		t.Fatal(err)
	}

	streamName := "foobar"
	prefix := "aaa-prefix"
	fh := firehose.NewFromConfig(*awsConf, func(o *firehose.Options) {
		o.BaseEndpoint = &testserver.URL
	})
	for _, bucket := range []string{"", "foobarbaz"} {
		t.Run(fmt.Sprintf("bucket: [%s]", bucket), func(t *testing.T) {
			out, err := fh.CreateDeliveryStream(context.Background(), &firehose.CreateDeliveryStreamInput{
				DeliveryStreamName: &streamName,
				DeliveryStreamType: fhtypes.DeliveryStreamTypeDirectPut,
				S3DestinationConfiguration: &fhtypes.S3DestinationConfiguration{
					BucketARN: aws.String("arn:aws:s3:::" + bucket),
					BufferingHints: &fhtypes.BufferingHints{
						SizeInMBs:         aws.Int32(32),
						IntervalInSeconds: aws.Int32(60),
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
		bufferingHints := &fhtypes.BufferingHints{
			SizeInMBs:         aws.Int32(32),
			IntervalInSeconds: aws.Int32(60),
		}
		cout, err := fh.CreateDeliveryStream(context.Background(), &firehose.CreateDeliveryStreamInput{
			DeliveryStreamName: &streamName,
			DeliveryStreamType: fhtypes.DeliveryStreamTypeDirectPut,
			S3DestinationConfiguration: &fhtypes.S3DestinationConfiguration{
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

		dout, err := fh.DescribeDeliveryStream(context.Background(), &firehose.DescribeDeliveryStreamInput{
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
		out, err := fh.PutRecord(context.Background(), &firehose.PutRecordInput{
			DeliveryStreamName: &streamName,
			Record: &fhtypes.Record{
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
		out, err := fh.PutRecordBatch(context.Background(), &firehose.PutRecordBatchInput{
			DeliveryStreamName: &streamName,
			Records: []fhtypes.Record{
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
		var contents []s3types.Object
		for i := 0; i < 100; i++ {
			out, err := s3cli.ListObjects(context.Background(), &s3.ListObjectsInput{
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
			obj, err := s3cli.GetObject(context.Background(), &s3.GetObjectInput{
				Bucket: &bucketName,
				Key:    content.Key,
			})
			if err != nil {
				t.Fatal(err)
			}
			body, err := io.ReadAll(obj.Body)
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
		if _, err := fh.DeleteDeliveryStream(context.Background(), &firehose.DeleteDeliveryStreamInput{
			DeliveryStreamName: &streamName,
		}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("list deliveryStreams", func(t *testing.T) {
		streamNameBase := "stream-for-listing-%d"
		input := &firehose.CreateDeliveryStreamInput{
			DeliveryStreamType: fhtypes.DeliveryStreamTypeDirectPut,
			S3DestinationConfiguration: &fhtypes.S3DestinationConfiguration{
				BucketARN: aws.String("arn:aws:s3:::" + bucketName),
				BufferingHints: &fhtypes.BufferingHints{
					SizeInMBs:         aws.Int32(1),
					IntervalInSeconds: aws.Int32(60),
				},
				Prefix:  &prefix,
				RoleARN: aws.String("foo"),
			},
		}
		for i := 1; i <= 15; i++ {
			input.DeliveryStreamName = aws.String(fmt.Sprintf(streamNameBase, i))
			out, err := fh.CreateDeliveryStream(context.Background(), input)
			if err != nil {
				t.Fatal(err)
			}
			if out.DeliveryStreamARN == nil {
				t.Fatal("deliveryStreamARN not found")
			}
			time.Sleep(100 * time.Millisecond)
		}
		out, err := fh.ListDeliveryStreams(context.Background(), &firehose.ListDeliveryStreamsInput{
			DeliveryStreamType: fhtypes.DeliveryStreamTypeDirectPut,
			Limit:              aws.Int32(10),
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
			if out.DeliveryStreamNames[i] != expected {
				t.Errorf("index:%d expected:%s, actual:%s", i, expected, out.DeliveryStreamNames[i])
			}
		}
		out, err = fh.ListDeliveryStreams(context.Background(), &firehose.ListDeliveryStreamsInput{
			DeliveryStreamType:               fhtypes.DeliveryStreamTypeDirectPut,
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
			if out.DeliveryStreamNames[i] != expected {
				t.Errorf("index:%d expected:%s, actual:%s", i, expected, out.DeliveryStreamNames[i])
			}
		}
		out, err = fh.ListDeliveryStreams(context.Background(), &firehose.ListDeliveryStreamsInput{
			DeliveryStreamType: fhtypes.DeliveryStreamTypeKinesisStreamAsSource,
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
