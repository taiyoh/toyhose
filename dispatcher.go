package toyhose

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/firehose/types"
	"github.com/aws/smithy-go"
	"github.com/google/uuid"
)

// DispatcherConfig represents configuration data struct for Dispatcher.
type DispatcherConfig struct {
	S3InjectedConf      S3InjectedConf
	KinesisInjectedConf KinesisInjectedConf
	AWSConf             *aws.Config
}

// S3InjectedConf represents injection to S3 destination BufferingHints forcely.
type S3InjectedConf struct {
	SizeInMBs         *int
	IntervalInSeconds *int
	EndPoint          *string
	DisableBuffering  bool
}

// KinesisInjectedConf represents configuration of KinesisStream source.
type KinesisInjectedConf struct {
	Endpoint *string
}

// NewDispatcher returns Dispatcher object.
func NewDispatcher(ctx context.Context, conf *DispatcherConfig) *Dispatcher {
	cred, _ := conf.AWSConf.Credentials.Retrieve(ctx)
	return &Dispatcher{
		conf:                conf.AWSConf,
		accountID:           cred.AccessKeyID,
		region:              conf.AWSConf.Region,
		s3InjectedConf:      conf.S3InjectedConf,
		kinesisInjectedConf: conf.KinesisInjectedConf,
		pool: &deliveryStreamPool{
			pool: map[string]*deliveryStream{},
		},
	}
}

// Dispatcher represents firehose API handler.
type Dispatcher struct {
	conf                *aws.Config
	accountID           string
	region              string
	s3InjectedConf      S3InjectedConf
	kinesisInjectedConf KinesisInjectedConf
	pool                *deliveryStreamPool
}

// Dispatch handlers HTTP request as http.HandlerFunc interface.
func (d *Dispatcher) Dispatch(w http.ResponseWriter, r *http.Request) {
	reqID := uuid.New().String()
	w.Header().Add("x-amzn-RequestId", reqID)
	w.Header().Add("Content-Type", r.Header.Get("Content-Type"))
	op, err := parseTarget(r.Header.Get("X-Amz-Target"))
	if err != nil {
		outputForJSON(w, nil, err)
		return
	}
	b := bytes.NewBuffer([]byte{})
	if _, err := io.Copy(b, r.Body); err != nil {
		outputForJSON(w, nil, err)
		return
	}
	if err := verifyV4(d.conf, r, b.Bytes()); err != nil {
		outputForJSON(w, nil, err)
		return
	}
	svc := &DeliveryStreamService{
		awsConf:             d.conf,
		region:              d.region,
		accountID:           d.accountID,
		s3InjectedConf:      d.s3InjectedConf,
		kinesisInjectedConf: d.kinesisInjectedConf,
		pool:                d.pool,
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
	case "ListDeliveryStreams":
		out, err := svc.Listing(r.Context(), b.Bytes())
		outputForJSON(w, out, err)
	case "DescribeDeliveryStream":
		out, err := svc.Describe(r.Context(), b.Bytes())
		outputForJSON(w, out, err)
	default:
		err := &smithy.OperationError{
			ServiceID:     "Firehose",
			OperationName: op,
			Err:           fmt.Errorf("invalid action received"),
		}
		outputForJSON(w, nil, err)
	}
}

type outputSerializable interface {
	String() string
	GoString() string
}

func outputForJSON(w http.ResponseWriter, out any, err error) {
	if err == nil {
		if err := json.NewEncoder(w).Encode(out); err != nil {
			log.Error().Err(err).Msg("failed to decode output json")
		}
		return
	}
	var e1 *types.ServiceUnavailableException
	var e2 *types.ResourceNotFoundException
	if errors.As(err, &e1) {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else if errors.As(err, &e2) {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
	if err2 := json.NewEncoder(w).Encode(err); err2 != nil {
		log.Error().Err(err2).Msg("failed to decode output error json")
	}
}

var errInvalidTargetHeader = errors.New("invalid X-Amz-Target")

func parseTarget(tgt string) (string, error) {
	// ex) Firehose_20150804.CreateDeliveryStream
	parts := strings.Split(tgt, ".")
	if len(parts) != 2 || !strings.HasPrefix(parts[0], "Firehose_") {
		return "", &smithy.GenericAPIError{
			Code:    "MissingAction",
			Message: "no action received",
			Fault:   smithy.FaultClient,
		}
	}
	return parts[1], nil
}
