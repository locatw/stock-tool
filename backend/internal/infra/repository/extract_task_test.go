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
	cmp.FilterPath(func(p cmp.Path) bool {
		if len(p) == 0 {
			return false
		}
		last := p[len(p)-1]
		if sf, ok := last.(cmp.StructField); ok {
			return sf.Name() == "CreatedAt" || sf.Name() == "UpdatedAt"
		}
		return false
	}, cmp.Ignore()),
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

	s3File := extract.NewExtractedDataS3("path/to/key.csv")
	exec := extract.NewExtractTaskExecution(targetDateTime, "created")
	exec.AddS3File(s3File)
	task := extract.NewExtractTask("j-quants", "daily-quotes", "daily")
	task.AddExecution(exec)

	err := s.repo.Create(ctx, task)

	s.NoError(err)

	var dbTasks []*ExtractTask
	err = s.db.Preload("ExtractTaskExecutions.ExtractedDataS3s").Find(&dbTasks).Error
	s.NoError(err)

	expectedTasks := []*ExtractTask{
		{
			ID:       1,
			Source:   "j-quants",
			DataType: "daily-quotes",
			Timing:   "daily",
			ExtractTaskExecutions: []*ExtractTaskExecution{
				{
					ID:             1,
					ExtractTaskID:  1,
					TargetDateTime: targetDateTime,
					Status:         "created",
					ErrorInfo:      nil,
					StartedAt:      nil,
					FinishedAt:     nil,
					ExtractedDataS3s: []*ExtractedDataS3{
						{
							ID:                     1,
							ExtractTaskExecutionID: 1,
							Key:                    "path/to/key.csv",
						},
					},
				},
			},
		},
	}

	if diff := cmp.Diff(expectedTasks, dbTasks, cmpOpts); diff != "" {
		s.T().Errorf("task mismatch (-want +got):\n%s", diff)
	}
}
