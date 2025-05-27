package repository

import (
	"context"
	"fmt"
	"time"

	"stock-tool/database"
	"stock-tool/internal/domain/extract"

	"github.com/samber/lo"
	"gorm.io/gorm"
)

type ExtractTask struct {
	ID         int
	Source     string
	DataType   string
	Status     string
	ErrorInfo  string
	StartedAt  *time.Time
	FinishedAt *time.Time
	CreatedAt  time.Time `gorm:"autoCreateTime:false"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime:false"`
	S3Files    []*ExtractedDataS3
}

func (t *ExtractTask) ToEntity() *extract.ExtractTask {
	s3Files := lo.Map(t.S3Files, func(f *ExtractedDataS3, _ int) *extract.ExtractedDataS3 {
		return f.ToEntity()
	})

	return extract.NewExtractTaskDirectly(
		t.ID,
		t.Source,
		t.DataType,
		t.Status,
		t.ErrorInfo,
		t.StartedAt,
		t.FinishedAt,
		t.CreatedAt,
		t.UpdatedAt,
		s3Files,
	)
}

func toExtractTask(e *extract.ExtractTask) *ExtractTask {
	return &ExtractTask{
		ID:         e.ID(),
		Source:     e.Source(),
		DataType:   e.DataType(),
		Status:     e.Status(),
		ErrorInfo:  e.ErrorInfo(),
		StartedAt:  e.StartedAt(),
		FinishedAt: e.FinishedAt(),
		CreatedAt:  e.CreatedAt(),
		UpdatedAt:  e.UpdatedAt(),
	}
}

type ExtractedDataS3 struct {
	ID             int
	ExtractTaskID  int
	TargetDateTime time.Time
	Bucket         string
	Key            string
	CreatedAt      time.Time `gorm:"autoCreateTime:false"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime:false"`
}

func (s *ExtractedDataS3) TableName() string {
	return fmt.Sprintf("%s.extracted_data_s3s", database.SchemaName)
}

func (s *ExtractedDataS3) ToEntity() *extract.ExtractedDataS3 {
	return extract.NewExtractedDataS3Directly(
		s.ID,
		s.ExtractTaskID,
		s.TargetDateTime,
		s.Bucket,
		s.Key,
		s.CreatedAt,
		s.UpdatedAt,
	)
}

func toExtractedDataS3(e *extract.ExtractedDataS3) *ExtractedDataS3 {
	return &ExtractedDataS3{
		ID:             e.ID(),
		ExtractTaskID:  e.ExtractTaskID(),
		TargetDateTime: e.TargetDateTime(),
		Bucket:         e.Bucket(),
		Key:            e.Key(),
		CreatedAt:      e.CreatedAt(),
		UpdatedAt:      e.UpdatedAt(),
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
	if err := r.db.WithContext(ctx).Create(dbTask).Error; err != nil {
		return err
	}

	for _, s3File := range task.S3Files() {
		dbS3File := toExtractedDataS3(s3File)
		dbS3File.ExtractTaskID = dbTask.ID
		if err := r.db.WithContext(ctx).Create(dbS3File).Error; err != nil {
			return err
		}
	}

	return nil
}
