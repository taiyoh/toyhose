package toyhose

import (
	"net/http"

	"github.com/taiyoh/toyhose/actions"
	"github.com/taiyoh/toyhose/actions/port"
	"github.com/taiyoh/toyhose/gateway"
)

type Adapter struct {
	mux       *http.ServeMux
	region    string
	accountID string
	dsRepo    *gateway.DeliveryStream
	destRepo  *gateway.Destination
}

func NewServe(region, accountID string, dsRepo *gateway.DeliveryStream, destRepo *gateway.Destination) *Adapter {
	mux := http.NewServeMux()
	a := &Adapter{
		mux:       mux,
		region:    region,
		accountID: accountID,
		dsRepo:    dsRepo,
		destRepo:  destRepo,
	}
	mux.HandleFunc("/", a.handleFn)
	return a
}

func (a *Adapter) handleFn(res http.ResponseWriter, req *http.Request) {
	if !a.validateRequest(res, req) {
		return
	}
	fn := a.Dispatch(req.Header.Get("X-Amz-Target"))
	if fn == nil {
		http.NotFound(res, req)
		return
	}
	input, output := port.New(req.Body)

	fn(input, output)

	output.Fill(res)
}

func (a *Adapter) validateRequest(res http.ResponseWriter, req *http.Request) bool {
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return false
	}
	if req.URL.Path != "/" {
		http.NotFound(res, req)
		return false
	}
	return true
}

func (a *Adapter) ServeMux() *http.ServeMux {
	return a.mux
}

type UseCaseFn func(*port.Input, *port.Output)

func (a *Adapter) Dispatch(target string) UseCaseFn {
	d := actions.NewDeliveryStream(a.dsRepo, a.destRepo, a.region, a.accountID)
	switch FindType(target) {
	case CreateDeliveryStream:
		return d.Create
	case DescribeDeliveryStream:
		return d.Describe
	case ListDeliveryStreams:
		return d.List
	case PutRecord:
		return d.PutRecord
	case PutRecordBatch:
		return d.PutRecordBatch
	}
	return nil
}
