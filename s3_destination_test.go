package toyhose

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	fhtypes "github.com/aws/aws-sdk-go-v2/service/firehose/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
)

func TestS3DestinationForExceededDataSize(t *testing.T) {
	awsConf := awsConfig(t)
	s3cli := s3Client(awsConf, s3EndpointURL)
	bucketName := "store-s3-test-" + uuid.New().String()
	if err := setupS3(t, s3cli, bucketName); err != nil {
		t.Fatal(err)
	}

	ch := make(chan *deliveryRecord, 128)
	dst := &s3Destination{
		awsConf: awsConf,
		injectedConf: S3InjectedConf{
			EndPoint: &s3EndpointURL,
		},
		deliveryName: "foobar",
		prefix:       aws.String("aaa"),
		bucketARN:    "arn:aws:s3:::" + bucketName,
		bufferingHints: &fhtypes.BufferingHints{
			SizeInMBs:         aws.Int32(1),
			IntervalInSeconds: aws.Int32(60),
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := dst.Setup(ctx)
	if err != nil {
		t.Fatal(err)
	}
	go dst.Run(ctx, conf, ch)

	b, err := os.ReadFile(filepath.Join("testdata", "dummy.json"))
	if err != nil {
		t.Fatal(err)
	}
	b = append(b, []byte("\n")...)
	sent := 0
	oneMB := 1 * 1024 * 1024
	for sent < oneMB {
		ch <- newDeliveryRecord(b)
		sent += len(b)
	}
	var obj s3types.Object
	for i := 0; i < 20; i++ {
		out, err := s3cli.ListObjects(ctx, &s3.ListObjectsInput{
			Bucket: &bucketName,
			Prefix: dst.prefix,
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(out.Contents) == 1 {
			obj = out.Contents[0]
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if obj.Key == nil {
		t.Fatal("s3 object not found")
	}
	out, err := s3cli.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    obj.Key,
	})
	if err != nil {
		t.Fatal(err)
	}
	content, err := io.ReadAll(out.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(content) != sent {
		t.Errorf("unexpected stored data size: %v", len(content))
	}
}

func TestS3DestinationForExceededInterval(t *testing.T) {
	awsConf := awsConfig(t)
	s3cli := s3Client(awsConf, s3EndpointURL)
	bucketName := "store-s3-test-" + uuid.New().String()
	if err := setupS3(t, s3cli, bucketName); err != nil {
		t.Fatal(err)
	}

	ch := make(chan *deliveryRecord, 128)
	dst := &s3Destination{
		awsConf: awsConf,
		injectedConf: S3InjectedConf{
			EndPoint: &s3EndpointURL,
		},
		deliveryName: "foobar",
		bucketARN:    "arn:aws:s3:::" + bucketName,
		bufferingHints: &fhtypes.BufferingHints{
			SizeInMBs:         aws.Int32(50),
			IntervalInSeconds: aws.Int32(1), // for test use only
		},
		prefix: aws.String("bbb"),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := dst.Setup(ctx)
	if err != nil {
		t.Fatal(err)
	}
	go dst.Run(ctx, conf, ch)

	b, err := os.ReadFile(filepath.Join("testdata", "dummy.json"))
	if err != nil {
		t.Fatal(err)
	}
	b = append(b, []byte("\n")...)
	for i := 0; i < 3; i++ {
		ch <- newDeliveryRecord(b)
		time.Sleep(800 * time.Millisecond)
	}
	var objects []s3types.Object
	for i := 0; i < 50; i++ {
		out, err := s3cli.ListObjects(ctx, &s3.ListObjectsInput{
			Bucket: &bucketName,
			Prefix: dst.prefix,
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(out.Contents) == 2 {
			objects = out.Contents
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if len(objects) == 0 {
		t.Fatal("s3 objects not found")
	}
	for _, c := range objects {
		out, err := s3cli.GetObject(ctx, &s3.GetObjectInput{
			Bucket: &bucketName,
			Key:    c.Key,
		})
		if err != nil {
			t.Fatal(err)
		}
		content, err := io.ReadAll(out.Body)
		if err != nil {
			t.Fatal(err)
		}
		switch len(content) {
		case len(b) * 2, len(b):
		default:
			t.Errorf("unexpected content captured: %s", string(content))
		}
	}
}

var re = regexp.MustCompile("\\d$")

func TestS3DestinationWithDisableBuffering(t *testing.T) {
	awsConf := awsConfig(t)
	s3cli := s3Client(awsConf, s3EndpointURL)
	bucketName := "store-s3-test-" + uuid.New().String()
	if err := setupS3(t, s3cli, bucketName); err != nil {
		t.Fatal(err)
	}

	ch := make(chan *deliveryRecord, 128)
	dst := &s3Destination{
		awsConf: awsConf,
		injectedConf: S3InjectedConf{
			EndPoint:         &s3EndpointURL,
			DisableBuffering: true,
		},
		deliveryName: "foobar",
		prefix:       aws.String("aaa"),
		bucketARN:    "arn:aws:s3:::" + bucketName,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	conf, err := dst.Setup(ctx)
	if err != nil {
		t.Fatal(err)
	}
	go dst.Run(ctx, conf, ch)

	b, err := os.ReadFile(filepath.Join("testdata", "dummy.json"))
	if err != nil {
		t.Fatal(err)
	}
	b = append(b, byte('\n'))
	for i := 0; i < 5; i++ {
		ch <- newDeliveryRecord(append(b, []byte(fmt.Sprint(i))...))
	}
	var objects []s3types.Object
	for i := 0; i < 20; i++ {
		out, err := s3cli.ListObjects(ctx, &s3.ListObjectsInput{
			Bucket: &bucketName,
			Prefix: dst.prefix,
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(out.Contents) == 5 {
			objects = out.Contents
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if len(objects) == 0 {
		t.Fatal("s3 objects not found")
	}
	for _, obj := range objects {
		out, err := s3cli.GetObject(ctx, &s3.GetObjectInput{
			Bucket: &bucketName,
			Key:    obj.Key,
		})
		if err != nil {
			t.Fatal(err)
		}
		content, err := io.ReadAll(out.Body)
		if err != nil {
			t.Fatal(err)
		}
		if !re.Match(content) {
			t.Errorf("unexpected stored data captured: %v", string(content))
		}
	}
}
