package toyhose

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-sdk-go/service/kinesis"
)

var (
	streamARNRE         = regexp.MustCompile(`^arn:aws:kinesis:(.+?):(.+?):stream/(.+)$`)
	errInvalidStreamARN = errors.New("invalid arn")
)

func getStreamNameFromARN(conf *aws.Config, arn string) (string, error) {
	// sample: arn:aws:kinesis:region:account-id:stream/stream-name
	matches := streamARNRE.FindStringSubmatch(arn)
	if len(matches) != 4 {
		return "", errInvalidStreamARN
	}
	cred, _ := conf.Credentials.Get()
	if matches[1] != *conf.Region || matches[2] != cred.AccessKeyID {
		return "", errInvalidSignature
	}
	return matches[3], nil
}

func newKinesisConsumer(ctx context.Context, conf *aws.Config, sourceConf *firehose.KinesisStreamSourceConfiguration, injectConf KinesisInjectedConf) (*kinesisConsumer, error) {
	arn := sourceConf.KinesisStreamARN
	if arn == nil {
		return nil, awserr.New(firehose.ErrCodeInvalidArgumentException, "StreamARN not found", errInvalidStreamARN)
	}
	streamName, err := getStreamNameFromARN(conf, *arn)
	if err != nil {
		return nil, awserr.New(firehose.ErrCodeInvalidArgumentException, "invalid StreamARN", err)
	}
	endpoint := injectConf.Endpoint
	if endpoint == nil {
		return nil, awserr.New(firehose.ErrCodeInvalidArgumentException, "KINESIS_STREAM_ENDPOINT_URL not found", errors.New("endpoint not found"))
	}
	cli := kinesis.New(session.Must(session.NewSession(conf.Copy().WithEndpoint(*endpoint))))
	out, err := cli.DescribeStreamWithContext(ctx, &kinesis.DescribeStreamInput{
		StreamName: &streamName,
	})
	if err != nil {
		return nil, awserr.New(firehose.ErrCodeServiceUnavailableException, "unable to connect Kinesis Streams", err)
	}
	desc := out.StreamDescription
	if *desc.StreamStatus != "ACTIVE" {
		return nil, awserr.New(firehose.ErrCodeServiceUnavailableException, "unable to connect Kinesis Streams", fmt.Errorf("stream status is %s", *desc.StreamStatus))
	}
	shardIter := map[string]string{}
	for _, shard := range desc.Shards {
		out, err := cli.GetShardIteratorWithContext(ctx, &kinesis.GetShardIteratorInput{
			StreamName:        &streamName,
			ShardId:           shard.ShardId,
			ShardIteratorType: aws.String("TRIM_HORIZON"),
		})
		if err != nil {
			return nil, awserr.New(firehose.ErrCodeServiceUnavailableException, "failed to get shard iterator", err)
		}
		shardIter[*shard.ShardId] = *out.ShardIterator
	}
	return &kinesisConsumer{
		streamName: streamName,
		cli:        cli,
		shardIter:  shardIter,
	}, nil
}

type kinesisConsumer struct {
	streamName string
	cli        *kinesis.Kinesis
	shardIter  map[string]string
}

func (c *kinesisConsumer) Run(ctx context.Context, recordCh chan *deliveryRecord) {
	log.Debug().Str("stream_name", c.streamName).Msg("starting to subscribe stream")
	wg := &sync.WaitGroup{}
	wg.Add(len(c.shardIter))
	for shardID, iter := range c.shardIter {
		go func(shardID, iter string) {
			defer wg.Done()
			for {
				out, err := c.cli.GetRecordsWithContext(ctx, &kinesis.GetRecordsInput{
					ShardIterator: &iter,
				})
				if err != nil {
					switch ae, ok := err.(awserr.Error); {
					case ok && ae.Code() == "RequestCanceled", err == context.Canceled:
						log.Debug().Str("shard_id", shardID).Msg("context canceled")
						return
					default:
						log.Debug().Str("shard_id", shardID).Err(err).Msg("GetRecord error")
						time.Sleep(time.Second)
						continue
					}
				}
				if l := len(out.Records); l > 0 {
					log.Debug().Str("shard_id", shardID).Msgf("captured %d records", l)
				}
				for _, record := range out.Records {
					recordCh <- newDeliveryRecord(record.Data)
				}
				iter = *out.NextShardIterator
				time.Sleep(time.Second)
			}
		}(shardID, iter)
	}
	wg.Wait()

	log.Debug().Str("stream_name", c.streamName).Msg("ended to subscribe stream")
}
