package toyhose

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

func init() {
	time.Local = nil
}

type s3Destination struct {
	deliveryName string
	source       <-chan []byte
	conf         *firehose.S3DestinationConfiguration
	captured     []byte
}

var (
	s3cli   *s3.S3
	awsConf *aws.Config
)

func awsConfig() *aws.Config {
	if awsConf != nil {
		return awsConf
	}
	awsConf = aws.NewConfig().
		WithRegion(os.Getenv("AWS_DEFAULT_REGION")).
		WithCredentials(credentials.NewEnvCredentials())
	return awsConf
}

func s3Client() *s3.S3 {
	if s3cli != nil {
		return s3cli
	}
	endpoint := os.Getenv("S3_ENDPOINT_URL")
	if endpoint == "" {
		panic("require S3_ENDPOINT_URL")
	}
	s3cli = s3.New(session.New(awsConfig().WithEndpoint(endpoint).WithS3ForcePathStyle(true).WithDisableSSL(true)))
	return s3cli
}

type storeToS3Resource struct {
	deliveryName       string
	bucketName         string
	prefix             string
	shouldGZipCompress bool
}

func storeToS3(ctx context.Context, resource storeToS3Resource, ts time.Time, data []byte) {
	if len(data) < 1 {
		return
	}
	var seekable []byte
	if resource.shouldGZipCompress {
		b := bytes.NewBuffer([]byte{})
		w := gzip.NewWriter(b)
		w.Write(data)
		w.Close()
		seekable = b.Bytes()
	} else {
		seekable = data
	}
	pref := strings.TrimSuffix(keyPrefix(resource.prefix, ts), "/")
	key := fmt.Sprintf("%s/%s-1-%s-%s", pref, resource.deliveryName, ts.Format("2006-01-02-15-04-05"), uuid.New())
	input := &s3.PutObjectInput{
		Bucket: &resource.bucketName,
		Body:   bytes.NewReader(seekable),
		Key:    &key,
	}
	cli := s3Client()
	for i := 0; i < 30; i++ {
		switch _, err := cli.PutObjectWithContext(ctx, input); err {
		case nil, context.Canceled:
			return
		default:
			if baseErr, ok := err.(awserr.Error); ok && baseErr.Code() == "RequestCanceled" {
				return
			}
			// TODO: logging
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (c *s3Destination) Run(ctx context.Context) {
	size := int(*c.conf.BufferingHints.SizeInMBs * 1024 * 1024)
	c.captured = make([]byte, 0, size)
	dur := time.Duration(*c.conf.BufferingHints.IntervalInSeconds) * time.Second
	tick := time.Tick(dur)
	resource := storeToS3Resource{
		deliveryName:       c.deliveryName,
		bucketName:         strings.ReplaceAll(*c.conf.BucketARN, "arn:aws:s3:::", ""),
		prefix:             *c.conf.Prefix,
		shouldGZipCompress: c.conf.CompressionFormat != nil && *c.conf.CompressionFormat == "GZIP",
	}
	for {
		select {
		case <-ctx.Done():
			newCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			storeToS3(newCtx, resource, time.Now(), c.captured)
			return
		case r := <-c.source:
			c.captured = append(c.captured, r...)
			if len(c.captured) >= size {
				storeToS3(ctx, resource, time.Now(), c.captured)
				c.captured = make([]byte, 0, size)
				// reset timer
				tick = time.Tick(dur)
			}
		case <-tick:
			storeToS3(ctx, resource, time.Now(), c.captured)
			c.captured = make([]byte, 0, size)
		}
	}
}
