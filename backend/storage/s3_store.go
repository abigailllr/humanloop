package storage

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type S3Store struct {
	client *s3.Client
	bucket string
}

func NewS3Store() (*S3Store, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}
	bucket := os.Getenv("AWS_BUCKET")
	return &S3Store{client: s3.NewFromConfig(cfg), bucket: bucket}, nil
}

func (s *S3Store) Save(challengeID, userID string, r io.Reader) (string, error) {
	key := "videos/" + challengeID + "/" + userID + "/" + uuid.New().String() + ".mp4"
	_, err := s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
		Body:   r,
	})
	if err != nil {
		return "", err
	}
	return "s3://" + s.bucket + "/" + key, nil
}

func (s *S3Store) BaseDir() string {
	return "s3://" + s.bucket
}

func (s *S3Store) PresignGetHMDF(hmdfPath string, ttl time.Duration) (string, error) {
	key := strings.TrimPrefix(hmdfPath, "s3://"+s.bucket+"/")
	presignClient := s3.NewPresignClient(s.client)
	req, err := presignClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	}, func(o *s3.PresignOptions) {
		o.Expires = ttl
	})
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

func (s *S3Store) UploadHMDF(submissionID string, r io.Reader) (string, error) {
	key := "hmdf/" + submissionID + ".hmdf.json.gz"
	_, err := s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
		Body:   r,
	})
	if err != nil {
		return "", err
	}
	return "s3://" + s.bucket + "/" + key, nil
}
