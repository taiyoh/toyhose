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
	"github.com/google/uuid"
)

// DeliveryStreamService represents interface for operating DeliveryStream resources.
type DeliveryStreamService struct {
	region    string
	accountID string
	pool      *deliveryStreamPool
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
	source := make(chan *deliveryRecord, 128)
	ds := &deliveryStream{
		arn:       arn,
		source:    source,
		closer:    dsCancel,
		createdAt: time.Now(),
	}
	if i.S3DestinationConfiguration != nil {
		s3DestCtx, s3DestCancel := context.WithCancel(dsCtx)
		s3dest := &s3Destination{
			deliveryName: *i.DeliveryStreamName,
			source:       source,
			conf:         i.S3DestinationConfiguration,
			closer:       s3DestCancel,
		}
		go s3dest.Run(s3DestCtx)
		ds.s3Dest = s3dest
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
		dst := make([]byte, 0, 1024)
		if _, err := base64.StdEncoding.Decode(dst, record.Data); err != nil {
			dst = record.Data
		}
		recID := uuid.New().String()
		ds.source <- &deliveryRecord{id: recID, data: dst}
		recordIDs = append(recordIDs, recID)
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
