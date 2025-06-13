package toyhose

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/firehose"
	"github.com/aws/aws-sdk-go-v2/service/firehose/types"
)

// DeliveryStreamService represents interface for operating DeliveryStream resources.
type DeliveryStreamService struct {
	awsConf             aws.Config
	region              string
	accountID           string
	s3InjectedConf      S3InjectedConf
	kinesisInjectedConf KinesisInjectedConf
	pool                *deliveryStreamPool
}

// Custom type for JSON marshaling
type DeliveryStreamDescriptionForJSON struct {
	CreateTimestamp      int64                          `json:"CreateTimestamp"`
	DeliveryStreamARN    *string                        `json:"DeliveryStreamARN"`
	DeliveryStreamStatus types.DeliveryStreamStatus     `json:"DeliveryStreamStatus"`
	DeliveryStreamName   *string                        `json:"DeliveryStreamName"`
	DeliveryStreamType   types.DeliveryStreamType       `json:"DeliveryStreamType"`
	Destinations         []types.DestinationDescription `json:"Destinations"`
	Source               *types.SourceDescription       `json:"Source"`
	VersionId            *string                        `json:"VersionId"`
}

type DescribeDeliveryStreamOutputForJSON struct {
	DeliveryStreamDescription DeliveryStreamDescriptionForJSON `json:"DeliveryStreamDescription"`
}

func (s *DeliveryStreamService) arnName(streamName string) string {
	return fmt.Sprintf("arn:aws:firehose:%s:%s:deliverystream/%s", s.region, s.accountID, streamName)
}

// Create provides creating DeliveryStream resource operation.
func (s *DeliveryStreamService) Create(ctx context.Context, input []byte) (*firehose.CreateDeliveryStreamOutput, error) {
	i := &firehose.CreateDeliveryStreamInput{}
	if err := json.Unmarshal(input, i); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}
	// i.Validate() is not available in v2, perform manual validation if needed
	arn := s.arnName(*i.DeliveryStreamName)
	dsCtx, dsCancel := context.WithCancel(context.Background())
	recordCh := make(chan *deliveryRecord, 128)
	dsType := types.DeliveryStreamTypeDirectPut
	if i.DeliveryStreamType != "" {
		dsType = i.DeliveryStreamType
	}
	ds := &deliveryStream{
		arn:                arn,
		deliveryStreamName: *i.DeliveryStreamName,
		deliveryStreamType: dsType,
		recordCh:           recordCh,
		closer:             dsCancel,
		destDesc:           &types.DestinationDescription{},
		createdAt:          time.Now(),
	}
	if i.S3DestinationConfiguration != nil {
		s3dest := &s3Destination{
			deliveryName:      *i.DeliveryStreamName,
			bucketARN:         *i.S3DestinationConfiguration.BucketARN,
			bufferingHints:    i.S3DestinationConfiguration.BufferingHints,
			compressionFormat: i.S3DestinationConfiguration.CompressionFormat,
			errorOutputPrefix: i.S3DestinationConfiguration.ErrorOutputPrefix,
			prefix:            i.S3DestinationConfiguration.Prefix,
			injectedConf:      s.s3InjectedConf,
			awsConf:           s.awsConf,
		}
		conf, err := s3dest.Setup(dsCtx)
		if err != nil {
			return nil, &types.ResourceNotFoundException{Message: aws.String("invalid BucketName")}
		}
		ds.destDesc.S3DestinationDescription = &types.S3DestinationDescription{
			BucketARN:               i.S3DestinationConfiguration.BucketARN,
			BufferingHints:          i.S3DestinationConfiguration.BufferingHints,
			CompressionFormat:       i.S3DestinationConfiguration.CompressionFormat,
			EncryptionConfiguration: i.S3DestinationConfiguration.EncryptionConfiguration,
			ErrorOutputPrefix:       i.S3DestinationConfiguration.ErrorOutputPrefix,
			Prefix:                  i.S3DestinationConfiguration.Prefix,
			RoleARN:                 i.S3DestinationConfiguration.RoleARN,
		}
		go s3dest.Run(dsCtx, conf, recordCh)
	}
	if ds.deliveryStreamType == types.DeliveryStreamTypeKinesisStreamAsSource && i.KinesisStreamSourceConfiguration != nil {
		consumer, err := newKinesisConsumer(ctx, s.awsConf, i.KinesisStreamSourceConfiguration, s.kinesisInjectedConf)
		if err != nil {
			ds.Close()
			return nil, err
		}
		ds.sourceDesc = &types.SourceDescription{
			KinesisStreamSourceDescription: &types.KinesisStreamSourceDescription{
				DeliveryStartTimestamp: &ds.createdAt,
				KinesisStreamARN:       i.KinesisStreamSourceConfiguration.KinesisStreamARN,
				RoleARN:                i.KinesisStreamSourceConfiguration.RoleARN,
			},
		}
		go consumer.Run(dsCtx, recordCh)
	}
	s.pool.Add(ds)
	output := &firehose.CreateDeliveryStreamOutput{
		DeliveryStreamARN: &arn,
	}
	return output, nil
}

