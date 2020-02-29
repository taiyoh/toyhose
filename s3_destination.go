package toyhose

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
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
	id           int64
	deliveryName string
	source       <-chan *firehose.Record
	conf         *firehose.S3DestinationConfiguration
	captured     []byte
}

func (c *s3Destination) ID() int64 {
	return c.id
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
			// TODO: logging
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (c *s3Destination) Run(ctx context.Context) {
	size := int(*c.conf.BufferingHints.SizeInMBs * 1024 * 1024)
	c.captured = make([]byte, 0, size)
	tick := time.Tick(time.Duration(*c.conf.BufferingHints.IntervalInSeconds) * time.Second)
	resource := storeToS3Resource{
		deliveryName:       c.deliveryName,
		bucketName:         strings.Trim(*c.conf.BucketARN, "arn:aws:s3:::"),
		prefix:             *c.conf.Prefix,
		shouldGZipCompress: *c.conf.CompressionFormat == "GZIP",
	}
	for {
		select {
		case <-ctx.Done():
			storeToS3(ctx, resource, time.Now(), c.captured)
			return
		case r := <-c.source:
			b, _ := base64.StdEncoding.DecodeString(string(r.Data))
			c.captured = append(c.captured, b...)
			if len(c.captured) >= size {
				go storeToS3(ctx, resource, time.Now(), c.captured)
				c.captured = make([]byte, 0, size)
				// reset timer
				tick = time.Tick(time.Duration(*c.conf.BufferingHints.IntervalInSeconds) * time.Second)
			}
		case <-tick:
			go storeToS3(ctx, resource, time.Now(), c.captured)
			c.captured = make([]byte, 0, size)
		}
	}
}

type destination interface {
	ID() int64
	Run(context.Context)
}
