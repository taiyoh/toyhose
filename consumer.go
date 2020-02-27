package toyhose

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
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

var (
	consumers map[int64]consumer
	idSeed    int64
)

func newID() int64 {
	atomic.AddInt64(&idSeed, 1)
	return idSeed
}

func addConsumer(c consumer) {
	consumers[c.ID()] = c
}

type extendedS3Consumer struct {
	id           int64
	deliveryName string
	source       <-chan *firehose.Record
	conf         *firehose.ExtendedS3DestinationConfiguration
	captured     []byte
}

func (c *extendedS3Consumer) ID() int64 {
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
	s3cli = s3.New(session.New(awsConfig().WithEndpoint(endpoint)))
	return s3cli
}

type storeToS3Resource struct {
	deliveryName       string
	arn                string
	prefix             string
	shouldGZipCompress bool
	ts                 time.Time
}

func (r storeToS3Resource) objectName() string {
	return fmt.Sprintf("%s-1-%s-%s", r.deliveryName, r.ts.Format("2006-01-02-15-04-05"), uuid.New())
}

func (r storeToS3Resource) keyName() string {
	prefix := r.prefix
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}
	return fmt.Sprintf("%s%s", prefix, r.objectName())
}

func storeToS3(ctx context.Context, resource storeToS3Resource, data []byte) {
	resource.ts = time.Now()
	bucketName := strings.Trim(resource.arn, "arn:aws:s3:::")
	var seekable []byte
	if resource.shouldGZipCompress {
		b := bytes.NewBuffer([]byte{})
		gzip.NewWriter(b).Write(data)
		seekable = b.Bytes()
	} else {
		seekable = data
	}
	key := resource.keyName()
	input := &s3.PutObjectInput{
		Bucket: &bucketName,
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

func (c *extendedS3Consumer) Run(ctx context.Context) {
	size := int(*c.conf.BufferingHints.SizeInMBs * 1024 * 1024)
	c.captured = make([]byte, 0, size)
	tick := time.Tick(time.Duration(*c.conf.BufferingHints.IntervalInSeconds) * time.Second)
	resource := storeToS3Resource{
		deliveryName:       c.deliveryName,
		arn:                *c.conf.BucketARN,
		prefix:             *c.conf.Prefix,
		shouldGZipCompress: *c.conf.CompressionFormat == "GZIP",
	}
	for {
		select {
		case <-ctx.Done():
			go storeToS3(ctx, resource, c.captured)
			return
		case r := <-c.source:
			b, _ := base64.StdEncoding.DecodeString(string(r.Data))
			c.captured = append(c.captured, b...)
			if len(c.captured) >= size {
				go storeToS3(ctx, resource, c.captured)
				c.captured = make([]byte, 0, size)
				// reset timer
				tick = time.Tick(time.Duration(*c.conf.BufferingHints.IntervalInSeconds) * time.Second)
			}
		case <-tick:
			go storeToS3(ctx, resource, c.captured)
			c.captured = make([]byte, 0, size)
		}
	}
}

type consumer interface {
	ID() int64
	Run(context.Context)
}
