package usecase

import "time"

type FetchDataRequest struct {
	Source    string
	DataType  string
	Code      *string
	DestURL   string
	StartDate *time.Time
	EndDate   *time.Time
}

type FetchDataResponse struct{}
