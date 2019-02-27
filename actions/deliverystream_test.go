package actions_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/taiyoh/toyhose/actions"
	"github.com/taiyoh/toyhose/actions/port"
	"github.com/taiyoh/toyhose/datatypes/arn"
	"github.com/taiyoh/toyhose/gateway"
)

var (
	region    = "toyhose"
	accountID = "nyanya"
)

func TestCreateDeliveryStream(t *testing.T) {
	dsRepo := gateway.NewDeliveryStream()
	dstRepo := gateway.NewDestination()
	app := actions.NewDeliveryStream(dsRepo, dstRepo, region, accountID)

	t.Run("bad request", func(t *testing.T) {
		invalidRawJson := bytes.NewBufferString(`{}[`)
		in, out := port.New(ioutil.NopCloser(invalidRawJson))
		app.Create(in, out)
		r := httptest.NewRecorder()
		out.Fill(r)
		if r.Code != http.StatusBadRequest {
			t.Error("wrong code captured")
		}
	})

	validRawJson := bytes.NewBufferString(`{"DeliveryStreamName":"foobar","DeliveryStreamType":"DirectPut","S3DestinationConfiguration":{"BucketARN":"arn:aws:s3:::bucket_name","RoleARN":"arn:aws:iam::accoun_id:role/role_name"}}`)

	t.Run("success request", func(t *testing.T) {
		ctx := context.Background()
		s := arn.NewDeliveryStream(region, accountID, "foobar")
		if dsRepo.Find(ctx, s) != nil {
			t.Error("not yet registered")
		}
		if lst := dstRepo.FindBySource(ctx, s); len(lst) != 0 {
			t.Error("not yet registered")
		}

		in, out := port.New(ioutil.NopCloser(validRawJson))
		app.Create(in, out)
		r := httptest.NewRecorder()
		out.Fill(r)
		if r.Code != http.StatusOK {
			t.Error("wrong code captured")
			return
		}
		if res := r.Body.String(); res != `{"DeliveryStreamARN":"arn:aws:firehose:toyhose:nyanya:deliverystream/foobar"}` {
			t.Errorf("wrong body captured: %s", res)
		}

		ds := dsRepo.Find(ctx, s)
		if ds == nil {
			t.Error("already registered")
		}
		dsts := dstRepo.FindBySource(ctx, s)
		if len(dsts) != 1 {
			t.Error("already registered")
		}
	})

	t.Run("duplicate request", func(t *testing.T) {
		in, out := port.New(ioutil.NopCloser(validRawJson))
		app.Create(in, out)
		r := httptest.NewRecorder()
		out.Fill(r)
		if r.Code != http.StatusBadRequest {
			t.Error("wrong code captured")
		}
	})
}
