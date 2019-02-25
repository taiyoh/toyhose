package s3_test

import (
	"encoding/json"
	"testing"

	"github.com/taiyoh/toyhose/datatypes/s3"
)

func TestConf(t *testing.T) {
	t.Run("incorrect", func(t *testing.T) {
		for _, tt := range []struct {
			label    string
			rawjson  string
			hasError bool
		}{
			{
				"empty fields", `{}`, true,
			},
			{
				"invalid BucketARN", `{"BucketARN":"arn:aws:s3:::"}`, true,
			},
			{
				"BucketARN only correct1", `{"BucketARN":"arn:aws:s3:::foo/bar","RoleARN":""}`, true,
			},
			{
				"BucketARN only correct2", `{"BucketARN":"arn:aws:s3:::foo/bar","RoleARN":"arn:aws:iam::hoge:role:fuga"}`, true,
			},
			{
				"wrong BufferingHints", `{"BucketARN":"arn:aws:s3:::foo/bar","RoleARN":"arn:aws:iam::hoge:role/fuga","CompressionFormat":"GZIP","BufferingHints":{"IntervalInSeconds":50}}`, true,
			},
			{
				"wrong CompressionFormat", `{"BucketARN":"arn:aws:s3:::foo/bar","RoleARN":"arn:aws:iam::hoge:role/fuga","CompressionFormat":"foo"}`, true,
			},
		} {
			c := s3.Conf{}
			json.Unmarshal([]byte(tt.rawjson), &c)
			if err := c.Validate(); (err != nil) != tt.hasError {
				t.Errorf(`label="%s" msg="error should be returned"`, tt.label)
			}
		}
	})

	t.Run("correct with required only", func(t *testing.T) {
		rawjson := `{"BucketARN":"arn:aws:s3:::foo/bar","RoleARN":"arn:aws:iam::hoge:role/fuga"}`
		c := s3.Conf{}
		json.Unmarshal([]byte(rawjson), &c)
		if err := c.Validate(); err != nil {
			t.Error("error exists:", err)
		}
		if cf := c.CompressionFormat; cf != nil {
			t.Error("CompressionFormat is already assigned:", *cf)
		}
		if secP := c.BufferingHints.IntervalInSeconds; secP != nil {
			t.Error("BufferingHints.IntervalInSeconds is already assigned:", *secP)
		}
		if mbP := c.BufferingHints.SizeInMBs; mbP != nil {
			t.Error("BufferingHints.SizeInMBs is already assigned:", *mbP)
		}
		if flg := c.DataFormatConf.Enabled; flg != nil {
			t.Error("DataFormatConf.Enabled is already assigned:", *flg)
		}

		c.FillDefaultValue()
		cf := c.CompressionFormat
		if cf == nil {
			t.Error("CompressionFormat should be filled")
		} else if *cf != "UNCOMPRESSED" {
			t.Error("wrong CompressionFormat is assigned:", *cf)
		}

		if secP := c.BufferingHints.IntervalInSeconds; secP == nil {
			t.Error("BufferingHints.IntervalInSeconds should be assigned")
		} else if *secP != 300 {
			t.Error("wrong BufferingHints.IntervalInSeconds is assigned:", *secP)
		}
		if mbP := c.BufferingHints.SizeInMBs; mbP == nil {
			t.Error("BufferingHints.SizeInMBs should be assigned")
		} else if *mbP != 5 {
			t.Error("wrong BufferingHints.SizeInMBs is assigned:", *mbP)
		}

		if flg := c.DataFormatConf.Enabled; flg == nil {
			t.Error("DataFormatConf.Enabled should be assigned")
		} else if !*flg {
			t.Error("wrong DataFormatConf.Enabled is assigned:", *flg)
		}
	})

	t.Run("correct with value filled", func(t *testing.T) {
		rawjson := `{"BucketARN":"arn:aws:s3:::foo/bar","RoleARN":"arn:aws:iam::hoge:role/fuga","CompressionFormat":"GZIP","BufferingHints":{"IntervalInSeconds":60,"SizeInMBs":2},"DataFormatConversionConfiguration":{"Enabled":false}}`
		c := s3.Conf{}
		json.Unmarshal([]byte(rawjson), &c)
		if err := c.Validate(); err != nil {
			t.Error("error exists:", err)
		}
		if cf := c.CompressionFormat; cf == nil {
			t.Error("CompressionFormat should be assigned:", *cf)
		}
		if secP := c.BufferingHints.IntervalInSeconds; secP == nil {
			t.Error("BufferingHints.IntervalInSeconds should be assigned:", *secP)
		}
		if mbP := c.BufferingHints.SizeInMBs; mbP == nil {
			t.Error("BufferingHints.SizeInMBs should be assigned:", *mbP)
		}
		if flg := c.DataFormatConf.Enabled; flg == nil {
			t.Error("DataFormatConf.Enabled should be assigned:", *flg)
		}
		c.FillDefaultValue()
		if cf := c.CompressionFormat; *cf != "GZIP" {
			t.Error("wrong CompressionFormat is assigned:", *cf)
		}
		if secP := c.BufferingHints.IntervalInSeconds; *secP != 60 {
			t.Error("wrong BufferingHints.IntervalInSeconds is assigned:", *secP)
		}
		if mbP := c.BufferingHints.SizeInMBs; *mbP != 2 {
			t.Error("wrong BufferingHints.SizeInMBs is assigned:", *mbP)
		}
		if flg := c.DataFormatConf.Enabled; *flg {
			t.Error("wrong DataFormatConf.Enabled is assigned:", *flg)
		}
	})
}
