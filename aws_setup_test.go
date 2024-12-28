package toyhose

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var awsConf *aws.Config
var s3EndpointURL = "http://localhost:9000"
var kinesisEndpointURL = "http://localhost:4567"

func init() {
	awsConf = aws.NewConfig()
	awsConf.Region = "us-east-1"
	awsConf.Credentials = credentials.NewStaticCredentialsProvider("XXXXXXXX", "YYYYYYYY", "")
	_ = SetupLogger()
}

func setupS3(t *testing.T, s3cli *s3.Client, bucket string) error {
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
