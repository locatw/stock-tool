package storage

import (
	"bytes"
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Config struct {
	Endpoint       string
	Bucket         string
	AccessKey      string
	SecretKey      string
	Region         string
	ForcePathStyle bool
}

type S3Client struct {
	client *s3.Client
	bucket string
}

func NewS3Client(cfg S3Config) *S3Client {
	client := s3.New(s3.Options{
		BaseEndpoint: &cfg.Endpoint,
		Region:       cfg.Region,
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		UsePathStyle: cfg.ForcePathStyle,
	})

	return &S3Client{
		client: client,
		bucket: cfg.Bucket,
	}
}

func NewS3ClientFromClient(client *s3.Client, bucket string) *S3Client {
	return &S3Client{
		client: client,
		bucket: bucket,
	}
}

func (c *S3Client) PutObject(ctx context.Context, key string, data []byte) error {
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	return err
}

func (c *S3Client) CreateBucket(ctx context.Context) error {
	_, err := c.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(c.bucket),
	})
	return err
}
