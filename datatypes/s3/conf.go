package s3

import (
	"github.com/taiyoh/toyhose/datatypes/arn"
	"github.com/taiyoh/toyhose/errors"
)

// https://docs.aws.amazon.com/ja_jp/firehose/latest/APIReference/API_S3DestinationConfiguration.html

// Conf describes the configuration of a destination in Amazon S3
type Conf struct {
	BucketARN         string         `json:"BucketARN"`
	RoleARN           string         `json:"RoleARN"`
	BufferingHints    BufferingHints `json:"BufferingHints"`
	CompressionFormat *string        `json:"CompressionFormat"` // default: UNCOMPRESSED
	Prefix            *string        `json:"Prefix"`
}

func (c *Conf) validateARN() errors.Raised {
	bArn := c.BucketARN
	if bArn == "" {
		return errors.NewInvalidArgumentException("BucketARN is required")
	}
	if l := len(bArn); 2048 < l {
		return errors.NewInvalidArgumentException("BucketARN value length is over")
	}
	if _, err := arn.RestoreS3FromRaw(bArn); err != nil {
		return errors.NewInvalidArgumentException("BucketARN value is invalid format")
	}
	rArn := c.RoleARN
	if rArn == "" {
		return errors.NewInvalidArgumentException("RoleARN is required")
	}
	if l := len(rArn); 512 < l {
		return errors.NewInvalidArgumentException("RoleARN value length is over")
	}
	if _, err := arn.RestoreIAMRoleFromRaw(rArn); err != nil {
		return errors.NewInvalidArgumentException("RoleARN value is invalid format")
	}
	return nil
}

var compressionFormatMap = map[string]struct{}{
	"UNCOMPRESSED": struct{}{},
	"GZIP":         struct{}{},
	"ZIP":          struct{}{},
	"Snappy":       struct{}{},
}

func (c *Conf) validateCompressionFormat() errors.Raised {
	cf := c.CompressionFormat
	if cf == nil {
		return nil
	}
	if _, ok := compressionFormatMap[*cf]; !ok {
		return errors.NewInvalidArgumentException("CompressionFormat")
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
}

// Validate provides validating each field
func (c *Conf) Validate() errors.Raised {
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
