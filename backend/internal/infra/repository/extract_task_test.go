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
			switch sf.Name() {
			case "CreatedAt", "UpdatedAt", "StartedAt", "FinishedAt":
				return true
			}
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
	targetDateTime := time.Now().UTC().Add(-24 * time.Hour).Truncate(time.Microsecond)

	s3File := extract.NewExtractedDataS3("path/to/key.csv")
	exec := extract.NewRunningExecution(targetDateTime)
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
					Status:         "running",
					ErrorInfo:      nil,
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

func (s *ExtractTaskRepositoryTestSuite) TestFindBySourceAndDataType_Found() {
	ctx := context.Background()

	task := extract.NewExtractTask("jquants", "brand", "daily")
	s.Require().NoError(s.repo.Create(ctx, task))

	// distractor: same source/type but different timing
	distractor := extract.NewExtractTask("jquants", "brand", "weekly")
	s.Require().NoError(s.repo.Create(ctx, distractor))

	found, err := s.repo.FindBySourceAndDataType(ctx, "jquants", "brand", "daily")

	s.NoError(err)
	s.NotNil(found)
	s.Equal("jquants", found.Source())
	s.Equal("brand", found.DataType())
	s.Equal("daily", found.Timing())
}

func (s *ExtractTaskRepositoryTestSuite) TestFindBySourceAndDataType_NotFound() {
	ctx := context.Background()

	found, err := s.repo.FindBySourceAndDataType(ctx, "jquants", "brand", "daily")

	s.NoError(err)
	s.Nil(found)
}

func (s *ExtractTaskRepositoryTestSuite) TestCreateExecution() {
	ctx := context.Background()

	task := extract.NewExtractTask("jquants", "brand", "daily")
	s.Require().NoError(s.repo.Create(ctx, task))

	found, err := s.repo.FindBySourceAndDataType(ctx, "jquants", "brand", "daily")
	s.Require().NoError(err)

	targetDateTime := time.Now().UTC().Truncate(time.Microsecond)
	exec := extract.NewRunningExecution(targetDateTime)

	created, err := s.repo.CreateExecution(ctx, found.ID(), exec)

	s.NoError(err)
	s.NotNil(created)
	s.Greater(created.ID(), 0)
	s.Equal(extract.ExecutionStatusRunning, created.Status())
	s.Equal(targetDateTime, created.TargetDateTime())
}

func (s *ExtractTaskRepositoryTestSuite) TestUpdateExecution() {
	ctx := context.Background()

	task := extract.NewExtractTask("jquants", "brand", "daily")
	s.Require().NoError(s.repo.Create(ctx, task))

	found, err := s.repo.FindBySourceAndDataType(ctx, "jquants", "brand", "daily")
	s.Require().NoError(err)

	targetDateTime := time.Now().UTC().Truncate(time.Microsecond)
	exec := extract.NewRunningExecution(targetDateTime)
	created, err := s.repo.CreateExecution(ctx, found.ID(), exec)
	s.Require().NoError(err)

	// distractor: a second execution that should not be affected
	exec2, err := s.repo.CreateExecution(ctx, found.ID(), extract.NewRunningExecution(targetDateTime.Add(time.Hour)))
	s.Require().NoError(err)

	// Reconstruct the entity to simulate the domain transition
	reconstructed := extract.NewExtractTaskExecutionDirectly(
		created.ID(),
		created.TargetDateTime(),
		created.Status(),
		created.ErrorInfo(),
		created.StartedAt(),
		created.FinishedAt(),
		created.CreatedAt(),
		created.UpdatedAt(),
		[]*extract.ExtractedDataS3{},
	)
	reconstructed.Succeed()

	err = s.repo.UpdateExecution(ctx, reconstructed)

	s.NoError(err)

	// Verify status was updated
	var dbExec ExtractTaskExecution
	s.Require().NoError(s.db.First(&dbExec, created.ID()).Error)
	s.Equal("succeeded", dbExec.Status)
	s.NotNil(dbExec.FinishedAt)

	// Verify distractor exec2 was not affected
	var dbExec2 ExtractTaskExecution
	s.Require().NoError(s.db.First(&dbExec2, exec2.ID()).Error)
	s.Equal("running", dbExec2.Status)
}

func (s *ExtractTaskRepositoryTestSuite) TestCreateExtractedDataS3() {
	ctx := context.Background()

	task := extract.NewExtractTask("jquants", "brand", "daily")
	s.Require().NoError(s.repo.Create(ctx, task))

	found, err := s.repo.FindBySourceAndDataType(ctx, "jquants", "brand", "daily")
	s.Require().NoError(err)

	targetDateTime := time.Now().UTC().Truncate(time.Microsecond)
	exec := extract.NewRunningExecution(targetDateTime)
	created, err := s.repo.CreateExecution(ctx, found.ID(), exec)
	s.Require().NoError(err)

	s3File := extract.NewExtractedDataS3("landing/jquants/brand/2025/06/01/data.json")
	s3Created, err := s.repo.CreateExtractedDataS3(ctx, created.ID(), s3File)

	s.NoError(err)
	s.NotNil(s3Created)
	s.Greater(s3Created.ID(), 0)
	s.Equal("landing/jquants/brand/2025/06/01/data.json", s3Created.Key())
}
