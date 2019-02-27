package gateway_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/taiyoh/toyhose/datatypes/arn"
	"github.com/taiyoh/toyhose/datatypes/firehose"
	"github.com/taiyoh/toyhose/gateway"
)

func TestDeliveryStreamMemory(t *testing.T) {
	dRepo := gateway.NewDeliveryStream()

	ctx := context.Background()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	dsARN := arn.NewDeliveryStream("foo", "bar", "baz")
	go func() {
		defer wg.Done()
		ds, _ := firehose.NewDeliveryStream(dsARN, "DirectPut")
		dRepo.Save(ctx, ds)
	}()
	wg.Wait()

	ds := dRepo.Find(ctx, arn.NewDeliveryStream("hoge", "fuga", "piyo"))
	if ds != nil {
		t.Error("DeliveryStream object should not be returned")
	}
	ds = dRepo.Find(ctx, dsARN)
	if ds == nil {
		t.Error("DeliveryStream object should be returned")
	}
	activated := ds.Active()
	dRepo.Save(ctx, activated)

	if d := dRepo.Find(ctx, dsARN); d != activated {
		t.Error("DeliveryStream object not replaced")
	}
}

func TestDeliveryStreamMemoryFindMulti(t *testing.T) {
	dRepo := gateway.NewDeliveryStream()
	ctx := context.Background()
	for i := 1; i <= 10; i++ {
		ds, _ := firehose.NewDeliveryStream(
			arn.NewDeliveryStream("toyhose", "taiyoh", fmt.Sprintf("test-%01d", i)),
			"DirectPut",
		)
		dRepo.Save(ctx, ds)
	}

	t.Run("wrong account or region arn", func(t *testing.T) {
		a := arn.NewDeliveryStream("foo", "bar", "baz")
		streams, hasNext := dRepo.FindMulti(ctx, a, 3)
		if l := len(streams); l != 0 {
			t.Errorf("returned streams size should be empty: %d", l)
		}
		if hasNext {
			t.Error("hasNext should be false")
		}
	})

	var nextArn arn.DeliveryStream

	t.Run("page 1", func(t *testing.T) {
		a := arn.NewDeliveryStream("toyhose", "taiyoh", "*")
		streams, hasNext := dRepo.FindMulti(ctx, a, 3)
		if l := len(streams); l != 3 {
			t.Errorf("returned streams size is wrong: %d", l)
		}
		if !hasNext {
			t.Error("hasNext should be true")
		}
		for i, s := range streams {
			i := i + 1
			expected := arn.NewDeliveryStream("toyhose", "taiyoh", fmt.Sprintf("test-%01d", i))
			if s.ARN != expected {
				t.Errorf(`wrong order: expected=%s actual=%s`, expected.Code(), s.ARN.Code())
			}
		}
		nextArn = streams[2].ARN
	})

	t.Run("page 2", func(t *testing.T) {
		streams, hasNext := dRepo.FindMulti(ctx, nextArn, 3)
		if l := len(streams); l != 3 {
			t.Errorf("returned streams size is wrong: %d", l)
		}
		if !hasNext {
			t.Error("hasNext should be true")
		}

		for i, s := range streams {
			i := i + 4
			expected := arn.NewDeliveryStream("toyhose", "taiyoh", fmt.Sprintf("test-%01d", i))
			if s.ARN != expected {
				t.Errorf(`wrong order: expected=%s actual=%s`, expected.Code(), s.ARN.Code())
			}
		}
		nextArn = streams[2].ARN
	})

	t.Run("page 3", func(t *testing.T) {
		streams, hasNext := dRepo.FindMulti(ctx, nextArn, 5)
		if l := len(streams); l != 4 {
			t.Errorf("returned streams size is wrong: %d", l)
		}
		if hasNext {
			t.Error("hasNext should be false")
		}

		nextArn = streams[3].ARN
	})

	t.Run("page 4(last ARN is assigned)", func(t *testing.T) {
		streams, hasNext := dRepo.FindMulti(ctx, nextArn, 5)
		if l := len(streams); l != 0 {
			t.Errorf("returned streams size should be empty: %d", l)
		}
		if hasNext {
			t.Error("hasNext should be false")
		}
	})
}
