package actions

import (
	"github.com/taiyoh/toyhose/actions/port"
	"github.com/taiyoh/toyhose/datatypes/actions/deliverystream"
	"github.com/taiyoh/toyhose/datatypes/arn"
	"github.com/taiyoh/toyhose/datatypes/firehose"
	"github.com/taiyoh/toyhose/errors"
)

// DeliveryStream provides application service for deliverystream operation
type DeliveryStream struct {
	dsRepo    DeliveryStreamRepository
	destRepo  DestinationRepository
	region    string
	accountID string
}

// NewDeliveryStream returns DeliveryStream service object
func NewDeliveryStream(dsRepo DeliveryStreamRepository, destRepo DestinationRepository, region, accountID string) *DeliveryStream {
	return &DeliveryStream{dsRepo, destRepo, region, accountID}
}

// Create provides operation for creating new delivery stream
func (d *DeliveryStream) Create(input *port.Input, output *port.Output) {
	ci, err := deliverystream.NewCreateInput(input.Arg())
	if err != nil {
		output.Set(nil, err)
		return
	}
	ctx := input.Ctx()
	dsARN := arn.NewDeliveryStream(d.region, d.accountID, ci.Name)
	if ds := d.dsRepo.Find(ctx, dsARN); ds != nil {
		output.Set(nil, errors.NewResourceInUse(ci.Name))
		return
	}
	ds, _ := firehose.NewDeliveryStream(dsARN, ci.Type)

	d.destRepo.Save(ctx, &firehose.Destination{
		ID:        d.destRepo.DispenseID(ctx),
		SourceARN: ds.ARN,
		S3Conf:    ci.S3Conf,
	})

	ds = ds.Active()
	d.dsRepo.Save(ctx, ds)

	output.Set(&deliverystream.CreateOutput{ARN: ds.ARN.Code()}, nil)
}

func (d *DeliveryStream) Describe(input *port.Input, output *port.Output) {
}

func (d *DeliveryStream) List(input *port.Input, output *port.Output) {
	li, err := deliverystream.NewListInput(input.Arg())
	if err != nil {
		output.Set(nil, err)
		return
	}
	ctx := input.Ctx()
	dsARN := arn.NewDeliveryStream(d.region, d.accountID, li.ExclusiveStartDeliveryStreamName())
	streams, hasNext := d.dsRepo.FindMulti(ctx, dsARN, *li.Limit)

	resource := &deliverystream.ListOutput{
		Names:   make([]string, 0, len(streams)),
		HasNext: hasNext,
	}
	for _, stream := range streams {
		resource.Names = append(resource.Names, stream.ARN.Name())
	}

	output.Set(resource, nil)
}

func (d *DeliveryStream) PutRecord(input *port.Input, output *port.Output) {
}

func (d *DeliveryStream) PutRecordBatch(input *port.Input, output *port.Output) {
}
