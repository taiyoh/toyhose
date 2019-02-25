package arn

import (
	"errors"
	"fmt"
	"strings"
)

// arn:aws:s3:::bucket_name
// arn:aws:s3:::bucket_name/key_name

// S3 provides AWS ARN for S3
type S3 struct {
	bucketName string
	keyName    string
}

// Code returns ARN as string
func (s S3) Code() string {
	if s.keyName == "" {
		return fmt.Sprintf("arn:aws:s3:::%s", s.bucketName)
	}
	return fmt.Sprintf("arn:aws:s3:::%s/%s", s.bucketName, s.keyName)
}

// RestoreS3FromRaw returns S3 ARN. also error returns if it is invalid
func RestoreS3FromRaw(raw string) (S3, error) {
	prefix := "arn:aws:s3:::"
	if !strings.HasPrefix(raw, prefix) {
		return S3{}, errors.New("invalid prefix")
	}
	cut := strings.Replace(raw, prefix, "", 1)
	if len(cut) < 1 {
		return S3{}, errors.New("no bucket_name")
	}
	names := strings.Split(cut, "/")
	if len(names) == 1 {
		return S3{names[0], ""}, nil
	}
	bucketName := names[0]
	names = names[1:]
	return S3{bucketName, strings.Join(names, "/")}, nil
}
