package toyhose

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var s3EndpointURL = "http://localhost:9000"
var kinesisEndpointURL = "http://localhost:4567"

func awsConfig(t *testing.T) aws.Config {
	t.Helper()
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("XXXXXXXX", "YYYYYYYY", ""),
		),
	)
	if err != nil {
		t.Fatalf("failed to load aws config: %v", err)
	}
	return cfg
}

func init() {
	_ = SetupLogger()
}

func setupS3(t *testing.T, s3cli *s3.Client, bucket string) error {
	t.Helper()
	if _, err := s3cli.CreateBucket(context.Background(), &s3.CreateBucketInput{
		Bucket: &bucket,
	}); err != nil {
		return err
	}
	t.Cleanup(func() {
		_, _ = s3cli.DeleteBucket(context.Background(), &s3.DeleteBucketInput{Bucket: &bucket})
	})
	return nil
}

func setupKinesisStream(t *testing.T, kinCli *kinesis.Client, streamName string, shardCount int32) error {
	t.Helper()
	if _, err := kinCli.CreateStream(context.Background(), &kinesis.CreateStreamInput{
		StreamName: &streamName,
		ShardCount: &shardCount,
	}); err != nil {
		return err
	}
	t.Cleanup(func() {
		_, _ = kinCli.DeleteStream(context.Background(), &kinesis.DeleteStreamInput{
			StreamName: &streamName,
		})
	})
	return nil
}
