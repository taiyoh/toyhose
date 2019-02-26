package driver

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
	dRepo := NewDestinationMemory()
	ctx := context.Background()
	wg := &sync.WaitGroup{}
	wg.Add(10)
	for i := 1; i <= 10; i++ {
		go func(idx int) {
			defer wg.Done()
			time.Sleep(1 * time.Millisecond)
			id := firehose.DestinationID(fmt.Sprintf("destID-%01d", dRepo.DispenceSequence(ctx)))
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
}
