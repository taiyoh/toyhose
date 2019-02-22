package toyhose

import (
	"strings"
)

type ActionType int

const (
	NotFound ActionType = iota
	CreateDeliveryStream
	DeleteDeliveryStream
	DescribeDeliveryStream
	ListDeliveryStreams
	_ // ListTagsForDeliveryStream
	PutRecord
	PutRecordBatch
)

var commandTable = map[string]ActionType{
	"CreateDeliveryStream":   CreateDeliveryStream,
	"DeleteDeliveryStream":   DeleteDeliveryStream,
	"DescribeDeliveryStream": DescribeDeliveryStream,
	"ListDeliveryStream":     ListDeliveryStreams,
	"PutRecord":              PutRecord,
	"PutRecordBatch":         PutRecordBatch,
}

const (
	currentVersion = "Firehose_20150804"
)

func FindType(target string) ActionType {
	command := strings.Replace(target, currentVersion+".", "", 1)
	if cmd, found := commandTable[command]; found {
		return cmd
	}
	return NotFound
}
