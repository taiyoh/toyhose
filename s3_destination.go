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

type s3Destination struct {
	deliveryName   string
	source         <-chan *deliveryRecord
	conf           *firehose.S3DestinationConfiguration
	closer         context.CancelFunc
	captured       []*deliveryRecord
	bufferingHints S3BufferingHints
	capturedSize   int
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
		WithRegion(os.Getenv("AWS_REGION")).
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

type storeToS3Config struct {
	deliveryName       string
	bucketName         string
	prefix             string
	shouldGZipCompress bool
}

func storeToS3(ctx context.Context, conf storeToS3Config, ts time.Time, records []*deliveryRecord) {
	data := make([]byte, 0, 1024*1024)
	for _, rec := range records {
		data = append(data, rec.data...)
	}
	if len(data) < 1 {
		return
	}
	var seekable []byte
	if conf.shouldGZipCompress {
		b := bytes.NewBuffer([]byte{})
		w := gzip.NewWriter(b)
		w.Write(data)
		w.Close()
		seekable = b.Bytes()
	} else {
		seekable = data
	}
	pref := strings.TrimSuffix(keyPrefix(conf.prefix, ts), "/")
	key := fmt.Sprintf("%s/%s-1-%s-%s", pref, conf.deliveryName, ts.Format("2006-01-02-15-04-05"), uuid.New())
	input := &s3.PutObjectInput{
		Bucket: &conf.bucketName,
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

func (c *s3Destination) setup() (int, time.Duration) {
	size := int(*c.conf.BufferingHints.SizeInMBs * 1024 * 1024)
	if c.bufferingHints.SizeInMBs != nil {
		size = *c.bufferingHints.SizeInMBs * 1024 * 1024
	}
	dur := int(*c.conf.BufferingHints.IntervalInSeconds)
	if c.bufferingHints.IntervalInSeconds != nil {
		dur = *c.bufferingHints.IntervalInSeconds
	}
	return size, time.Duration(dur) * time.Second
}

func (c *s3Destination) reset(dur time.Duration) <-chan time.Time {
	c.captured = make([]*deliveryRecord, 0, 2048)
	c.capturedSize = 0
	return time.Tick(dur)
}

func (c *s3Destination) finalize(conf storeToS3Config) {
	newCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	storeToS3(newCtx, conf, time.Now(), c.captured)
}

func (c *s3Destination) Run(ctx context.Context) {
	size, dur := c.setup()
	conf := storeToS3Config{
		deliveryName:       c.deliveryName,
		bucketName:         strings.ReplaceAll(*c.conf.BucketARN, "arn:aws:s3:::", ""),
		prefix:             *c.conf.Prefix,
		shouldGZipCompress: c.conf.CompressionFormat != nil && *c.conf.CompressionFormat == "GZIP",
	}
	tick := c.reset(dur)
	for {
		select {
		case <-ctx.Done():
			c.finalize(conf)
			return
		case r, ok := <-c.source:
			if !ok {
				c.finalize(conf)
				return
			}
			c.captured = append(c.captured, r)
			c.capturedSize += len(r.data)
			if c.capturedSize >= size {
				storeToS3(ctx, conf, time.Now(), c.captured)
				tick = c.reset(dur)
			}
		case <-tick:
			storeToS3(ctx, conf, time.Now(), c.captured)
			tick = c.reset(dur)
		}
	}
}

func (c *s3Destination) Close() {
	c.closer()
}
