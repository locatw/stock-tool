package storage

import (
	"context"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/suite"

	"stock-tool/internal/util/testutil"
)

type S3ClientTestSuite struct {
	testutil.S3Test
	client *S3Client
}

func TestS3Client(t *testing.T) {
	suite.Run(t, new(S3ClientTestSuite))
}

func (s *S3ClientTestSuite) SetupSuite() {
	s.S3Test.SetupSuite()
	s.client = NewS3Client(S3Config{
		Endpoint:       s.Endpoint,
		Bucket:         testutil.TestS3Bucket,
		AccessKey:      testutil.TestS3AccessKey,
		SecretKey:      testutil.TestS3SecretKey,
		Region:         testutil.TestS3Region,
		ForcePathStyle: true,
	})
}

func (s *S3ClientTestSuite) TestPutObject() {
	type TestCase struct {
		name string
		key  string
		data []byte
	}
	tests := []TestCase{
		{
			name: "simple JSON",
			key:  "test/path/data.json",
			data: []byte(`{"hello":"world"}`),
		},
		{
			name: "preserves raw bytes with multibyte characters",
			key:  "landing/jquants/brand/2025/06/01/20250601T120000Z_abc12345.json",
			data: []byte(`{"info":[{"Code":"86970","CompanyName":"日本取引所グループ"}]}`),
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()

			err := s.client.PutObject(ctx, tt.key, tt.data)
			s.Require().NoError(err)

			body := s.getObject(ctx, tt.key)
			s.Equal(tt.data, body)
		})
	}
}

func (s *S3ClientTestSuite) getObject(ctx context.Context, key string) []byte {
	rawClient := s3.New(s3.Options{
		BaseEndpoint: aws.String(s.Endpoint),
		Region:       testutil.TestS3Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			testutil.TestS3AccessKey, testutil.TestS3SecretKey, "",
		),
		UsePathStyle: true,
	})

	result, err := rawClient.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(testutil.TestS3Bucket),
		Key:    aws.String(key),
	})
	s.Require().NoError(err)
	defer result.Body.Close()

	body, err := io.ReadAll(result.Body)
	s.Require().NoError(err)
	return body
}
