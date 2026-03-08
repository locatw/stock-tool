package testutil

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/suite"
)

const (
	TestS3Bucket    = "test-bucket"
	TestS3AccessKey = "testAccessKey"
	TestS3SecretKey = "testSecretKey"
	TestS3Region    = "ap-northeast-1"
)

type S3Test struct {
	suite.Suite
	pool     *dockertest.Pool
	resource *dockertest.Resource
	Endpoint string
}

func (s *S3Test) setupDockerTest() error {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return fmt.Errorf("could not construct pool: %w", err)
	}

	s.pool = pool

	if err := s.pool.Client.Ping(); err != nil {
		return fmt.Errorf("could not connect to Docker: %w", err)
	}

	resource, err := s.pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: "chrislusf/seaweedfs",
			Tag:        "latest",
			Cmd:        []string{"server", "-s3", "-dir=/data"},
		},
		func(config *docker.HostConfig) {
			config.AutoRemove = true
			config.RestartPolicy = docker.RestartPolicy{
				Name: "no",
			}
		},
	)
	if err != nil {
		return fmt.Errorf("could not start resource: %w", err)
	}
	s.resource = resource

	s.Endpoint = fmt.Sprintf("http://localhost:%s", s.resource.GetPort("8333/tcp"))

	pool.MaxWait = 30 * time.Second
	err = s.pool.Retry(func() error {
		client := s3.New(s3.Options{
			BaseEndpoint: &s.Endpoint,
			Region:       TestS3Region,
			Credentials: credentials.NewStaticCredentialsProvider(
				TestS3AccessKey,
				TestS3SecretKey,
				"",
			),
			UsePathStyle: true,
		})

		_, err := client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
		return err
	})
	if err != nil {
		return fmt.Errorf("could not connect to SeaweedFS: %w", err)
	}

	client := s3.New(s3.Options{
		BaseEndpoint: &s.Endpoint,
		Region:       TestS3Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			TestS3AccessKey,
			TestS3SecretKey,
			"",
		),
		UsePathStyle: true,
	})
	if _, err := client.CreateBucket(context.Background(), &s3.CreateBucketInput{
		Bucket: &[]string{TestS3Bucket}[0],
	}); err != nil {
		return fmt.Errorf("could not create test bucket: %w", err)
	}

	return nil
}

func (s *S3Test) SetupSuite() {
	if err := s.setupDockerTest(); err != nil {
		s.T().Fatal(err)
	}
}

func (s *S3Test) TearDownSuite() {
	if s.resource != nil {
		if err := s.pool.Purge(s.resource); err != nil {
			s.T().Errorf("Could not purge resource: %v", err)
		}
	}
}
