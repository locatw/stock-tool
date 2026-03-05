package usecase

import (
	"context"
	"fmt"
	"time"

	"stock-tool/internal/domain/ingestion"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

type DataSourceRepository interface {
	Create(ctx context.Context, src *ingestion.DataSource) (*ingestion.DataSource, error)
	FindByID(ctx context.Context, id uuid.UUID) (*ingestion.DataSource, error)
	List(ctx context.Context) ([]*ingestion.DataSource, error)
	Update(ctx context.Context, src *ingestion.DataSource) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type CreateDataSourceRequest struct {
	Name     string
	Enabled  bool
	Timezone string
	Settings map[string]any
}

type UpdateDataSourceRequest struct {
	ID       uuid.UUID
	Name     string
	Enabled  bool
	Timezone string
	Settings map[string]any
}

type DataSourceResponse struct {
	ID        uuid.UUID
	Name      string
	Enabled   bool
	Timezone  string
	Settings  map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
}

func newDataSourceResponse(e *ingestion.DataSource) *DataSourceResponse {
	return &DataSourceResponse{
		ID:        e.ID(),
		Name:      e.Name(),
		Enabled:   e.Enabled(),
		Timezone:  e.TimezoneString(),
		Settings:  e.Settings(),
		CreatedAt: e.CreatedAt(),
		UpdatedAt: e.UpdatedAt(),
	}
}

type DataSourceUseCase struct {
	repo DataSourceRepository
}

func NewDataSourceUseCase(repo DataSourceRepository) *DataSourceUseCase {
	return &DataSourceUseCase{repo: repo}
}

func (uc *DataSourceUseCase) Create(ctx context.Context, req *CreateDataSourceRequest) (*DataSourceResponse, error) {
	ds, err := ingestion.NewDataSource(req.Name, req.Enabled, req.Timezone, req.Settings)
	if err != nil {
		return nil, &ValidationError{Message: err.Error()}
	}
	created, err := uc.repo.Create(ctx, ds)
	if err != nil {
		return nil, fmt.Errorf("failed to create data source: %w", err)
	}
	return newDataSourceResponse(created), nil
}

func (uc *DataSourceUseCase) Get(ctx context.Context, id uuid.UUID) (*DataSourceResponse, error) {
	found, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find data source: %w", err)
	}
	if found == nil {
		return nil, nil
	}
	return newDataSourceResponse(found), nil
}

func (uc *DataSourceUseCase) List(ctx context.Context) ([]*DataSourceResponse, error) {
	sources, err := uc.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list data sources: %w", err)
	}
	return lo.Map(sources, func(s *ingestion.DataSource, _ int) *DataSourceResponse {
		return newDataSourceResponse(s)
	}), nil
}

func (uc *DataSourceUseCase) Update(ctx context.Context, req *UpdateDataSourceRequest) (*DataSourceResponse, error) {
	existing, err := uc.repo.FindByID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find data source: %w", err)
	}
	if existing == nil {
		return nil, nil
	}

	if err := existing.Update(req.Name, req.Enabled, req.Timezone, req.Settings); err != nil {
		return nil, &ValidationError{Message: err.Error()}
	}
	if err := uc.repo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update data source: %w", err)
	}

	result, err := uc.repo.FindByID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find updated data source: %w", err)
	}
	return newDataSourceResponse(result), nil
}

func (uc *DataSourceUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return uc.repo.Delete(ctx, id)
}
