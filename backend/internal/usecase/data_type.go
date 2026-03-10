package usecase

import (
	"context"
	"fmt"
	"time"

	"stock-tool/internal/domain/ingestion"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

var validUpdateFrequencies = map[string]bool{
	"daily":   true,
	"weekly":  true,
	"monthly": true,
	"hourly":  true,
}

// DataTypeRepository provides persistence for DataType entities.
type DataTypeRepository interface {
	// Create persists a new DataType. Returns the created entity with
	// server-assigned fields.
	Create(ctx context.Context, dt *ingestion.DataType) (*ingestion.DataType, error)

	// FindByID returns the DataType with the given ID, or (nil, nil) if not found.
	FindByID(ctx context.Context, id uuid.UUID) (*ingestion.DataType, error)

	// ListBySourceID returns all DataType entities belonging to the given DataSource.
	// Returns an empty slice when none are found.
	ListBySourceID(ctx context.Context, dataSourceID uuid.UUID) ([]*ingestion.DataType, error)

	// Update persists changes to an existing DataType.
	Update(ctx context.Context, dt *ingestion.DataType) error

	// Delete removes the DataType with the given ID.
	Delete(ctx context.Context, id uuid.UUID) error
}

type CreateDataTypeRequest struct {
	DataSourceID        uuid.UUID
	Name                string
	Enabled             bool
	UpdateFrequency     string
	UpdateTimes         []string
	BackfillEnabled     bool
	StaleTimeoutMinutes int
	Settings            map[string]any
}

type UpdateDataTypeRequest struct {
	ID                  uuid.UUID
	Name                string
	Enabled             bool
	UpdateFrequency     string
	UpdateTimes         []string
	BackfillEnabled     bool
	StaleTimeoutMinutes int
	Settings            map[string]any
}

type DataTypeResponse struct {
	ID                  uuid.UUID
	DataSourceID        uuid.UUID
	Name                string
	Enabled             bool
	UpdateFrequency     string
	UpdateTimes         []string
	BackfillEnabled     bool
	StaleTimeoutMinutes int
	Settings            map[string]any
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func newDataTypeResponse(e *ingestion.DataType) *DataTypeResponse {
	return &DataTypeResponse{
		ID:                  e.ID(),
		DataSourceID:        e.DataSourceID(),
		Name:                e.Name(),
		Enabled:             e.Enabled(),
		UpdateFrequency:     e.UpdateFrequency(),
		UpdateTimes:         e.UpdateTimes(),
		BackfillEnabled:     e.BackfillEnabled(),
		StaleTimeoutMinutes: e.StaleTimeoutMinutes(),
		Settings:            e.Settings(),
		CreatedAt:           e.CreatedAt(),
		UpdatedAt:           e.UpdatedAt(),
	}
}

type DataTypeUseCase struct {
	repo DataTypeRepository
}

func NewDataTypeUseCase(repo DataTypeRepository) *DataTypeUseCase {
	return &DataTypeUseCase{repo: repo}
}

// Create creates a new data type. Returns a ValidationError on invalid input.
func (uc *DataTypeUseCase) Create(ctx context.Context, req *CreateDataTypeRequest) (*DataTypeResponse, error) {
	if err := validateUpdateFrequency(req.UpdateFrequency); err != nil {
		return nil, err
	}

	entity := ingestion.NewDataType(
		req.DataSourceID,
		req.Name,
		req.Enabled,
		req.UpdateFrequency,
		req.UpdateTimes,
		req.BackfillEnabled,
		req.StaleTimeoutMinutes,
		req.Settings,
	)
	created, err := uc.repo.Create(ctx, entity)
	if err != nil {
		return nil, fmt.Errorf("failed to create data type: %w", err)
	}
	return newDataTypeResponse(created), nil
}

// Get retrieves a data type by ID. Returns (nil, nil) when not found.
func (uc *DataTypeUseCase) Get(ctx context.Context, id uuid.UUID) (*DataTypeResponse, error) {
	found, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find data type: %w", err)
	}
	if found == nil {
		return nil, nil
	}
	return newDataTypeResponse(found), nil
}

// List returns all data types belonging to the given data source.
func (uc *DataTypeUseCase) List(ctx context.Context, dataSourceID uuid.UUID) ([]*DataTypeResponse, error) {
	types, err := uc.repo.ListBySourceID(ctx, dataSourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list data types: %w", err)
	}
	return lo.Map(types, func(t *ingestion.DataType, _ int) *DataTypeResponse {
		return newDataTypeResponse(t)
	}), nil
}

// Update applies changes to an existing data type. Returns (nil, nil) when
// not found, or a ValidationError on invalid input.
func (uc *DataTypeUseCase) Update(ctx context.Context, req *UpdateDataTypeRequest) (*DataTypeResponse, error) {
	if err := validateUpdateFrequency(req.UpdateFrequency); err != nil {
		return nil, err
	}

	existing, err := uc.repo.FindByID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find data type: %w", err)
	}
	if existing == nil {
		return nil, nil
	}

	existing.Update(
		req.Name,
		req.Enabled,
		req.UpdateFrequency,
		req.UpdateTimes,
		req.BackfillEnabled,
		req.StaleTimeoutMinutes,
		req.Settings,
	)
	if err := uc.repo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update data type: %w", err)
	}

	result, err := uc.repo.FindByID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find updated data type: %w", err)
	}
	return newDataTypeResponse(result), nil
}

// Delete removes a data type by ID.
func (uc *DataTypeUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return uc.repo.Delete(ctx, id)
}

func validateUpdateFrequency(freq string) error {
	if !validUpdateFrequencies[freq] {
		return &ValidationError{
			Message: fmt.Sprintf(
				"invalid update frequency: %s (allowed: daily, weekly, monthly, hourly)", freq,
			),
		}
	}
	return nil
}
