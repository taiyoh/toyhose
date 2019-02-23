package actions

import (
	"context"

	"github.com/taiyoh/toyhose/datatypes/actions/deliverystream"
	"github.com/taiyoh/toyhose/datatypes/firehose"
	"github.com/taiyoh/toyhose/exception"
)

type DeliveryStream struct {
	dsRepo    DeliveryStreamRepository
	destRepo  DestinationRepository
	region    string
	accountID string
}

func NewDeliveryStream(dsRepo DeliveryStreamRepository, destRepo DestinationRepository, region, accountID string) *DeliveryStream {
	return &DeliveryStream{dsRepo, destRepo, region, accountID}
}

func (d *DeliveryStream) Create(ctx context.Context, arg []byte) error {
	input, err := deliverystream.NewCreateInput(d.region, d.accountID, arg)
	if err != nil {
		return err
	}
	if ds := d.dsRepo.Find(ctx, input.ARN()); ds != nil {
		return exception.NewResourceInUse(input.Name)
	}
	ds := input.Entity()

	d.destRepo.Save(ctx, &firehose.Destination{
		ID:        d.destRepo.DispenseID(ctx),
		SourceARN: ds.ARN,
		S3Conf:    input.S3Conf,
	})

	ds = ds.Active()
	d.dsRepo.Save(ctx, ds)
	return nil
}

func (d *DeliveryStream) Describe(ctx context.Context, arg []byte) error {
	return nil
}

func (d *DeliveryStream) List(ctx context.Context, arg []byte) error {
	return nil
}

func (d *DeliveryStream) PutRecord(ctx context.Context, arg []byte) error {
	return nil
}

func (d *DeliveryStream) PutRecordBatch(ctx context.Context, arg []byte) error {
	return nil
}
