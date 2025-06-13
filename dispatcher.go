package toyhose

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/firehose/types"
	"github.com/google/uuid"
)

// DispatcherConfig represents configuration data struct for Dispatcher.
type DispatcherConfig struct {
	S3InjectedConf      S3InjectedConf
	KinesisInjectedConf KinesisInjectedConf
	AWSConf             aws.Config
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
func NewDispatcher(conf *DispatcherConfig) *Dispatcher {
	return &Dispatcher{
		conf:                conf.AWSConf,
		accountID:           "", // FIXME: Get AccountID from STS or other means if needed.
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
	conf                aws.Config
	accountID           string
	region              string
	s3InjectedConf      S3InjectedConf
	kinesisInjectedConf KinesisInjectedConf
	pool                *deliveryStreamPool
}

// Dispatch handlers HTTP request as http.HandlerFunc interface.
func (d *Dispatcher) Dispatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := uuid.New().String()
	w.Header().Add("x-amzn-RequestId", reqID)
	w.Header().Add("Content-Type", r.Header.Get("Content-Type"))
	op, err := parseTarget(r.Header.Get("X-Amz-Target"))
	if err != nil {
		outputForJSON(w, nil, err)
		return
	}
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		outputForJSON(w, nil, err)
		return
	}
	if err := verifyV4(ctx, d.conf, r, bodyBytes); err != nil {
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
		out, err := svc.Create(ctx, bodyBytes)
		outputForJSON(w, out, err)
	case "DeleteDeliveryStream":
		out, err := svc.Delete(ctx, bodyBytes)
		outputForJSON(w, out, err)
	case "PutRecord":
		out, err := svc.Put(ctx, bodyBytes)
		outputForJSON(w, out, err)
	case "PutRecordBatch":
		out, err := svc.PutBatch(ctx, bodyBytes)
		outputForJSON(w, out, err)
	case "ListDeliveryStreams":
		out, err := svc.Listing(ctx, bodyBytes)
		outputForJSON(w, out, err)
	case "DescribeDeliveryStream":
		// Use custom struct for JSON marshaling
		out, err := svc.Describe(ctx, bodyBytes)
		outputForJSON(w, out, err)
	default:
		// Consider creating a new error type for v2 or using a generic one
		outputForJSON(w, nil, errors.New("InvalidAction: invalid action received"))
	}
}

type outputSerializable interface {
	String() string
	GoString() string
}

func outputForJSON(w http.ResponseWriter, out interface{}, err error) {
	if err == nil {
		if err := json.NewEncoder(w).Encode(out); err != nil {
			log.Error().Err(err).Msg("failed to decode output json")
		}
		return
	}

	// Simplified error handling for now.
	// TODO: Implement proper error type checking for v2 SDK errors.
	var resourceNotFoundException *types.ResourceNotFoundException
	if errors.As(err, &resourceNotFoundException) {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}

	// Create a serializable error message
	errMsg := struct {
		Message string `json:"message"`
	}{
		Message: err.Error(),
	}
	if err2 := json.NewEncoder(w).Encode(errMsg); err2 != nil {
		log.Error().Err(err2).Msg("failed to decode output error json")
	}
}

var errInvalidTargetHeader = errors.New("invalid X-Amz-Target")

func parseTarget(tgt string) (string, error) {
	// ex) Firehose_20150804.CreateDeliveryStream
	parts := strings.Split(tgt, ".")
	if len(parts) != 2 || !strings.HasPrefix(parts[0], "Firehose_") {
		return "", errors.New("MissingAction: no action received")
	}
	return parts[1], nil
}
