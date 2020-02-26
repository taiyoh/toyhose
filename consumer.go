package toyhose

import (
	"bytes"
	"compress/gzip"
	"context"
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
	id       int64
	source   <-chan []byte
	conf     *firehose.ExtendedS3DestinationConfiguration
	captured []byte
}

func (c *extendedS3Consumer) ID() int64 {
	return c.id
}

var s3cli *s3.S3

func s3Client() *s3.S3 {
	if s3cli != nil {
		return s3cli
	}
	endpoint := os.Getenv("S3_ENDPOINT_URL")
	if endpoint == "" {
		panic("require S3_ENDPOINT_URL")
	}
	conf := aws.NewConfig().
		WithRegion(os.Getenv("AWS_DEFAULT_REGION")).
		WithEndpoint(endpoint).
		WithCredentials(credentials.NewEnvCredentials())
	s3cli = s3.New(session.New(conf))
	return s3cli
}

func (c *extendedS3Consumer) dispatch(ctx context.Context, data []byte) {
	bucketName := strings.Trim(*c.conf.BucketARN, "arn:aws:s3:::")
	cli := s3Client()
	var seekable []byte
	if *c.conf.CompressionFormat == "GZIP" {
		b := bytes.NewBuffer([]byte{})
		gzip.NewWriter(b).Write(data)
		seekable = b.Bytes()
	} else {
		seekable = data
	}
	key := fmt.Sprintf("%s%s", *c.conf.Prefix, uuid.New())
	cli.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: &bucketName,
		Body:   bytes.NewReader(seekable),
		Key:    &key,
	})
}

func (c *extendedS3Consumer) Run(ctx context.Context) {
	size := int(*c.conf.BufferingHints.SizeInMBs * 1024 * 1024)
	c.captured = make([]byte, 0, size)
	tick := time.Tick(time.Duration(*c.conf.BufferingHints.IntervalInSeconds) * time.Second)
	for {
		select {
		case <-ctx.Done():
			c.dispatch(ctx, c.captured)
			return
		case b := <-c.source:
			c.captured = append(c.captured, b...)
			if len(c.captured) >= size {
				go c.dispatch(ctx, c.captured)
				c.captured = make([]byte, 0, size)
			}
		case <-tick:
			go c.dispatch(ctx, c.captured)
			c.captured = make([]byte, 0, size)
		}
	}
}

type consumer interface {
	ID() int64
	Run(context.Context)
}
