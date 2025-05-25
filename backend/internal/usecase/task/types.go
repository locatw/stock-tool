package usecase

import "time"

type ExtractTaskRequest struct {
	Source    string
	DataType  string
	Code      *string
	DestURL   string
	StartDate *time.Time
	EndDate   *time.Time
}

type ExtractTaskResponse struct{}
