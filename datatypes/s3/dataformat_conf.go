package s3

// https://docs.aws.amazon.com/ja_jp/firehose/latest/APIReference/API_DataFormatConversionConfiguration.html

// DataFormatConf specifies that you want Kinesis Data Firehose to convert data from the JSON format to the Parquet or ORC format before writing it to Amazon S3.
type DataFormatConf struct {
	Enabled    *bool       `json:"Enabled"` // default: true
	SchemaConf *SchemaConf `json:"SchemaConfiguration"`
}

var defaultEnabled = true

// FillDefaultValue provides filling default value if not assigned
func (c *DataFormatConf) FillDefaultValue() {
	if ePtr := c.Enabled; ePtr == nil {
		c.Enabled = &defaultEnabled
	}
}
