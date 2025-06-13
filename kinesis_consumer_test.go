package toyhose

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/firehose"
	fhtypes "github.com/aws/aws-sdk-go-v2/service/firehose/types"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

func TestKinesisConsumer(t *testing.T) {
	ctx := context.Background()
	awsConf := awsConfig(t)
	streamName := "stream-" + uuid.New().String()
	kiConf := KinesisInjectedConf{
		Endpoint: aws.String(kinesisEndpointURL),
	}
	t.Run("ErrCodeInvalidArgumentException", func(t *testing.T) {
		for _, tt := range [...]struct {
			label  string
			fhConf *fhtypes.KinesisStreamSourceConfiguration
			kiConf KinesisInjectedConf
		}{
			{
				label:  "no arn supplied",
				fhConf: &fhtypes.KinesisStreamSourceConfiguration{},
				kiConf: kiConf,
			},
			{
				label: "wrong arn",
				fhConf: &fhtypes.KinesisStreamSourceConfiguration{
					KinesisStreamARN: aws.String(fmt.Sprintf("arn:aws:kinesas:region:XXXXXXXX:stream/%s", streamName)),
				},
				kiConf: kiConf,
			},
			{
				label: "no endpoint supplied",
				fhConf: &fhtypes.KinesisStreamSourceConfiguration{
					KinesisStreamARN: aws.String(fmt.Sprintf("arn:aws:kinesis:region:foobarbaz:stream/%s", streamName)),
				},
			},
		} {
			t.Run(tt.label, func(t *testing.T) {
				consumer, err := newKinesisConsumer(context.Background(), awsConf, tt.fhConf, tt.kiConf)
				if consumer != nil {
					t.Fatalf("%#v", consumer)
				}
				if err == nil {
					t.Fatal("error should be exists")
				}
				var invalidArg *fhtypes.InvalidArgumentException
				if !errors.As(err, &invalidArg) {
					t.Errorf("unexpected error type: %v", err)
				}
			})
		}
	})
	kinCli := kinesis.NewFromConfig(awsConf, func(o *kinesis.Options) {
		o.BaseEndpoint = kiConf.Endpoint
	})
	if _, err := kinCli.CreateStream(ctx, &kinesis.CreateStreamInput{
		ShardCount: aws.Int32(1),
		StreamName: &streamName,
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = kinCli.DeleteStream(ctx, &kinesis.DeleteStreamInput{
			StreamName: &streamName,
		})
	})
	t.Run("inactive to active", func(t *testing.T) {
		fhConf := &fhtypes.KinesisStreamSourceConfiguration{
			KinesisStreamARN: aws.String(fmt.Sprintf("arn:aws:kinesis:%s:%s:stream/%s", awsConf.Region, "XXXXXXXX", streamName)),
		}
		var consumer *kinesisConsumer
		var err error
		for i := 0; i < 50; i++ {
			consumer, err = newKinesisConsumer(context.Background(), awsConf, fhConf, kiConf)
			if consumer != nil {
				break
			}
			if err != nil {
				var serviceUnavailable *fhtypes.ServiceUnavailableException
				if !errors.As(err, &serviceUnavailable) {
					t.Errorf("unexpected error type: %v", err)
				}
			}
			time.Sleep(50 * time.Millisecond)
		}
		if consumer == nil {
			t.Error("consumer not found")
		}
	})
}

