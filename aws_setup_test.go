package toyhose

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3"
)

var awsConf *aws.Config
var s3EndpointURL = "http://localhost:9000"
var kinesisEndpointURL = "http://localhost:4567"

func init() {
	awsConf = aws.NewConfig().
		WithRegion("us-east-1").
		WithCredentials(credentials.NewStaticCredentials("XXXXXXXX", "YYYYYYYY", ""))
}

func setupS3(s3cli *s3.S3, bucket string) (func(), error) {
	if _, err := s3cli.CreateBucket(&s3.CreateBucketInput{
		Bucket: &bucket,
	}); err != nil {
		return nil, err
	}
	fn := func() {
		s3cli.DeleteBucket(&s3.DeleteBucketInput{Bucket: &bucket})
	}
	return fn, nil
}
