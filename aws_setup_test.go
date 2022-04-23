package toyhose

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/s3"
)

var awsConf *aws.Config
var s3EndpointURL = "http://localhost:9000"
var kinesisEndpointURL = "http://localhost:4567"

func init() {
	awsConf = aws.NewConfig().
		WithRegion("us-east-1").
		WithCredentials(credentials.NewStaticCredentials("XXXXXXXX", "YYYYYYYY", ""))
	_ = SetupLogger()
}

func setupS3(t *testing.T, s3cli *s3.S3, bucket string) error {
	if _, err := s3cli.CreateBucket(&s3.CreateBucketInput{
		Bucket: &bucket,
	}); err != nil {
		return err
	}
	t.Cleanup(func() {
		_, _ = s3cli.DeleteBucket(&s3.DeleteBucketInput{Bucket: &bucket})
	})
	return nil
}

func setupKinesisStream(t *testing.T, kinCli *kinesis.Kinesis, streamName string, shardCount int64) error {
	if _, err := kinCli.CreateStream(&kinesis.CreateStreamInput{
		StreamName: &streamName,
		ShardCount: &shardCount,
	}); err != nil {
		return err
	}
	t.Cleanup(func() {
		_, _ = kinCli.DeleteStream(&kinesis.DeleteStreamInput{
			StreamName: &streamName,
		})
	})
	return nil
}
