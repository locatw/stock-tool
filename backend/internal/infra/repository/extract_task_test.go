package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"stock-tool/database"
	"stock-tool/internal/domain/extract"
	"stock-tool/internal/util/testutil"
)

var cmpOpts = cmp.Options{
	cmp.Comparer(func(x, y time.Time) bool {
		// PostgreSQL only supports microsecond precision, so truncate to microseconds before comparison
		return x.Truncate(time.Microsecond).Equal(y.Truncate(time.Microsecond))
	}),
}

type ExtractTaskRepositoryTestSuite struct {
	testutil.DBTest
	repo *ExtractTaskRepository
	db   *gorm.DB
}

func TestExtractTaskRepository(t *testing.T) {
	suite.Run(t, new(ExtractTaskRepositoryTestSuite))
}

func (s *ExtractTaskRepositoryTestSuite) SetupTest() {
	s.ApplyMigrations()

	db, err := database.CreateGormDB(s.GetDB())
	s.Require().NoError(err)

	s.db = db
	s.repo = NewExtractTaskRepository(db)
}

func (s *ExtractTaskRepositoryTestSuite) TearDownTest() {
	s.Require().NoError(s.CleanupMigrations())
}

func (s *ExtractTaskRepositoryTestSuite) TestCreate() {
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	targetDateTime := now.Add(-24 * time.Hour).Truncate(time.Microsecond)

	task := extract.NewExtractTask("j-quants", "daily-quotes", "pending")
	s3File := extract.NewExtractedDataS3(targetDateTime, "my-bucket", "path/to/key.csv")
	task.AddS3File(s3File)

	err := s.repo.Create(ctx, task)

	s.NoError(err)

	var dbTasks []*ExtractTask
	err = s.db.Preload("S3Files").Find(&dbTasks).Error
	s.NoError(err)

	expectedTasks := []*ExtractTask{
		{
			ID:        1,
			Source:    "j-quants",
			DataType:  "daily-quotes",
			Status:    "pending",
			CreatedAt: now,
			UpdatedAt: now,
			S3Files: []*ExtractedDataS3{
				{
					ID:             1,
					ExtractTaskID:  1,
					TargetDateTime: targetDateTime,
					Bucket:         "my-bucket",
					Key:            "path/to/key.csv",
					CreatedAt:      now,
					UpdatedAt:      now,
				},
			},
		},
	}

	if diff := cmp.Diff(expectedTasks, dbTasks, cmpOpts); diff != "" {
		s.T().Errorf("task mismatch (-want +got):\n%s", diff)
	}
}
