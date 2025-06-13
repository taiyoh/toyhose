package toyhose

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/firehose/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type s3Destination struct {
	deliveryName      string
	bucketARN         string
	bufferingHints    *types.BufferingHints
	compressionFormat types.CompressionFormat
	errorOutputPrefix *string
	prefix            *string
	captured          []*deliveryRecord
	awsConf           aws.Config
	injectedConf      S3InjectedConf
	capturedSize      int
}

func s3Client(conf aws.Config, endpoint string) *s3.Client {
	return s3.NewFromConfig(conf, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})
}

type s3StoreConfig struct {
	deliveryName       string
	bucketName         string
	prefix             string
	shouldGZipCompress bool
	s3cli              *s3.Client
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
		_, _ = w.Write(data)
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
	log.Debug().Str("key", key).Int("size", len(seekable)).Msg("PutObject start")
	cli := conf.s3cli
	for i := 0; i < 30; i++ {
		if _, err := cli.PutObject(ctx, input); err == nil {
			log.Debug().Str("key", key).Msgf("PutObject succeeded. trial count: %d", i+1)
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	log.Debug().Str("key", key).Msg("PutObject failed")
}

func (c *s3Destination) bufferSizeInMBs() int32 {
	// https://docs.aws.amazon.com/firehose/latest/APIReference/API_BufferingHints.html
	// > The default value is 5.
	size := int32(5)
	if c.bufferingHints != nil && c.bufferingHints.SizeInMBs != nil {
		size = *c.bufferingHints.SizeInMBs
	}
	if c.injectedConf.SizeInMBs != nil {
		size = int32(*c.injectedConf.SizeInMBs)
	}
	return size
}

func (c *s3Destination) bufferIntervalSeconds() int32 {
	// https://docs.aws.amazon.com/firehose/latest/APIReference/API_BufferingHints.html
	// > The default value is 300.
	dur := int32(300)
	if c.bufferingHints != nil && c.bufferingHints.IntervalInSeconds != nil {
		dur = *c.bufferingHints.IntervalInSeconds
	}
	if c.injectedConf.IntervalInSeconds != nil {
		dur = int32(*c.injectedConf.IntervalInSeconds)
	}
	return dur
}

func (c *s3Destination) Setup(ctx context.Context) (s3StoreConfig, error) {
	s3cli := s3Client(c.awsConf, *c.injectedConf.EndPoint)
	bucketName := strings.ReplaceAll(c.bucketARN, "arn:aws:s3:::", "")
	if bucketName == "" {
		return s3StoreConfig{}, errors.New("required bucket_name")
	}
	if _, err := s3cli.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: &bucketName,
	}); err != nil {
		return s3StoreConfig{}, err
	}
	prefix := ""
	if c.prefix != nil {
		prefix = *c.prefix
	}
	conf := s3StoreConfig{
		deliveryName:       c.deliveryName,
		bucketName:         bucketName,
		prefix:             prefix,
		shouldGZipCompress: c.compressionFormat == types.CompressionFormatGzip,
		s3cli:              s3cli,
		bufferSize:         int(c.bufferSizeInMBs()) * 1024 * 1024,
		tickDuration:       time.Duration(c.bufferIntervalSeconds()) * time.Second,
	}
	return conf, nil
}

func (c *s3Destination) reset() {
	c.captured = make([]*deliveryRecord, 0, 2048)
	c.capturedSize = 0
}

func (c *s3Destination) finalize(conf s3StoreConfig) {
	newCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	storeToS3(newCtx, conf, time.Now(), c.captured)
}

func (c *s3Destination) Run(ctx context.Context, conf s3StoreConfig, recordCh chan *deliveryRecord) {
	c.reset()
	ticker := time.NewTicker(conf.tickDuration)
	for {
		select {
		case <-ctx.Done():
			log.Debug().Msgf("finish S3Destination in deliveryStream:%s", conf.deliveryName)
			c.finalize(conf)
			return
		case r, ok := <-recordCh:
			if !ok {
				log.Debug().Msgf("deliveryStream:%s is deleted", conf.deliveryName)
				c.finalize(conf)
				return
			}
			c.captured = append(c.captured, r)
			c.capturedSize += len(r.data)
			log.Debug().Int("current", c.capturedSize).Int("limit", conf.bufferSize).Msgf("data captured. size: %d", len(r.data))
			if c.injectedConf.DisableBuffering || c.capturedSize >= conf.bufferSize {
				storeToS3(ctx, conf, time.Now(), c.captured)
				c.reset()
				ticker.Reset(conf.tickDuration)
			}
		case <-ticker.C:
			storeToS3(ctx, conf, time.Now(), c.captured)
			c.reset()
		}
	}
}
