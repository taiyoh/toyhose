package toyhose

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-sdk-go/service/kinesis"
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
	kinCli := kinesis.New(session.New(awsConf.WithEndpoint(*kiConf.Endpoint)))
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
