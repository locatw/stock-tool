package ingestion

import (
	"context"
	"errors"
	"time"

	"stock-tool/internal/util/clock"
	"stock-tool/internal/util/idp"

	"github.com/google/uuid"
)

// ErrDataTypeNameConflict is returned when a data type name
// already exists for the same data source.
var ErrDataTypeNameConflict = errors.New("data type name already exists")

// DataType represents a category of data belonging to a DataSource.
// It holds ingestion configuration: update schedule, backfill policy,
// and stale timeout.
type DataType struct {
	id                  uuid.UUID
	dataSourceID        uuid.UUID
	name                string
	enabled             bool
	schedule            Schedule
	backfillEnabled     bool
	staleTimeoutMinutes int
	settings            map[string]any
	createdAt           time.Time
	updatedAt           time.Time
}

func NewDataType(
	ctx context.Context,
	dataSourceID uuid.UUID,
	name string,
	enabled bool,
	schedule Schedule,
	backfillEnabled bool,
	staleTimeoutMinutes int,
	settings map[string]any,
) *DataType {
	now := clock.Now(ctx)
	return &DataType{
		id:                  idp.NewV7(ctx),
		dataSourceID:        dataSourceID,
		name:                name,
		enabled:             enabled,
		schedule:            schedule,
		backfillEnabled:     backfillEnabled,
		staleTimeoutMinutes: staleTimeoutMinutes,
		settings:            settings,
		createdAt:           now,
		updatedAt:           now,
	}
}

func NewDataTypeDirectly(
	id uuid.UUID,
	dataSourceID uuid.UUID,
	name string,
	enabled bool,
	schedule Schedule,
	backfillEnabled bool,
	staleTimeoutMinutes int,
	settings map[string]any,
	createdAt time.Time,
	updatedAt time.Time,
) *DataType {
	return &DataType{
		id:                  id,
		dataSourceID:        dataSourceID,
		name:                name,
		enabled:             enabled,
		schedule:            schedule,
		backfillEnabled:     backfillEnabled,
		staleTimeoutMinutes: staleTimeoutMinutes,
		settings:            settings,
		createdAt:           createdAt,
		updatedAt:           updatedAt,
	}
}

func (t *DataType) ID() uuid.UUID            { return t.id }
func (t *DataType) DataSourceID() uuid.UUID  { return t.dataSourceID }
func (t *DataType) Name() string             { return t.name }
func (t *DataType) Enabled() bool            { return t.enabled }
func (t *DataType) Schedule() Schedule       { return t.schedule }
func (t *DataType) BackfillEnabled() bool    { return t.backfillEnabled }
func (t *DataType) StaleTimeoutMinutes() int { return t.staleTimeoutMinutes }
func (t *DataType) Settings() map[string]any { return t.settings }
func (t *DataType) CreatedAt() time.Time     { return t.createdAt }
func (t *DataType) UpdatedAt() time.Time     { return t.updatedAt }

func (t *DataType) Update(
	ctx context.Context,
	name string,
	enabled bool,
	schedule Schedule,
	backfillEnabled bool,
	staleTimeoutMinutes int,
	settings map[string]any,
) {
	t.name = name
	t.enabled = enabled
	t.schedule = schedule
	t.backfillEnabled = backfillEnabled
	t.staleTimeoutMinutes = staleTimeoutMinutes
	t.settings = settings
	t.updatedAt = clock.Now(ctx)
}

func (t *DataType) StaleTimeout() time.Duration {
	return time.Duration(t.staleTimeoutMinutes) * time.Minute
}
