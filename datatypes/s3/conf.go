package s3

import (
	"errors"
	"strings"

	"github.com/taiyoh/toyhose/datatypes/common"
)

type Conf struct {
	BucketARN                string                           `json:"BucketARN"`
	RoleARN                  string                           `json:"RoleARN"`
	BufferingHints           *common.BufferingHints           `json:"BufferingHints"`
	CloudWatchLoggingOptions *common.CloudWatchLoggingOptions `json:"CloudWatchLoggingOptions"`
	CompressionFormat        string                           `json:"CompressionFormat"` // default: UNCOMPRESSED
	DataFormatConf           *DataFormateConf                 `json:"DataFormatConversionConfiguration"`
	ErrorOutputPrefix        *string                          `json:"ErrorOutputPrefix"`
	Prefix                   *string                          `json:"Prefix"`
}

func (c *Conf) validateARN() error {
	bArn := c.BucketARN
	if l := len(bArn); l < 1 || 2048 < l {
		return errors.New("BucketARN length is invalid")
	}
	if !strings.HasPrefix(bArn, "arn:") {
		return errors.New("BucketARN pattern unmatched")
	}
	rArn := c.RoleARN
	if l := len(rArn); l < 1 || 512 < l {
		return errors.New("RoleARN length is invalid")
	}
	if !strings.HasPrefix(rArn, "arn:") {
		return errors.New("RoleARN pattern unmatched")
	}
	return nil
}

func (c *Conf) validateCompressionFormat() error {
	cf := c.CompressionFormat
	for _, f := range []string{"UNCOMPRESSED", "GZIP", "ZIP", "Snappy"} {
		if cf == f {
			return nil
		}
	}
	return errors.New("CompressionFormat is invalid")
}

func (c *Conf) FillDefaultValue() {
	if c.CompressionFormat == "" {
		c.CompressionFormat = "UNCOMPRESSED"
	}
}

func (c *Conf) Validate() error {
	if err := c.validateARN(); err != nil {
		return err
	}
	if bh := c.BufferingHints; bh != nil {
		if err := bh.Validate(); err != nil {
			return err
		}
	}
	if err := c.validateCompressionFormat(); err != nil {
		return err
	}
	return nil
}
