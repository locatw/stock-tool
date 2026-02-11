package usecase

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"stock-tool/database"
	"stock-tool/internal/domain/extract"
	"stock-tool/internal/infra/repository"
	"stock-tool/internal/infra/storage"
	"stock-tool/internal/util/testutil"
)

// --- Mock (external service only) ---

type BrandDataFetcherMock struct {
	mock.Mock
}

func (m *BrandDataFetcherMock) FetchBrands(
	ctx context.Context,
	code *string,
	date *time.Time,
) ([]byte, error) {
	args := m.Called(ctx, code, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// --- Suite ---

type ExtractTaskUseCaseTestSuite struct {
	testutil.DBTest
	s3Test   testutil.S3Test
	db       *gorm.DB
	repo     *repository.ExtractTaskRepository
	s3Client *storage.S3Client
}

func TestExtractTaskUseCase(t *testing.T) {
	suite.Run(t, new(ExtractTaskUseCaseTestSuite))
}

func (s *ExtractTaskUseCaseTestSuite) SetupSuite() {
	s.DBTest.SetupSuite()
	s.s3Test.SetT(s.T())
	s.s3Test.SetupSuite()

	s.s3Client = storage.NewS3Client(storage.S3Config{
		Endpoint:       s.s3Test.Endpoint,
		Bucket:         testutil.TestS3Bucket,
		AccessKey:      testutil.TestS3AccessKey,
		SecretKey:      testutil.TestS3SecretKey,
		Region:         testutil.TestS3Region,
		ForcePathStyle: true,
	})
}

func (s *ExtractTaskUseCaseTestSuite) TearDownSuite() {
	s.s3Test.TearDownSuite()
	s.DBTest.TearDownSuite()
}

func (s *ExtractTaskUseCaseTestSuite) SetupTest() {
	s.ApplyMigrations()

	db, err := database.CreateGormDB(s.GetDB())
	s.Require().NoError(err)

	s.db = db
	s.repo = repository.NewExtractTaskRepository(db)
}

func (s *ExtractTaskUseCaseTestSuite) TearDownTest() {
	s.Require().NoError(s.CleanupMigrations())
}

func (s *ExtractTaskUseCaseTestSuite) newUseCase(
	fetcher BrandDataFetcher,
) *ExtractTaskUseCase {
	return NewExtractTaskUseCase(fetcher, s.s3Client, s.repo)
}

func (s *ExtractTaskUseCaseTestSuite) TestExtract_Success() {
	ctx := context.Background()
	rawBody := []byte(`{"info":[{"Code":"86970","CompanyName":"日本取引所グループ"}]}`)

	fetcher := new(BrandDataFetcherMock)
	fetcher.On("FetchBrands", ctx, (*string)(nil), (*time.Time)(nil)).
		Return(rawBody, nil)

	uc := s.newUseCase(fetcher)
	resp, err := uc.Extract(ctx, &ExtractTaskRequest{
		Source:   "jquants",
		DataType: "brand",
		Timing:   "daily",
	})

	s.Require().NoError(err)
	s.Equal(extract.ExecutionStatusSucceeded, resp.Status)
	s.NotEmpty(resp.S3Key)

	// Verify DB: task exists
	task, err := s.repo.FindBySourceAndDataType(ctx, "jquants", "brand", "daily")
	s.Require().NoError(err)
	s.NotNil(task)

	// Verify DB: execution with succeeded status
	var dbExecs []repository.ExtractTaskExecution
	s.Require().NoError(
		s.db.Where("extract_task_id = ?", task.ID()).Find(&dbExecs).Error,
	)
	s.Require().Len(dbExecs, 1)
	s.Equal("succeeded", dbExecs[0].Status)
	s.NotNil(dbExecs[0].FinishedAt)

	// Verify DB: S3 file record
	var dbS3Files []repository.ExtractedDataS3
	s.Require().NoError(
		s.db.Where("extract_task_execution_id = ?", dbExecs[0].ID).Find(&dbS3Files).Error,
	)
	s.Require().Len(dbS3Files, 1)
	s.Equal(resp.S3Key, dbS3Files[0].Key)

	// Verify S3: object content matches raw body
	body := s.getS3Object(ctx, resp.S3Key)
	s.Equal(rawBody, body)

	fetcher.AssertExpectations(s.T())
}

func (s *ExtractTaskUseCaseTestSuite) TestExtract_ReusesExistingTask() {
	ctx := context.Background()
	rawBody := []byte(`{"info":[]}`)

	fetcher := new(BrandDataFetcherMock)
	fetcher.On("FetchBrands", ctx, (*string)(nil), (*time.Time)(nil)).
		Return(rawBody, nil)

	uc := s.newUseCase(fetcher)
	req := &ExtractTaskRequest{Source: "jquants", DataType: "brand", Timing: "daily"}

	_, err := uc.Extract(ctx, req)
	s.Require().NoError(err)
	_, err = uc.Extract(ctx, req)
	s.Require().NoError(err)

	// Verify: same task reused (only 1 task record)
	var taskCount int64
	s.db.Model(&repository.ExtractTask{}).Count(&taskCount)
	s.Equal(int64(1), taskCount)

	// Verify: 2 executions created
	var execCount int64
	s.db.Model(&repository.ExtractTaskExecution{}).Count(&execCount)
	s.Equal(int64(2), execCount)
}

func (s *ExtractTaskUseCaseTestSuite) TestExtract_APIError_MarksExecutionFailed() {
	ctx := context.Background()

	fetcher := new(BrandDataFetcherMock)
	fetcher.On("FetchBrands", ctx, (*string)(nil), (*time.Time)(nil)).
		Return(nil, errors.New("API connection timeout"))

	uc := s.newUseCase(fetcher)
	_, err := uc.Extract(ctx, &ExtractTaskRequest{
		Source:   "jquants",
		DataType: "brand",
		Timing:   "daily",
	})

	s.Error(err)
	s.Contains(err.Error(), "API connection timeout")

	// Verify DB: execution marked as failed with error info
	var dbExec repository.ExtractTaskExecution
	s.Require().NoError(s.db.First(&dbExec).Error)
	s.Equal("failed", dbExec.Status)
	s.NotNil(dbExec.ErrorInfo)
	s.NotNil(dbExec.FinishedAt)
}

func (s *ExtractTaskUseCaseTestSuite) TestExtract_UnsupportedSource_MarksExecutionFailed() {
	ctx := context.Background()

	fetcher := new(BrandDataFetcherMock)

	uc := s.newUseCase(fetcher)
	_, err := uc.Extract(ctx, &ExtractTaskRequest{
		Source:   "unknown",
		DataType: "brand",
		Timing:   "daily",
	})

	s.Error(err)
	s.Contains(err.Error(), "unsupported source")

	// Verify DB: execution marked as failed
	var dbExec repository.ExtractTaskExecution
	s.Require().NoError(s.db.First(&dbExec).Error)
	s.Equal("failed", dbExec.Status)
}

func (s *ExtractTaskUseCaseTestSuite) getS3Object(ctx context.Context, key string) []byte {
	rawClient := s3.New(s3.Options{
		BaseEndpoint: aws.String(s.s3Test.Endpoint),
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
