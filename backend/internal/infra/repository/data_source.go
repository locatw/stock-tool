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

type DataSource struct {
	ID        uuid.UUID `gorm:"type:uuid"`
	Name      string
	Enabled   bool
	Timezone  string
	Settings  datatypes.JSONType[map[string]any]
	CreatedAt time.Time `gorm:"autoCreateTime:false"`
	UpdatedAt time.Time `gorm:"autoUpdateTime:false"`
}

func (m *DataSource) toEntity() *ingestion.DataSource {
	loc, _ := time.LoadLocation(m.Timezone)
	return ingestion.NewDataSourceDirectly(
		m.ID,
		m.Name,
		m.Enabled,
		loc,
		m.Settings.Data(),
		m.CreatedAt,
		m.UpdatedAt,
	)
}

func toDataSourceDBModel(e *ingestion.DataSource) *DataSource {
	return &DataSource{
		ID:        e.ID(),
		Name:      e.Name(),
		Enabled:   e.Enabled(),
		Timezone:  e.TimezoneString(),
		Settings:  datatypes.NewJSONType(e.Settings()),
		CreatedAt: e.CreatedAt(),
		UpdatedAt: e.UpdatedAt(),
	}
}

type DataSourceRepository struct {
	db *gorm.DB
}

func NewDataSourceRepository(db *gorm.DB) *DataSourceRepository {
	return &DataSourceRepository{db: db}
}

func (r *DataSourceRepository) Create(ctx context.Context, src *ingestion.DataSource) (*ingestion.DataSource, error) {
	dbModel := toDataSourceDBModel(src)
	if err := r.db.WithContext(ctx).Create(dbModel).Error; err != nil {
		return nil, err
	}
	return dbModel.toEntity(), nil
}

func (r *DataSourceRepository) FindByID(ctx context.Context, id uuid.UUID) (*ingestion.DataSource, error) {
	var dbSource DataSource
	err := r.db.WithContext(ctx).First(&dbSource, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return dbSource.toEntity(), nil
}

func (r *DataSourceRepository) List(ctx context.Context) ([]*ingestion.DataSource, error) {
	var dbSources []DataSource
	if err := r.db.WithContext(ctx).Find(&dbSources).Error; err != nil {
		return nil, err
	}
	return lo.Map(dbSources, func(s DataSource, _ int) *ingestion.DataSource { return s.toEntity() }), nil
}

func (r *DataSourceRepository) Update(ctx context.Context, src *ingestion.DataSource) error {
	return r.db.WithContext(ctx).
		Model(&DataSource{}).
		Where("id = ?", src.ID()).
		Updates(map[string]any{
			"name":       src.Name(),
			"enabled":    src.Enabled(),
			"timezone":   src.TimezoneString(),
			"settings":   datatypes.NewJSONType(src.Settings()),
			"updated_at": src.UpdatedAt(),
		}).Error
}

// Delete deletes the data source with the given ID.
// Associated data_types rows are automatically deleted via ON DELETE CASCADE on the foreign key constraint.
func (r *DataSourceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&DataSource{}).Error
}
