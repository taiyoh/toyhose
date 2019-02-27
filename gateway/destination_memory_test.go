package gateway

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/taiyoh/toyhose/datatypes/arn"
	"github.com/taiyoh/toyhose/datatypes/firehose"
)

func TestDestinationMemory(t *testing.T) {
	dRepo := NewDestination()
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	wg.Add(10)
	for i := 1; i <= 10; i++ {
		go func(idx int) {
			defer wg.Done()
			time.Sleep(1 * time.Millisecond)
			id := dRepo.DispenseID(ctx)
			source := arn.NewDeliveryStream("foo", "bar", fmt.Sprintf("delivery-%d", idx))
			dRepo.Save(ctx, &firehose.Destination{
				ID:        id,
				SourceARN: source,
			})
		}(i)
	}
	wg.Wait()

	if dRepo.sequence != uint64(10) {
		t.Errorf("wrong number of sequence: %d (actual: 10)", dRepo.sequence)
	}

	dst := dRepo.list[6]
	origSource := dst.SourceARN

	dRepo.Save(ctx, &firehose.Destination{
		ID:        dst.ID,
		SourceARN: arn.NewDeliveryStream("hoge", "fuga", "piyo"),
	})

	dst = dRepo.list[6]
	if origSource == dst.SourceARN {
		t.Error("SourceARN not replaced")
	}

	source := arn.NewDeliveryStream("foo", "bar", "delivery-test")

	dests := dRepo.FindBySource(ctx, source)
	if len(dests) != 0 {
		t.Error("already exists!")
	}
	for i := 0; i < 5; i++ {
		dRepo.Save(ctx, &firehose.Destination{
			ID:        dRepo.DispenseID(ctx),
			SourceARN: source,
		})
	}
	dests = dRepo.FindBySource(ctx, source)
	if len(dests) != 5 {
		t.Error("wtf!")
	}
}
