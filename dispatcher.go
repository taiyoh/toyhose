package toyhose

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/google/uuid"
)

func newDispatcher(accountID, region string) *dispatcher {
	return &dispatcher{
		accountID: accountID,
		region:    region,
		pool: &deliveryStreamPool{
			pool: map[string]*deliveryStream{},
		},
	}
}

type dispatcher struct {
	accountID string
	region    string
	pool      *deliveryStreamPool
}

func (d *dispatcher) Dispatch(w http.ResponseWriter, r *http.Request) {
	reqID := uuid.New().String()
	w.Header().Add("x-amzn-RequestId", reqID)
	w.Header().Add("Content-Type", r.Header.Get("Content-Type"))
	op, err := parseTarget(r.Header.Get("X-Amz-Target"))
	if err != nil {
		outputForJSON(w, nil, err)
		return
	}
	b := bytes.NewBuffer([]byte{})
	io.Copy(b, r.Body)
	if err := verifyV4(r, bytes.NewReader(b.Bytes())); err != nil {
		outputForJSON(w, nil, err)
		return
	}
	svc := &DeliveryStreamService{
		region:    d.region,
		accountID: d.accountID,
		pool:      d.pool,
	}
	switch op {
	case "CreateDeliveryStream":
		out, err := svc.Create(r.Context(), b.Bytes())
		outputForJSON(w, out, err)
	case "DeleteDeliveryStream":
		out, err := svc.Delete(r.Context(), b.Bytes())
		outputForJSON(w, out, err)
	case "PutRecord":
		out, err := svc.Put(r.Context(), b.Bytes())
		outputForJSON(w, out, err)
	case "PutRecordBatch":
		out, err := svc.PutBatch(r.Context(), b.Bytes())
		outputForJSON(w, out, err)
	default:
		err := awserr.New("InvalidAction", "invalid action received", errInvalidTargetHeader)
		outputForJSON(w, nil, err)
	}
}

type outputSerializable interface {
	String() string
	GoString() string
}

func outputForJSON(w http.ResponseWriter, out outputSerializable, err error) {
	if err == nil {
		json.NewEncoder(w).Encode(out)
		return
	}
	switch e := err.(type) {
	case awserr.Error:
		switch e.Code() {
		case "InternalFailure":
			w.WriteHeader(http.StatusInternalServerError)
		case firehose.ErrCodeServiceUnavailableException:
			w.WriteHeader(http.StatusServiceUnavailable)
		case firehose.ErrCodeResourceNotFoundException:
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(err)
}

var errInvalidTargetHeader = errors.New("invalid X-Amz-Target")

func parseTarget(tgt string) (string, error) {
	// ex) Firehose_20150804.CreateDeliveryStream
	parts := strings.Split(tgt, ".")
	if len(parts) != 2 || !strings.HasPrefix(parts[0], "Firehose_") {
		return "", awserr.New("MissingAction", "no action received", errInvalidTargetHeader)
	}
	return parts[1], nil
}
