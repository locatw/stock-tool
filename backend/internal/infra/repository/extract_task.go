package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"stock-tool/database"
	"stock-tool/internal/domain/extract"

	"github.com/samber/lo"
	"gorm.io/gorm"
)

type ExtractTask struct {
	ID                    int
	Source                string
	DataType              string
	Timing                string
	CreatedAt             time.Time               `gorm:"autoCreateTime:false"`
	UpdatedAt             time.Time               `gorm:"autoUpdateTime:false"`
	ExtractTaskExecutions []*ExtractTaskExecution `gorm:"foreignKey:ExtractTaskID"`
}

func (t *ExtractTask) ToEntity() *extract.ExtractTask {
	execs := lo.Map(t.ExtractTaskExecutions, func(f *ExtractTaskExecution, _ int) *extract.ExtractTaskExecution {
		return f.ToEntity()
	})

	return extract.NewExtractTaskDirectly(
		t.ID,
		t.Source,
		t.DataType,
		t.Timing,
		t.CreatedAt,
		t.UpdatedAt,
		execs,
	)
}

func toExtractTask(e *extract.ExtractTask) *ExtractTask {
	return &ExtractTask{
		ID:        e.ID(),
		Source:    e.Source(),
		DataType:  e.DataType(),
		Timing:    e.Timing(),
		CreatedAt: e.CreatedAt(),
		UpdatedAt: e.UpdatedAt(),
		ExtractTaskExecutions: lo.Map(
			e.Executions(),
			func(exec *extract.ExtractTaskExecution, _ int) *ExtractTaskExecution {
				return toExtractTaskExecution(exec)
			},
		),
	}
}

type ExtractTaskExecution struct {
	ID               int
	ExtractTaskID    int
	TargetDateTime   time.Time
	Status           string
	ErrorInfo        *string
	StartedAt        *time.Time
	FinishedAt       *time.Time
	CreatedAt        time.Time          `gorm:"autoCreateTime:false"`
	UpdatedAt        time.Time          `gorm:"autoUpdateTime:false"`
	ExtractedDataS3s []*ExtractedDataS3 `gorm:"foreignKey:ExtractTaskExecutionID"`
}

func (t *ExtractTaskExecution) ToEntity() *extract.ExtractTaskExecution {
	s3s := lo.Map(t.ExtractedDataS3s, func(data *ExtractedDataS3, _ int) *extract.ExtractedDataS3 {
		return data.ToEntity()
	})

	return extract.NewExtractTaskExecutionDirectly(
		t.ID,
		t.TargetDateTime,
		extract.ExecutionStatus(t.Status),
		t.ErrorInfo,
		t.StartedAt,
		t.FinishedAt,
		t.CreatedAt,
		t.UpdatedAt,
		s3s,
	)
}

func toExtractTaskExecution(e *extract.ExtractTaskExecution) *ExtractTaskExecution {
	return &ExtractTaskExecution{
		ID:             e.ID(),
		TargetDateTime: e.TargetDateTime(),
		Status:         string(e.Status()),
		ErrorInfo:      e.ErrorInfo(),
		StartedAt:      e.StartedAt(),
		FinishedAt:     e.FinishedAt(),
		CreatedAt:      e.CreatedAt(),
		UpdatedAt:      e.UpdatedAt(),
		ExtractedDataS3s: lo.Map(
			e.S3Files(),
			func(s3 *extract.ExtractedDataS3, _ int) *ExtractedDataS3 {
				return toExtractedDataS3(s3)
			},
		),
	}
}

type ExtractedDataS3 struct {
	ID                     int
	ExtractTaskExecutionID int
	Key                    string
	CreatedAt              time.Time `gorm:"autoCreateTime:false"`
	UpdatedAt              time.Time `gorm:"autoUpdateTime:false"`
}

func (s *ExtractedDataS3) TableName() string {
	return fmt.Sprintf("%s.extracted_data_s3s", database.SchemaName)
}

func (s *ExtractedDataS3) ToEntity() *extract.ExtractedDataS3 {
	return extract.NewExtractedDataS3Directly(
		s.ID,
		s.Key,
		s.CreatedAt,
		s.UpdatedAt,
	)
}

func toExtractedDataS3(e *extract.ExtractedDataS3) *ExtractedDataS3 {
	return &ExtractedDataS3{
		ID:        e.ID(),
		Key:       e.Key(),
		CreatedAt: e.CreatedAt(),
		UpdatedAt: e.UpdatedAt(),
	}
}

type ExtractTaskRepository struct {
	db *gorm.DB
}

func NewExtractTaskRepository(db *gorm.DB) *ExtractTaskRepository {
	return &ExtractTaskRepository{db: db}
}

func (r *ExtractTaskRepository) Create(ctx context.Context, task *extract.ExtractTask) error {
	dbTask := toExtractTask(task)
	return r.db.WithContext(ctx).Create(dbTask).Error
}

func (r *ExtractTaskRepository) FindBySourceAndDataType(
	ctx context.Context,
	source string,
	dataType string,
	timing string,
) (*extract.ExtractTask, error) {
	var dbTask ExtractTask
	err := r.db.WithContext(ctx).
		Where("source = ? AND data_type = ? AND timing = ?", source, dataType, timing).
		First(&dbTask).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return dbTask.ToEntity(), nil
}

func (r *ExtractTaskRepository) CreateExecution(
	ctx context.Context,
	taskID int,
	exec *extract.ExtractTaskExecution,
) (*extract.ExtractTaskExecution, error) {
	dbExec := toExtractTaskExecution(exec)
	dbExec.ExtractTaskID = taskID
	if err := r.db.WithContext(ctx).Create(dbExec).Error; err != nil {
		return nil, err
	}
	return dbExec.ToEntity(), nil
}

func (r *ExtractTaskRepository) UpdateExecution(ctx context.Context, exec *extract.ExtractTaskExecution) error {
	dbExec := toExtractTaskExecution(exec)
	return r.db.WithContext(ctx).
		Model(&ExtractTaskExecution{}).
		Where("id = ?", dbExec.ID).
		Updates(map[string]any{
			"status":      dbExec.Status,
			"error_info":  dbExec.ErrorInfo,
			"finished_at": dbExec.FinishedAt,
			"updated_at":  dbExec.UpdatedAt,
		}).Error
}

func (r *ExtractTaskRepository) CreateExtractedDataS3(
	ctx context.Context,
	executionID int,
	s3File *extract.ExtractedDataS3,
) (*extract.ExtractedDataS3, error) {
	dbS3 := toExtractedDataS3(s3File)
	dbS3.ExtractTaskExecutionID = executionID
	if err := r.db.WithContext(ctx).Create(dbS3).Error; err != nil {
		return nil, err
	}
	return dbS3.ToEntity(), nil
}

func (r *ExtractTaskRepository) Transaction(ctx context.Context, f func(tx *ExtractTaskRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return f(&ExtractTaskRepository{db: tx})
	})
}
