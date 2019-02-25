package s3

import (
	"github.com/taiyoh/toyhose/datatypes/arn"
	"github.com/taiyoh/toyhose/exception"
)

// https://docs.aws.amazon.com/ja_jp/firehose/latest/APIReference/API_ExtendedS3DestinationConfiguration.html

// Conf describes the configuration of a destination in Amazon S3
type Conf struct {
	BucketARN         string         `json:"BucketARN"`
	RoleARN           string         `json:"RoleARN"`
	BufferingHints    BufferingHints `json:"BufferingHints"`
	CompressionFormat *string        `json:"CompressionFormat"` // default: UNCOMPRESSED
	DataFormatConf    DataFormatConf `json:"DataFormatConversionConfiguration"`
	ErrorOutputPrefix *string        `json:"ErrorOutputPrefix"`
	Prefix            *string        `json:"Prefix"`
}

func (c *Conf) validateARN() exception.Raised {
	bArn := c.BucketARN
	if l := len(bArn); l < 1 || 2048 < l {
		return exception.NewInvalidArgument("BucketARN")
	}
	if _, err := arn.RestoreS3FromRaw(bArn); err != nil {
		return exception.NewInvalidArgument("BucketARN")
	}
	rArn := c.RoleARN
	if l := len(rArn); l < 1 || 512 < l {
		return exception.NewInvalidArgument("RoleARN")
	}
	if _, err := arn.RestoreIAMRoleFromRaw(rArn); err != nil {
		return exception.NewInvalidArgument("RoleARN")
	}
	return nil
}

var compressionFormatMap = map[string]struct{}{
	"UNCOMPRESSED": struct{}{},
	"GZIP":         struct{}{},
	"ZIP":          struct{}{},
	"Snappy":       struct{}{},
}

func (c *Conf) validateCompressionFormat() exception.Raised {
	cf := c.CompressionFormat
	if cf == nil {
		return nil
	}
	if _, ok := compressionFormatMap[*cf]; !ok {
		return exception.NewInvalidArgument("CompressionFormat")
	}
	return nil
}

var defaultCompressionFormat = "UNCOMPRESSED"

// FillDefaultValue provides filling value if field is not assigned
func (c *Conf) FillDefaultValue() {
	if f := c.CompressionFormat; f == nil {
		c.CompressionFormat = &defaultCompressionFormat
	}

	(&c.BufferingHints).FillDefaultValue()
	(&c.DataFormatConf).FillDefaultValue()
}

// Validate provides validating each field
func (c *Conf) Validate() exception.Raised {
	if err := c.validateARN(); err != nil {
		return err
	}
	if err := c.BufferingHints.Validate(); err != nil {
		return err
	}
	if err := c.validateCompressionFormat(); err != nil {
		return err
	}
	return nil
}
