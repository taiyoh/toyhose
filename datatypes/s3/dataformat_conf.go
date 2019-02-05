package s3

import (
	"github.com/taiyoh/toyhose/datatypes/common"
)

// https://docs.aws.amazon.com/ja_jp/firehose/latest/APIReference/API_DataFormatConversionConfiguration.html

type DataFormateConf struct {
	Enabled    bool               `json:"Enabled"` // default: true
	SchemaConf *common.SchemaConf `json:"SchemaConfiguration"`
}
