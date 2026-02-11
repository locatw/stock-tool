package usecase

import (
	"stock-tool/internal/domain/extract"
	"time"
)

type ExtractTaskRequest struct {
	Source    string
	DataType string
	Timing   string
	Code     *string
	StartDate *time.Time
	EndDate   *time.Time
}

type ExtractTaskResponse struct {
	S3Key  string
	Status extract.ExecutionStatus
}
