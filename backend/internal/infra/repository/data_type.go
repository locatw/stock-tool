package repository

import (
	"context"
	"errors"
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
	UpdateFrequency     string
	UpdateTimes         datatypes.JSONType[[]string]
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
		m.UpdateFrequency,
		m.UpdateTimes.Data(),
		m.BackfillEnabled,
		m.StaleTimeoutMinutes,
		m.Settings.Data(),
		m.CreatedAt,
		m.UpdatedAt,
	)
}

func toDataTypeDBModel(e *ingestion.DataType) *DataType {
	return &DataType{
		ID:                  e.ID(),
		DataSourceID:        e.DataSourceID(),
		Name:                e.Name(),
		Enabled:             e.Enabled(),
		UpdateFrequency:     e.UpdateFrequency(),
		UpdateTimes:         datatypes.NewJSONType(e.UpdateTimes()),
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
	return r.db.WithContext(ctx).
		Model(&DataType{}).
		Where("id = ?", dt.ID()).
		Updates(map[string]any{
			"name":                  dt.Name(),
			"enabled":               dt.Enabled(),
			"update_frequency":      dt.UpdateFrequency(),
			"update_times":          datatypes.NewJSONType(dt.UpdateTimes()),
			"backfill_enabled":      dt.BackfillEnabled(),
			"stale_timeout_minutes": dt.StaleTimeoutMinutes(),
			"settings":              datatypes.NewJSONType(dt.Settings()),
			"updated_at":            dt.UpdatedAt(),
		}).Error
}

func (r *DataTypeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&DataType{}).Error
}
