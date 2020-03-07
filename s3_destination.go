package toyhose

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

type s3Destination struct {
	deliveryName string
	source       <-chan *deliveryRecord
	conf         *firehose.S3DestinationConfiguration
	closer       context.CancelFunc
	captured     []*deliveryRecord
	awsConf      *aws.Config
	injectedConf S3InjectedConf
	capturedSize int
}

func s3Client(conf *aws.Config, endpoint string) *s3.S3 {
	return s3.New(session.New(conf.WithEndpoint(endpoint).WithS3ForcePathStyle(true).WithDisableSSL(true)))
}

type s3StoreConfig struct {
	deliveryName       string
	bucketName         string
	prefix             string
	shouldGZipCompress bool
	s3cli              *s3.S3
	bufferSize         int // byte
	tickDuration       time.Duration
}

func storeToS3(ctx context.Context, conf s3StoreConfig, ts time.Time, records []*deliveryRecord) {
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
	logger := log.Debug().Str("key", key)
	logger.Int("size", len(seekable)).Msg("PutObject start")
	cli := conf.s3cli
	for i := 0; i < 30; i++ {
		switch _, err := cli.PutObjectWithContext(ctx, input); err {
		case nil:
			logger.Msgf("PutObject succeeded. trial count: %d", i+1)
		case context.Canceled:
			logger.Msg("context.Canceled")
			return
		default:
			if baseErr, ok := err.(awserr.Error); ok && baseErr.Code() == "RequestCanceled" {
				logger.Err(baseErr).Msg("awserr.RequestCanceled")
				return
			}
			// TODO: logging
			time.Sleep(100 * time.Millisecond)
		}
	}
	logger.Msg("PutObject failed")
}

func (c *s3Destination) Setup(ctx context.Context) (s3StoreConfig, error) {
	size := int(*c.conf.BufferingHints.SizeInMBs * 1024 * 1024)
	if c.injectedConf.SizeInMBs != nil {
		size = *c.injectedConf.SizeInMBs * 1024 * 1024
	}
	dur := int(*c.conf.BufferingHints.IntervalInSeconds)
	if c.injectedConf.IntervalInSeconds != nil {
		dur = *c.injectedConf.IntervalInSeconds
	}
	s3cli := s3Client(c.awsConf, *c.injectedConf.EndPoint)
	bucketName := strings.ReplaceAll(*c.conf.BucketARN, "arn:aws:s3:::", "")
	if bucketName == "" {
		return s3StoreConfig{}, errors.New("required bucket_name")
	}
	if _, err := s3cli.HeadBucketWithContext(ctx, &s3.HeadBucketInput{
		Bucket: &bucketName,
	}); err != nil {
		return s3StoreConfig{}, err
	}
	prefix := ""
	if c.conf.Prefix != nil {
		prefix = *c.conf.Prefix
	}
	conf := s3StoreConfig{
		deliveryName:       c.deliveryName,
		bucketName:         bucketName,
		prefix:             prefix,
		shouldGZipCompress: c.conf.CompressionFormat != nil && *c.conf.CompressionFormat == "GZIP",
		s3cli:              s3cli,
		bufferSize:         size,
		tickDuration:       time.Duration(dur) * time.Second,
	}
	return conf, nil
}

func (c *s3Destination) reset(dur time.Duration) <-chan time.Time {
	c.captured = make([]*deliveryRecord, 0, 2048)
	c.capturedSize = 0
	return time.Tick(dur)
}

func (c *s3Destination) finalize(conf s3StoreConfig) {
	newCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	storeToS3(newCtx, conf, time.Now(), c.captured)
}

func (c *s3Destination) Run(ctx context.Context, conf s3StoreConfig) {
	tick := c.reset(conf.tickDuration)
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
			if c.capturedSize >= conf.bufferSize {
				storeToS3(ctx, conf, time.Now(), c.captured)
				tick = c.reset(conf.tickDuration)
			}
		case <-tick:
			storeToS3(ctx, conf, time.Now(), c.captured)
			tick = c.reset(conf.tickDuration)
		}
	}
}

func (c *s3Destination) Close() {
	c.closer()
}
