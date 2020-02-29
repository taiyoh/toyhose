package toyhose

import (
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

const defaultRegion = "us-east-1"

func init() {
	os.Setenv("AWS_ACCESS_KEY_ID", "XXXXXXXX")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "YYYYYYYY")
	os.Setenv("AWS_DEFAULT_REGION", defaultRegion)
	os.Setenv("S3_ENDPOINT_URL", "http://localhost:9000")
}

func setupS3(s3cli *s3.S3, bucket string) (func(), error) {
	if _, err := s3cli.CreateBucket(&s3.CreateBucketInput{
		Bucket: &bucket,
	}); err != nil {
		return nil, err
	}
	fn := func() {
		s3cli.DeleteBucket(&s3.DeleteBucketInput{Bucket: &bucket})
	}
	return fn, nil
}

func TestStoreToS3ForNoSuppliedData(t *testing.T) {
	s3cli := s3Client()
	bucketName := "store-s3-test-" + uuid.New().String()
	closer, err := setupS3(s3cli, bucketName)
	if err != nil {
		t.Fatal(err)
	}
	defer closer()

	r := storeToS3Resource{
		deliveryName:       "foobar",
		bucketName:         bucketName,
		prefix:             "",
		shouldGZipCompress: false,
	}
	ts := time.Now()

	storeToS3(context.Background(), r, ts, nil)
	out, err := s3cli.ListObjects(&s3.ListObjectsInput{
		Bucket: &bucketName,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Contents) > 0 {
		t.Errorf("unexpected contents included: %#v", out.Contents)
	}
}

func TestStoreToS3ForRawData(t *testing.T) {
	s3cli := s3Client()
	bucketName := "store-s3-test-" + uuid.New().String()
	closer, err := setupS3(s3cli, bucketName)
	if err != nil {
		t.Fatal(err)
	}
	defer closer()

	r := storeToS3Resource{
		deliveryName:       "foobar",
		bucketName:         bucketName,
		prefix:             "",
		shouldGZipCompress: false,
	}
	ts := time.Now()

	content := "!!!!!!!!!!!!!!!!!!!!!!!!"
	storeToS3(context.Background(), r, ts, []byte(content))
	prefix := ts.Format("2006/01/02/15/")
	out, err := s3cli.ListObjects(&s3.ListObjectsInput{
		Bucket: &bucketName,
		Prefix: &prefix,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Contents) != 1 {
		t.Errorf("unexpected contents included: %#v", out.Contents)
	}
	obj := out.Contents[0]
	if m, _ := regexp.MatchString(fmt.Sprintf("%sfoobar-1-%s-.+", prefix, ts.Format("2006-01-02-15-04-05")), *obj.Key); !m {
		t.Errorf("wrong key name supplied: %v", *obj.Key)
	}
	c, err := s3cli.GetObject(&s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    obj.Key,
	})
	if err != nil {
		t.Fatal(err)
	}
	received, err := ioutil.ReadAll(c.Body)
	if err != nil {
		t.Fatal(err)
	}
	if r := string(received); r != content {
		t.Errorf("wrong content received: %s", r)
	}
}

func TestStoreToS3ForCompressedData(t *testing.T) {
	s3cli := s3Client()
	bucketName := "store-s3-test-" + uuid.New().String()
	closer, err := setupS3(s3cli, bucketName)
	if err != nil {
		t.Fatal(err)
	}
	defer closer()

	r := storeToS3Resource{
		deliveryName:       "foobar",
		bucketName:         bucketName,
		prefix:             "",
		shouldGZipCompress: true,
	}
	ts := time.Now()

	content := "!!!!!!!!!!!!!!!!!!!!!!!!"
	r.shouldGZipCompress = true
	storeToS3(context.Background(), r, ts, []byte(content))
	prefix := ts.Format("2006/01/02/15/")
	out, err := s3cli.ListObjects(&s3.ListObjectsInput{
		Bucket: &bucketName,
		Prefix: &prefix,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Contents) != 1 {
		t.Errorf("unexpected contents included: %#v", out.Contents)
	}
	obj := out.Contents[0]
	if m, _ := regexp.MatchString(fmt.Sprintf("^%sfoobar-1-%s-.+$", prefix, ts.Format("2006-01-02-15-04-05")), *obj.Key); !m {
		t.Errorf("wrong key name supplied: %v", *obj.Key)
	}
	c, err := s3cli.GetObject(&s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    obj.Key,
	})
	if err != nil {
		t.Fatal(err)
	}
	reader, err := gzip.NewReader(c.Body)
	if err != nil {
		t.Fatal(err)
	}
	received, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if r := string(received); r != content {
		t.Errorf("wrong content received: %s", r)
	}
}
