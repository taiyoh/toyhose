package common

// https://docs.aws.amazon.com/ja_jp/firehose/latest/APIReference/API_CloudWatchLoggingOptions.html

type CloudWatchLoggingOptions struct {
	Enabled       bool   `json:"Enabled"`
	LogGroupName  string `json:"LogGroupName"`
	LogStreamName string `json:"LogStreamName"`
}