func TestInputFromKinesis(t *testing.T) {
	ctx := context.Background()
	awsConf := awsConfig(t)
	streamName := "input-from-stream-" + uuid.New().String()
	s3cli := s3.NewFromConfig(awsConf, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(s3EndpointURL)
		o.UsePathStyle = true
	})
	kinCli := kinesis.NewFromConfig(awsConf, func(o *kinesis.Options) {
		o.BaseEndpoint = aws.String(kinesisEndpointURL)
	})
	if err := setupKinesisStream(t, kinCli, streamName, 1); err != nil {
		t.Fatal(err)
	}
	streamIsActive := false
	for i := 0; i < 20 || !streamIsActive; i++ {
		out, err := kinCli.DescribeStream(ctx, &kinesis.DescribeStreamInput{
			StreamName: &streamName,
		})
		if err != nil {
			t.Fatal(err)
		}
		if string(out.StreamDescription.StreamStatus) == string(fhtypes.DeliveryStreamStatusActive) {
			streamIsActive = true
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !streamIsActive {
		t.Fatal("stream is not activated")
	}
	if err := setupS3(t, s3cli, streamName); err != nil {
		t.Fatal(err)
	}

	svc := &DeliveryStreamService{
		awsConf:   awsConf,
		region:    awsConf.Region,
		accountID: "XXXXXXXX",
		s3InjectedConf: S3InjectedConf{
			IntervalInSeconds: aws.Int(1),
			EndPoint:          aws.String(s3EndpointURL),
		},
		kinesisInjectedConf: KinesisInjectedConf{
			Endpoint: aws.String(kinesisEndpointURL),
		},
		pool: &deliveryStreamPool{
			pool: map[string]*deliveryStream{},
		},
	}
	createInputBytes, _ := json.Marshal(&firehose.CreateDeliveryStreamInput{
		DeliveryStreamName: &streamName,
		DeliveryStreamType: fhtypes.DeliveryStreamTypeKinesisStreamAsSource,
		S3DestinationConfiguration: &fhtypes.S3DestinationConfiguration{
			BucketARN: aws.String("arn:aws:s3:::" + streamName),
			BufferingHints: &fhtypes.BufferingHints{
				IntervalInSeconds: aws.Int32(60),
				SizeInMBs:         aws.Int32(50),
			},
			CompressionFormat: fhtypes.CompressionFormatUncompressed,
			RoleARN:           aws.String("arn:aws:iam:role:foo-bar"),
		},
		KinesisStreamSourceConfiguration: &fhtypes.KinesisStreamSourceConfiguration{
			KinesisStreamARN: aws.String(fmt.Sprintf("arn:aws:kinesis:%s:%s:stream/%s", awsConf.Region, "XXXXXXXX", streamName)),
			RoleARN:          aws.String("arn:aws:iam:role:foo-bar"),
		},
	})
	createOutput, err := svc.Create(context.Background(), createInputBytes)
	if err != nil {
		t.Fatal(err)
	}
	deliveryStreamARN := *createOutput.DeliveryStreamARN

	if _, err := kinCli.PutRecord(ctx, &kinesis.PutRecordInput{
		StreamName:   &streamName,
		PartitionKey: aws.String("aaa"),
		Data:         []byte("aaaaaaaaaaaaaaaiiiiiiiiiiiiiiii"),
	}); err != nil {
		t.Fatal(err)
	}

	var captured []byte
	for i := 0; i < 10; i++ {
		time.Sleep(500 * time.Millisecond)
		out, err := s3cli.ListObjects(ctx, &s3.ListObjectsInput{Bucket: &streamName})
		if err != nil {
			t.Fatal(err)
		}
		if len(out.Contents) < 1 {
			continue
		}
		for _, o := range out.Contents {
			obj, err := s3cli.GetObject(ctx, &s3.GetObjectInput{
				Bucket: &streamName,
				Key:    o.Key,
			})
			if err != nil {
				t.Fatal(err)
			}
			b, _ := io.ReadAll(obj.Body)
			captured = append(captured, b...)
		}
		if len(captured) > 0 {
			break
		}
	}

	if c := string(captured); c != "aaaaaaaaaaaaaaaiiiiiiiiiiiiiiii" {
		t.Errorf("captured data is wrong: %s", c)
	}

	deleteInputBytes, _ := json.Marshal(firehose.DeleteDeliveryStreamInput{
		DeliveryStreamName: aws.String(strings.Split(deliveryStreamARN, "/")[1]),
	})
	if _, err := svc.Delete(context.Background(), deleteInputBytes); err != nil {
		t.Error(err)
	}
}
