package toyhose

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

func TestS3DestinationForExceededDataSize(t *testing.T) {
	s3cli := s3Client()
	bucketName := "store-s3-test-" + uuid.New().String()
	closer, err := setupS3(s3cli, bucketName)
	if err != nil {
		t.Fatal(err)
	}
	defer closer()

	ch := make(chan []byte, 10)
	dst := &s3Destination{
		deliveryName: "foobar",
		source:       ch,
		conf: &firehose.S3DestinationConfiguration{
			BucketARN: aws.String("arn:aws:s3:::" + bucketName),
			BufferingHints: &firehose.BufferingHints{
				SizeInMBs:         aws.Int64(1),
				IntervalInSeconds: aws.Int64(60),
			},
			Prefix: aws.String("aaa"),
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go dst.Run(ctx)

	b, err := ioutil.ReadFile(filepath.Join("testdata", "dummy.json"))
	if err != nil {
		t.Fatal(err)
	}
	b = append(b, []byte("\n")...)
	sent := 0
	oneMB := 1 * 1024 * 1024
	for sent < oneMB {
		ch <- b
		sent += len(b)
	}
	var obj *s3.Object
	for i := 0; i < 50; i++ {
		out, err := s3cli.ListObjects(&s3.ListObjectsInput{
			Bucket: &bucketName,
			Prefix: dst.conf.Prefix,
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(out.Contents) == 1 {
			obj = out.Contents[0]
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if obj == nil {
		t.Fatal("s3 object not found")
	}
	for i := 0; i < 4; i++ {
		out, err := s3cli.ListObjects(&s3.ListObjectsInput{
			Bucket: &bucketName,
			Prefix: obj.Key,
		})
		if err != nil {
			t.Fatal(err)
		}
		obj = out.Contents[0]
	}
	out, err := s3cli.GetObject(&s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    obj.Key,
	})
	if err != nil {
		t.Fatal(err)
	}
	content, err := ioutil.ReadAll(out.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(content) != sent {
		t.Errorf("unexpected stored data size: %v", len(content))
	}
}
