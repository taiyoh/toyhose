package s3

import (
	"strings"

	"github.com/taiyoh/toyhose/datatypes/common"
	"github.com/taiyoh/toyhose/exception"
)

type Conf struct {
	BucketARN         string                 `json:"BucketARN"`
	RoleARN           string                 `json:"RoleARN"`
	BufferingHints    *common.BufferingHints `json:"BufferingHints"`
	CompressionFormat string                 `json:"CompressionFormat"` // default: UNCOMPRESSED
	DataFormatConf    *DataFormateConf       `json:"DataFormatConversionConfiguration"`
	ErrorOutputPrefix *string                `json:"ErrorOutputPrefix"`
	Prefix            *string                `json:"Prefix"`
}

func (c *Conf) validateARN() exception.Raised {
	bArn := c.BucketARN
	if l := len(bArn); l < 1 || 2048 < l {
		return exception.NewInvalidArgument("BucketARN")
	}
	if !strings.HasPrefix(bArn, "arn:") {
		return exception.NewInvalidArgument("BucketARN")
	}
	rArn := c.RoleARN
	if l := len(rArn); l < 1 || 512 < l {
		return exception.NewInvalidArgument("RoleARN")
	}
	if !strings.HasPrefix(rArn, "arn:") {
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
	if _, ok := compressionFormatMap[cf]; !ok {
		return exception.NewInvalidArgument("CompressionFormat")
	}
	return nil
}

func (c *Conf) FillDefaultValue() {
	if c.CompressionFormat == "" {
		c.CompressionFormat = "UNCOMPRESSED"
	}
	if bh := c.BufferingHints; bh != nil {
		bh.FillDefaultValue()
	}
}

func (c *Conf) Validate() exception.Raised {
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
