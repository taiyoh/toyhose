package toyhose

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	fhtypes "github.com/aws/aws-sdk-go-v2/service/firehose/types"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
)

var (
	streamARNRE         = regexp.MustCompile(`^arn:aws:kinesis:(.+?):(.+?):stream/(.+)$`)
	errInvalidStreamARN = errors.New("invalid arn")
)

func getStreamNameFromARN(conf aws.Config, arn string) (string, error) {
	// sample: arn:aws:kinesis:region:account-id:stream/stream-name
	matches := streamARNRE.FindStringSubmatch(arn)
	if len(matches) != 4 {
		return "", errInvalidStreamARN
	}
	// FIXME: AccountID check
	// cred, _ := conf.Credentials.Retrieve(context.Background())
	// if matches[1] != conf.Region || matches[2] != cred.AccountID {
	// 	return "", errInvalidSignature
	// }
	return matches[3], nil
}

func newKinesisConsumer(ctx context.Context, conf aws.Config, sourceConf *fhtypes.KinesisStreamSourceConfiguration, injectConf KinesisInjectedConf) (*kinesisConsumer, error) {
	arn := sourceConf.KinesisStreamARN
	if arn == nil {
		return nil, &fhtypes.InvalidArgumentException{Message: aws.String("StreamARN not found")}
	}
	streamName, err := getStreamNameFromARN(conf, *arn)
	if err != nil {
		return nil, &fhtypes.InvalidArgumentException{Message: aws.String("invalid StreamARN")}
	}
	endpoint := injectConf.Endpoint
	if endpoint == nil {
		return nil, &fhtypes.InvalidArgumentException{Message: aws.String("KINESIS_STREAM_ENDPOINT_URL not found")}
	}
	cli := kinesis.NewFromConfig(conf, func(o *kinesis.Options) {
		o.BaseEndpoint = endpoint
	})
	out, err := cli.DescribeStream(ctx, &kinesis.DescribeStreamInput{
		StreamName: &streamName,
	})
	if err != nil {
		return nil, &fhtypes.ServiceUnavailableException{Message: aws.String("unable to connect Kinesis Streams")}
	}
	desc := out.StreamDescription
	if desc.StreamStatus != types.StreamStatusActive {
		return nil, &fhtypes.ServiceUnavailableException{Message: aws.String(fmt.Sprintf("stream status is %s", desc.StreamStatus))}
	}
	shardIter := map[string]string{}
	for _, shard := range desc.Shards {
		out, err := cli.GetShardIterator(ctx, &kinesis.GetShardIteratorInput{
			StreamName:        &streamName,
			ShardId:           shard.ShardId,
			ShardIteratorType: types.ShardIteratorTypeTrimHorizon,
		})
		if err != nil {
			return nil, &fhtypes.ServiceUnavailableException{Message: aws.String("failed to get shard iterator")}
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
	cli        *kinesis.Client
	shardIter  map[string]string
}

func (c *kinesisConsumer) Run(ctx context.Context, recordCh chan *deliveryRecord) {
	log.Debug().Str("stream_name", c.streamName).Msg("starting to subscribe stream")
	wg := &sync.WaitGroup{}
	wg.Add(len(c.shardIter))
	for shardID, iter := range c.shardIter {
		go func(shardID, iter string) {
			defer wg.Done()
			currentIter := iter
			for {
				select {
				case <-ctx.Done():
					log.Debug().Str("shard_id", shardID).Msg("context canceled")
					return
				default:
				}

				out, err := c.cli.GetRecords(ctx, &kinesis.GetRecordsInput{
					ShardIterator: &currentIter,
				})
				if err != nil {
					log.Debug().Str("shard_id", shardID).Err(err).Msg("GetRecord error")
					time.Sleep(time.Second)
					continue
				}
				if l := len(out.Records); l > 0 {
					log.Debug().Str("shard_id", shardID).Msgf("captured %d records", l)
				}
				for _, record := range out.Records {
					recordCh <- newDeliveryRecord(record.Data)
				}
				if out.NextShardIterator == nil {
					log.Debug().Str("shard_id", shardID).Msg("NextShardIterator is nil, finishing subscription")
					return
				}
				currentIter = *out.NextShardIterator
				time.Sleep(time.Second)
			}
		}(shardID, iter)
	}
	wg.Wait()

	log.Debug().Str("stream_name", c.streamName).Msg("ended to subscribe stream")
}
