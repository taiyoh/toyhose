package toyhose

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

func TestKinesisConsumer(t *testing.T) {
	streamName := "stream-" + uuid.New().String()
	kiConf := KinesisInjectedConf{
		Endpoint: aws.String(kinesisEndpointURL),
	}
	t.Run("ErrCodeInvalidArgumentException", func(t *testing.T) {
		for _, tt := range [...]struct {
			label  string
			fhConf *firehose.KinesisStreamSourceConfiguration
			kiConf KinesisInjectedConf
		}{
			{
				label:  "no arn supplied",
				fhConf: &firehose.KinesisStreamSourceConfiguration{},
				kiConf: kiConf,
			},
			{
				label: "wrong arn",
				fhConf: &firehose.KinesisStreamSourceConfiguration{
					KinesisStreamARN: aws.String(fmt.Sprintf("arn:aws:kinesas:region:XXXXXXXX:stream/%s", streamName)),
				},
				kiConf: kiConf,
			},
			{
				label: "wrong account_id",
				fhConf: &firehose.KinesisStreamSourceConfiguration{
					KinesisStreamARN: aws.String(fmt.Sprintf("arn:aws:kinesis:region:foobarbaz:stream/%s", streamName)),
				},
				kiConf: kiConf,
			},
			{
				label: "no endpoint supplied",
				fhConf: &firehose.KinesisStreamSourceConfiguration{
					KinesisStreamARN: aws.String(fmt.Sprintf("arn:aws:kinesis:region:foobarbaz:stream/%s", streamName)),
				},
			},
		} {
			t.Run(tt.label, func(t *testing.T) {
				consumer, err := newKinesisConsumer(context.Background(), awsConf, tt.fhConf, tt.kiConf)
				if consumer != nil {
					t.Fatalf("%#v", consumer)
				}
				ae, ok := err.(awserr.Error)
				if !ok {
					t.Error(err)
				}
				if ae.Code() != firehose.ErrCodeInvalidArgumentException {
					t.Error(ae)
				}
			})
		}
	})
	kinCli := kinesis.New(session.Must(session.NewSession(awsConf.Copy().WithEndpoint(*kiConf.Endpoint))))
	if _, err := kinCli.CreateStream(&kinesis.CreateStreamInput{
		ShardCount: aws.Int64(1),
		StreamName: &streamName,
	}); err != nil {
		t.Fatal(err)
	}
	defer func() {
		kinCli.DeleteStream(&kinesis.DeleteStreamInput{
			StreamName: &streamName,
		})
	}()
	t.Run("inactive to active", func(t *testing.T) {
		cred, _ := awsConf.Credentials.Get()
		fhConf := &firehose.KinesisStreamSourceConfiguration{
			KinesisStreamARN: aws.String(fmt.Sprintf("arn:aws:kinesis:%s:%s:stream/%s", *awsConf.Region, cred.AccessKeyID, streamName)),
		}
		var consumer *kinesisConsumer
		var err error
		for i := 0; i < 50; i++ {
			consumer, err = newKinesisConsumer(context.Background(), awsConf, fhConf, kiConf)
			if consumer != nil {
				break
			}
			if err != nil {
				switch ae, ok := err.(awserr.Error); {
				case !ok:
					t.Error(ae)
				case ae.Code() != firehose.ErrCodeServiceUnavailableException:
					t.Error(ae)
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
	streamName := "input-from-stream-" + uuid.New().String()
	s3cli := s3.New(session.Must(session.NewSession(
		awsConf.Copy().WithEndpoint(s3EndpointURL).WithS3ForcePathStyle(true).WithDisableSSL(true))))
	kinCli := kinesis.New(session.Must(session.NewSession(
		awsConf.Copy().WithEndpoint(kinesisEndpointURL))))
	closer, err := setupKinesisStream(kinCli, streamName, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer closer()
	streamIsActive := false
	for i := 0; i < 20 || !streamIsActive; i++ {
		out, err := kinCli.DescribeStream(&kinesis.DescribeStreamInput{
			StreamName: &streamName,
		})
		if err != nil {
			t.Fatal(err)
		}
		if *out.StreamDescription.StreamStatus == "ACTIVE" {
			streamIsActive = true
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !streamIsActive {
		t.Fatal("stream is not activated")
	}
	s3Closer, err := setupS3(s3cli, streamName)
	if err != nil {
		t.Fatal(err)
	}
	defer s3Closer()

	cred, _ := awsConf.Credentials.Get()
	svc := &DeliveryStreamService{
		awsConf:   awsConf,
		region:    *awsConf.Region,
		accountID: cred.AccessKeyID,
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
		DeliveryStreamType: aws.String("KinesisStreamAsSource"),
		S3DestinationConfiguration: &firehose.S3DestinationConfiguration{
			BucketARN: aws.String("arn:aws:s3:::" + streamName),
			BufferingHints: &firehose.BufferingHints{
				IntervalInSeconds: aws.Int64(60),
				SizeInMBs:         aws.Int64(50),
			},
			CompressionFormat: aws.String("UNCOMPRESSED"),
			RoleARN:           aws.String("arn:aws:iam:role:foo-bar"),
		},
		KinesisStreamSourceConfiguration: &firehose.KinesisStreamSourceConfiguration{
			KinesisStreamARN: aws.String(fmt.Sprintf("arn:aws:kinesis:%s:%s:stream/%s", *awsConf.Region, cred.AccessKeyID, streamName)),
			RoleARN:          aws.String("arn:aws:iam:role:foo-bar"),
		},
	})
	createOutput, err := svc.Create(context.Background(), createInputBytes)
	if err != nil {
		t.Fatal(err)
	}
	deliveryStreamARN := *createOutput.DeliveryStreamARN

	if _, err := kinCli.PutRecord(&kinesis.PutRecordInput{
		StreamName:   &streamName,
		PartitionKey: aws.String("aaa"),
		Data:         []byte("aaaaaaaaaaaaaaaiiiiiiiiiiiiiiii"),
	}); err != nil {
		t.Fatal(err)
	}

	var captured []byte
	for i := 0; i < 10; i++ {
		time.Sleep(500 * time.Millisecond)
		out, err := s3cli.ListObjects(&s3.ListObjectsInput{Bucket: &streamName})
		if err != nil {
			t.Fatal(err)
		}
		if len(out.Contents) < 1 {
			continue
		}
		for _, o := range out.Contents {
			obj, err := s3cli.GetObject(&s3.GetObjectInput{
				Bucket: &streamName,
				Key:    o.Key,
			})
			if err != nil {
				t.Fatal(err)
			}
			b, _ := ioutil.ReadAll(obj.Body)
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
