package usecase

import "time"

type ExtractRequest struct {
	Source    string
	DataType  string
	Code      *string
	DestURL   string
	StartDate *time.Time
	EndDate   *time.Time
}

type ExtractResponse struct{}
