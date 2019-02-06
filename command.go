package toyhose

import (
	"strings"
)

type Command int

const (
	CommandNotFound Command = iota
	CommandCreateDeliveryStream
	CommandDeleteDeliveryStream
	CommandDescribeDeliveryStream
	CommandListDeliveryStreams
	_ // CommandListTagsForDeliveryStream
	CommandPutRecord
	CommandPutRecordBatch
)

var commandTable = map[string]Command{
	"CreateDeliveryStream":   CommandCreateDeliveryStream,
	"DeleteDeliveryStream":   CommandDeleteDeliveryStream,
	"DescribeDeliveryStream": CommandDescribeDeliveryStream,
	"ListDeliveryStream":     CommandListDeliveryStreams,
	"PutRecord":              CommandPutRecord,
	"PutRecordBatch":         CommandPutRecordBatch,
}

const (
	currentVersion = "Firehose_20150804"
)

func FindCommand(target string) Command {
	command := strings.Replace(target, currentVersion+".", "", 1)
	if cmd, found := commandTable[command]; found {
		return cmd
	}
	return CommandNotFound
}
