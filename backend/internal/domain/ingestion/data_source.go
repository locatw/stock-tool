package ingestion

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type DataSource struct {
	id        uuid.UUID
	name      string
	enabled   bool
	timezone  *time.Location
	settings  map[string]any
	createdAt time.Time
	updatedAt time.Time
}

func NewDataSource(
	name string,
	enabled bool,
	timezone string,
	settings map[string]any,
) (*DataSource, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone: %s", timezone)
	}
	now := time.Now()
	return &DataSource{
		id:        uuid.Must(uuid.NewV7()),
		name:      name,
		enabled:   enabled,
		timezone:  loc,
		settings:  settings,
		createdAt: now,
		updatedAt: now,
	}, nil
}

func NewDataSourceDirectly(
	id uuid.UUID,
	name string,
	enabled bool,
	timezone *time.Location,
	settings map[string]any,
	createdAt time.Time,
	updatedAt time.Time,
) *DataSource {
	return &DataSource{
		id:        id,
		name:      name,
		enabled:   enabled,
		timezone:  timezone,
		settings:  settings,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (s *DataSource) Update(
	name string,
	enabled bool,
	timezone string,
	settings map[string]any,
) error {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return fmt.Errorf("invalid timezone: %s", timezone)
	}
	s.name = name
	s.enabled = enabled
	s.timezone = loc
	s.settings = settings
	s.updatedAt = time.Now()
	return nil
}

func (s *DataSource) ID() uuid.UUID            { return s.id }
func (s *DataSource) Name() string             { return s.name }
func (s *DataSource) Enabled() bool            { return s.enabled }
func (s *DataSource) Timezone() *time.Location { return s.timezone }
func (s *DataSource) TimezoneString() string   { return s.timezone.String() }
func (s *DataSource) Settings() map[string]any { return s.settings }
func (s *DataSource) CreatedAt() time.Time     { return s.createdAt }
func (s *DataSource) UpdatedAt() time.Time     { return s.updatedAt }
