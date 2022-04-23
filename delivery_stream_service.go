package toyhose

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/firehose"
)

// DeliveryStreamService represents interface for operating DeliveryStream resources.
type DeliveryStreamService struct {
	awsConf             *aws.Config
	region              string
	accountID           string
	s3InjectedConf      S3InjectedConf
	kinesisInjectedConf KinesisInjectedConf
	pool                *deliveryStreamPool
}

func (s *DeliveryStreamService) arnName(streamName string) string {
	return fmt.Sprintf("arn:aws:firehose:%s:%s:deliverystream/%s", s.region, s.accountID, streamName)
}

// Create provides creating DeliveryStream resource operation.
func (s *DeliveryStreamService) Create(ctx context.Context, input []byte) (*firehose.CreateDeliveryStreamOutput, error) {
	i := &firehose.CreateDeliveryStreamInput{}
	if err := json.Unmarshal(input, i); err != nil {
		return nil, awserr.NewUnmarshalError(err, "Unmarshal error", input)
	}
	if err := i.Validate(); err != nil {
		return nil, err
	}
	arn := s.arnName(*i.DeliveryStreamName)
	dsCtx, dsCancel := context.WithCancel(context.Background())
	recordCh := make(chan *deliveryRecord, 128)
	dsType := "DirectPut"
	if i.DeliveryStreamType != nil {
		dsType = *i.DeliveryStreamType
	}
	ds := &deliveryStream{
		arn:                arn,
		deliveryStreamName: *i.DeliveryStreamName,
		deliveryStreamType: dsType,
		recordCh:           recordCh,
		closer:             dsCancel,
		destDesc:           &firehose.DestinationDescription{},
		createdAt:          time.Now(),
	}
	//nolint
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
			return nil, awserr.New(firehose.ErrCodeResourceNotFoundException, "invalid BucketName", err)
		}
		ds.destDesc.S3DestinationDescription = &firehose.S3DestinationDescription{
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
	if ds.deliveryStreamType == "KinesisStreamAsSource" && i.KinesisStreamSourceConfiguration != nil {
		consumer, err := newKinesisConsumer(ctx, s.awsConf, i.KinesisStreamSourceConfiguration, s.kinesisInjectedConf)
		if err != nil {
			ds.Close()
			return nil, err
		}
		ds.sourceDesc = &firehose.SourceDescription{
			KinesisStreamSourceDescription: &firehose.KinesisStreamSourceDescription{
				DeliveryStartTimestamp: aws.Time(ds.createdAt),
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
		return nil, awserr.NewUnmarshalError(err, "Unmarshal error", input)
	}
	if err := i.Validate(); err != nil {
		return nil, err
	}
	arn := s.arnName(*i.DeliveryStreamName)
	ds := s.pool.Delete(arn)
	if ds == nil {
		return nil, awserr.New(firehose.ErrCodeResourceNotFoundException, "DeliveryStream not found", fmt.Errorf("DeliveryStreamName: %s not found", *i.DeliveryStreamName))
	}
	ds.Close()
	return &firehose.DeleteDeliveryStreamOutput{}, nil
}

// Put provides accepting single record data for sending to DeliveryStream.
func (s *DeliveryStreamService) Put(ctx context.Context, input []byte) (*firehose.PutRecordOutput, error) {
	i := &firehose.PutRecordInput{}
	if err := json.Unmarshal(input, i); err != nil {
		return nil, awserr.NewUnmarshalError(err, "Unmarshal error", input)
	}
	if err := i.Validate(); err != nil {
		return nil, err
	}
	ds := s.pool.Find(s.arnName(*i.DeliveryStreamName))
	if ds == nil {
		return nil, awserr.New(firehose.ErrCodeResourceNotFoundException, "DeliveryStream not found", fmt.Errorf("DeliveryStreamName: %s not found", *i.DeliveryStreamName))
	}
	log.Debug().Str("delivery_stream", *i.DeliveryStreamName).Msg("processing PutRecord request")
	recordIDs := putData(ds, []*firehose.Record{i.Record})
	output := &firehose.PutRecordOutput{
		Encrypted: aws.Bool(false),
		RecordId:  &recordIDs[0],
	}
	return output, nil
}

func putData(ds *deliveryStream, records []*firehose.Record) []string {
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
		return nil, awserr.NewUnmarshalError(err, "Unmarshal error", input)
	}
	if err := i.Validate(); err != nil {
		return nil, err
	}
	ds := s.pool.Find(s.arnName(*i.DeliveryStreamName))
	if ds == nil {
		return nil, awserr.New(firehose.ErrCodeResourceNotFoundException, "DeliveryStream not found", fmt.Errorf("DeliveryStreamName: %s not found", *i.DeliveryStreamName))
	}
	log.Debug().Str("delivery_stream", *i.DeliveryStreamName).Msgf("processing PutRecordBatch request for %d records", len(i.Records))
	recordIDs := putData(ds, i.Records)
	output := &firehose.PutRecordBatchOutput{
		FailedPutCount: aws.Int64(0),
		Encrypted:      aws.Bool(false),
	}
	for _, r := range recordIDs {
		output.RequestResponses = append(output.RequestResponses, &firehose.PutRecordBatchResponseEntry{
			RecordId: aws.String(r),
		})
	}
	return output, nil
}

// Listing returns registered deliveryStream names.
func (s *DeliveryStreamService) Listing(ctx context.Context, input []byte) (*firehose.ListDeliveryStreamsOutput, error) {
	i := &firehose.ListDeliveryStreamsInput{}
	if err := json.Unmarshal(input, i); err != nil {
		return nil, awserr.NewUnmarshalError(err, "Unmarshal error", input)
	}
	if err := i.Validate(); err != nil {
		return nil, err
	}

	streams, hasNext := s.pool.FindAllBySource(*i.DeliveryStreamType, i.ExclusiveStartDeliveryStreamName, i.Limit)
	responses := make([]*string, 0, len(streams))
	for _, ds := range streams {
		responses = append(responses, &ds.deliveryStreamName)
	}

	out := &firehose.ListDeliveryStreamsOutput{
		DeliveryStreamNames:    responses,
		HasMoreDeliveryStreams: aws.Bool(hasNext),
	}

	return out, nil
}

// Describe returns current deliveryStream definitions and statuses by supplied deliveryStreamName.
func (s *DeliveryStreamService) Describe(ctx context.Context, input []byte) (*firehose.DescribeDeliveryStreamOutput, error) {
	i := &firehose.DescribeDeliveryStreamInput{}
	if err := json.Unmarshal(input, i); err != nil {
		return nil, awserr.NewUnmarshalError(err, "Unmarshal error", input)
	}
	if err := i.Validate(); err != nil {
		return nil, err
	}
	ds := s.pool.Find(s.arnName(*i.DeliveryStreamName))
	if ds == nil {
		return nil, awserr.New(firehose.ErrCodeResourceNotFoundException, "DeliveryStream not found", fmt.Errorf("DeliveryStreamName: %s not found", *i.DeliveryStreamName))
	}

	out := &firehose.DescribeDeliveryStreamOutput{
		DeliveryStreamDescription: &firehose.DeliveryStreamDescription{
			CreateTimestamp:      aws.Time(ds.createdAt),
			DeliveryStreamARN:    &ds.arn,
			DeliveryStreamStatus: aws.String("ACTIVE"),
			DeliveryStreamName:   &ds.deliveryStreamName,
			DeliveryStreamType:   &ds.deliveryStreamType,
			Destinations:         []*firehose.DestinationDescription{ds.destDesc},
			Source:               ds.sourceDesc,
			VersionId:            aws.String("1"),
		},
	}

	return out, nil
}
