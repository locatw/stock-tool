package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"stock-tool/internal/domain/ingestion"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type DataType struct {
	ID                  uuid.UUID `gorm:"type:uuid"`
	DataSourceID        uuid.UUID `gorm:"type:uuid"`
	Name                string
	Enabled             bool
	Schedule            datatypes.JSONType[scheduleJSON]
	BackfillEnabled     bool
	StaleTimeoutMinutes int
	Settings            datatypes.JSONType[map[string]any]
	CreatedAt           time.Time `gorm:"autoCreateTime:false"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime:false"`
}

func (m *DataType) toEntity() *ingestion.DataType {
	return ingestion.NewDataTypeDirectly(
		m.ID,
		m.DataSourceID,
		m.Name,
		m.Enabled,
		m.Schedule.Data().toEntity(),
		m.BackfillEnabled,
		m.StaleTimeoutMinutes,
		m.Settings.Data(),
		m.CreatedAt,
		m.UpdatedAt,
	)
}

type scheduleJSON struct {
	Type  string   `json:"type"`
	Times []string `json:"times,omitempty"`
}

func (s scheduleJSON) toEntity() ingestion.Schedule {
	times := lo.Map(s.Times, func(t string, _ int) ingestion.TimeOfDay { return ingestion.TimeOfDay(t) })
	schedule, err := ingestion.NewDailySchedule(times)
	if err != nil {
		panic("repository: corrupt schedule in database: " + err.Error())
	}
	return schedule
}

func toScheduleJSON(s ingestion.Schedule) scheduleJSON {
	return scheduleJSON{
		Type:  string(ingestion.ScheduleTypeDaily),
		Times: lo.Map(s.Times(), func(t ingestion.TimeOfDay, _ int) string { return string(t) }),
	}
}

func toDataTypeDBModel(e *ingestion.DataType) *DataType {
	return &DataType{
		ID:                  e.ID(),
		DataSourceID:        e.DataSourceID(),
		Name:                e.Name(),
		Enabled:             e.Enabled(),
		Schedule:            datatypes.NewJSONType(toScheduleJSON(e.Schedule())),
		BackfillEnabled:     e.BackfillEnabled(),
		StaleTimeoutMinutes: e.StaleTimeoutMinutes(),
		Settings:            datatypes.NewJSONType(e.Settings()),
		CreatedAt:           e.CreatedAt(),
		UpdatedAt:           e.UpdatedAt(),
	}
}

type DataTypeRepository struct {
	db *gorm.DB
}

func NewDataTypeRepository(db *gorm.DB) *DataTypeRepository {
	return &DataTypeRepository{db: db}
}

func (r *DataTypeRepository) Create(ctx context.Context, dt *ingestion.DataType) (*ingestion.DataType, error) {
	dbModel := toDataTypeDBModel(dt)
	if err := r.db.WithContext(ctx).Create(dbModel).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("name %q: %w", dt.Name(), ingestion.ErrDataTypeNameConflict)
		}
		return nil, err
	}
	return dbModel.toEntity(), nil
}

func (r *DataTypeRepository) FindByID(ctx context.Context, id uuid.UUID) (*ingestion.DataType, error) {
	var dbDataType DataType
	err := r.db.WithContext(ctx).First(&dbDataType, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return dbDataType.toEntity(), nil
}

func (r *DataTypeRepository) ListBySourceID(
	ctx context.Context,
	dataSourceID uuid.UUID,
) ([]*ingestion.DataType, error) {
	var dbTypes []DataType
	err := r.db.WithContext(ctx).Where("data_source_id = ?", dataSourceID).Find(&dbTypes).Error
	if err != nil {
		return nil, err
	}
	return lo.Map(dbTypes, func(dt DataType, _ int) *ingestion.DataType { return dt.toEntity() }), nil
}

func (r *DataTypeRepository) Update(ctx context.Context, dt *ingestion.DataType) error {
	err := r.db.WithContext(ctx).
		Model(&DataType{}).
		Where("id = ?", dt.ID()).
		Updates(map[string]any{
			"name":                  dt.Name(),
			"enabled":               dt.Enabled(),
			"schedule":              datatypes.NewJSONType(toScheduleJSON(dt.Schedule())),
			"backfill_enabled":      dt.BackfillEnabled(),
			"stale_timeout_minutes": dt.StaleTimeoutMinutes(),
			"settings":              datatypes.NewJSONType(dt.Settings()),
			"updated_at":            dt.UpdatedAt(),
		}).Error
	if err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("name %q: %w", dt.Name(), ingestion.ErrDataTypeNameConflict)
		}
		return err
	}
	return nil
}

func (r *DataTypeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&DataType{}).Error
}