// Delete provides deleting DeliveryStream resource operation.
func (s *DeliveryStreamService) Delete(ctx context.Context, input []byte) (*firehose.DeleteDeliveryStreamOutput, error) {
	i := &firehose.DeleteDeliveryStreamInput{}
	if err := json.Unmarshal(input, i); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}
	arn := s.arnName(*i.DeliveryStreamName)
	ds := s.pool.Delete(arn)
	if ds == nil {
		return nil, &types.ResourceNotFoundException{Message: aws.String(fmt.Sprintf("DeliveryStreamName: %s not found", *i.DeliveryStreamName))}
	}
	ds.Close()
	return &firehose.DeleteDeliveryStreamOutput{}, nil
}

// Put provides accepting single record data for sending to DeliveryStream.
func (s *DeliveryStreamService) Put(ctx context.Context, input []byte) (*firehose.PutRecordOutput, error) {
	i := &firehose.PutRecordInput{}
	if err := json.Unmarshal(input, i); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}
	ds := s.pool.Find(s.arnName(*i.DeliveryStreamName))
	if ds == nil {
		return nil, &types.ResourceNotFoundException{Message: aws.String(fmt.Sprintf("DeliveryStreamName: %s not found", *i.DeliveryStreamName))}
	}
	log.Debug().Str("delivery_stream", *i.DeliveryStreamName).Msg("processing PutRecord request")
	recordIDs := putData(ds, []types.Record{*i.Record})
	output := &firehose.PutRecordOutput{
		Encrypted: aws.Bool(false),
		RecordId:  &recordIDs[0],
	}
	return output, nil
}

func putData(ds *deliveryStream, records []types.Record) []string {
	recordIDs := make([]string, 0, len(records))
	for _, record := range records {
		dst, err := base64.StdEncoding.DecodeString(string(record.Data))
		if err != nil {
			dst = record.Data
		}
		rec := newDeliveryRecord(dst)
		ds.recordCh <- rec
		recordIDs = append(recordIDs, rec.id)
	}
	return recordIDs
}

// PutBatch provides accepting multiple record data for sending to DeliveryStream.
func (s *DeliveryStreamService) PutBatch(ctx context.Context, input []byte) (*firehose.PutRecordBatchOutput, error) {
	i := &firehose.PutRecordBatchInput{}
	if err := json.Unmarshal(input, i); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}
	ds := s.pool.Find(s.arnName(*i.DeliveryStreamName))
	if ds == nil {
		return nil, &types.ResourceNotFoundException{Message: aws.String(fmt.Sprintf("DeliveryStreamName: %s not found", *i.DeliveryStreamName))}
	}
	log.Debug().Str("delivery_stream", *i.DeliveryStreamName).Msgf("processing PutRecordBatch request for %d records", len(i.Records))
	recordIDs := putData(ds, i.Records)
	output := &firehose.PutRecordBatchOutput{
		FailedPutCount: aws.Int32(0),
		Encrypted:      aws.Bool(false),
	}
	for _, r := range recordIDs {
		output.RequestResponses = append(output.RequestResponses, types.PutRecordBatchResponseEntry{
			RecordId: aws.String(r),
		})
	}
	return output, nil
}

// Listing returns registered deliveryStream names.
func (s *DeliveryStreamService) Listing(ctx context.Context, input []byte) (*firehose.ListDeliveryStreamsOutput, error) {
	i := &firehose.ListDeliveryStreamsInput{}
	if err := json.Unmarshal(input, i); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}

	streams, hasNext := s.pool.FindAllBySource(i.DeliveryStreamType, i.ExclusiveStartDeliveryStreamName, i.Limit)
	responses := make([]string, 0, len(streams))
	for _, ds := range streams {
		responses = append(responses, ds.deliveryStreamName)
	}

	out := &firehose.ListDeliveryStreamsOutput{
		DeliveryStreamNames:    responses,
		HasMoreDeliveryStreams: aws.Bool(hasNext),
	}

	return out, nil
}

// Describe returns current deliveryStream definitions and statuses by supplied deliveryStreamName.
func (s *DeliveryStreamService) Describe(ctx context.Context, input []byte) (*DescribeDeliveryStreamOutputForJSON, error) {
	i := &firehose.DescribeDeliveryStreamInput{}
	if err := json.Unmarshal(input, i); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}
	ds := s.pool.Find(s.arnName(*i.DeliveryStreamName))
	if ds == nil {
		return nil, &types.ResourceNotFoundException{Message: aws.String(fmt.Sprintf("DeliveryStreamName: %s not found", *i.DeliveryStreamName))}
	}

	out := &DescribeDeliveryStreamOutputForJSON{
		DeliveryStreamDescription: DeliveryStreamDescriptionForJSON{
			CreateTimestamp:      ds.createdAt.Unix(),
			DeliveryStreamARN:    &ds.arn,
			DeliveryStreamStatus: types.DeliveryStreamStatusActive,
			DeliveryStreamName:   &ds.deliveryStreamName,
			DeliveryStreamType:   ds.deliveryStreamType,
			Destinations:         []types.DestinationDescription{*ds.destDesc},
			Source:               ds.sourceDesc,
			VersionId:            aws.String("1"),
		},
	}

	return out, nil
}
