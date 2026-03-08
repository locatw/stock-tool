package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"stock-tool/database"
	"stock-tool/internal/domain/ingestion"
	"stock-tool/internal/util/testutil"
)

var dataTypeCmpOpts = cmp.Options{
	cmp.AllowUnexported(ingestion.DataType{}),
}

type DataTypeRepositoryTestSuite struct {
	testutil.DBTest
	repo *DataTypeRepository
	db   *gorm.DB
}

func TestDataTypeRepository(t *testing.T) {
	suite.Run(t, new(DataTypeRepositoryTestSuite))
}

func (s *DataTypeRepositoryTestSuite) SetupTest() {
	s.ApplyMigrations()

	db, err := database.CreateGormDB(s.GetDB())
	s.Require().NoError(err)

	s.db = db
	s.repo = NewDataTypeRepository(db)
}

func (s *DataTypeRepositoryTestSuite) TearDownTest() {
	s.Require().NoError(s.CleanupMigrations())
}

func (s *DataTypeRepositoryTestSuite) seedDataSource() uuid.UUID {
	now := time.Now()
	source := &DataSource{
		ID:        uuid.Must(uuid.NewV7()),
		Name:      "j-quants",
		Enabled:   true,
		Timezone:  "Asia/Tokyo",
		Settings:  datatypes.NewJSONType(map[string]any{"api_version": "v2"}),
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.Require().NoError(s.db.Create(source).Error)

	dataTypes := []*DataType{
		{
			ID:                  uuid.Must(uuid.NewV7()),
			DataSourceID:        source.ID,
			Name:                "daily-quotes",
			Enabled:             true,
			UpdateFrequency:     "daily",
			UpdateTimes:         datatypes.NewJSONType([]string{"18:00", "20:00"}),
			BackfillEnabled:     true,
			StaleTimeoutMinutes: 30,
			Settings:            datatypes.NewJSONType(map[string]any{"endpoint": "/quotes"}),
			CreatedAt:           now,
			UpdatedAt:           now,
		},
		{
			ID:                  uuid.Must(uuid.NewV7()),
			DataSourceID:        source.ID,
			Name:                "listed-info",
			Enabled:             false,
			UpdateFrequency:     "weekly",
			UpdateTimes:         datatypes.NewJSONType([]string{"09:00"}),
			BackfillEnabled:     false,
			StaleTimeoutMinutes: 60,
			Settings:            datatypes.NewJSONType(map[string]any{}),
			CreatedAt:           now,
			UpdatedAt:           now,
		},
	}
	for _, dt := range dataTypes {
		s.Require().NoError(s.db.Create(dt).Error)
	}
	return source.ID
}

func (s *DataTypeRepositoryTestSuite) TestCreate() {
	ctx := context.Background()
	srcID := s.seedDataSource()

	dt := ingestion.NewDataType(
		srcID,
		"new-type",
		true,
		"hourly",
		[]string{"12:00"},
		false,
		15,
		map[string]any{"x": "y"},
	)
	created, err := s.repo.Create(ctx, dt)

	s.Require().NoError(err)
	s.NotEqual(uuid.Nil, created.ID())
	expected := ingestion.NewDataTypeDirectly(
		created.ID(),
		srcID,
		"new-type",
		true,
		"hourly",
		[]string{"12:00"},
		false,
		15,
		map[string]any{"x": "y"},
		created.CreatedAt(),
		created.UpdatedAt(),
	)
	s.True(cmp.Equal(*expected, *created, dataTypeCmpOpts...), cmp.Diff(*expected, *created, dataTypeCmpOpts...))
}

func (s *DataTypeRepositoryTestSuite) TestFindByID() {
	ctx := context.Background()
	tests := []struct {
		name    string
		setup   func() (uuid.UUID, *ingestion.DataType)
		wantNil bool
	}{
		{
			name: "found",
			setup: func() (uuid.UUID, *ingestion.DataType) {
				srcID := s.seedDataSource()
				types, err := s.repo.ListBySourceID(ctx, srcID)
				s.Require().NoError(err)
				s.Require().NotEmpty(types)
				return types[0].ID(), types[0]
			},
			wantNil: false,
		},
		{
			name: "not found",
			setup: func() (uuid.UUID, *ingestion.DataType) {
				return uuid.Must(uuid.NewV7()), nil
			},
			wantNil: true,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			id, expected := tt.setup()
			result, err := s.repo.FindByID(ctx, id)
			s.NoError(err)
			if tt.wantNil {
				s.Nil(result)
				return
			}
			s.Require().NotNil(result)
			s.True(cmp.Equal(*expected, *result, dataTypeCmpOpts...), cmp.Diff(*expected, *result, dataTypeCmpOpts...))
		})
	}
}

func (s *DataTypeRepositoryTestSuite) TestListBySourceID() {
	ctx := context.Background()
	srcID := s.seedDataSource()

	now := time.Now()
	anotherSource := &DataSource{
		ID:        uuid.Must(uuid.NewV7()),
		Name:      "another-source",
		Enabled:   true,
		Timezone:  "UTC",
		Settings:  datatypes.NewJSONType(map[string]any{}),
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.Require().NoError(s.db.Create(anotherSource).Error)
	s.Require().NoError(s.db.Create(&DataType{
		ID:                  uuid.Must(uuid.NewV7()),
		DataSourceID:        anotherSource.ID,
		Name:                "other-type",
		Enabled:             true,
		UpdateFrequency:     "daily",
		UpdateTimes:         datatypes.NewJSONType([]string{"00:00"}),
		BackfillEnabled:     false,
		StaleTimeoutMinutes: 60,
		Settings:            datatypes.NewJSONType(map[string]any{}),
		CreatedAt:           now,
		UpdatedAt:           now,
	}).Error)

	types, err := s.repo.ListBySourceID(ctx, srcID)

	s.Require().NoError(err)
	s.Len(types, 2)
}

func (s *DataTypeRepositoryTestSuite) TestUpdate() {
	ctx := context.Background()
	srcID := s.seedDataSource()

	types, err := s.repo.ListBySourceID(ctx, srcID)
	s.Require().NoError(err)
	s.Require().NotEmpty(types)
	origDT := types[0]

	origDT.Update(
		"daily-quotes-updated",
		false,
		"monthly",
		[]string{"06:00"},
		false,
		90,
		map[string]any{"endpoint": "/quotes/v2"},
	)
	err = s.repo.Update(ctx, origDT)
	s.Require().NoError(err)

	result, err := s.repo.FindByID(ctx, origDT.ID())
	s.Require().NoError(err)
	expected := ingestion.NewDataTypeDirectly(
		origDT.ID(),
		origDT.DataSourceID(),
		"daily-quotes-updated",
		false,
		"monthly",
		[]string{"06:00"},
		false,
		90,
		map[string]any{"endpoint": "/quotes/v2"},
		origDT.CreatedAt(),
		result.UpdatedAt(),
	)
	s.True(cmp.Equal(*expected, *result, dataTypeCmpOpts...), cmp.Diff(*expected, *result, dataTypeCmpOpts...))

	// Verify distractor types[1] (listed-info) was not affected
	distractor, err := s.repo.FindByID(ctx, types[1].ID())
	s.Require().NoError(err)
	s.Require().NotNil(distractor)
	s.Equal(types[1].Name(), distractor.Name())
}

func (s *DataTypeRepositoryTestSuite) TestDelete() {
	ctx := context.Background()
	srcID := s.seedDataSource()

	types, err := s.repo.ListBySourceID(ctx, srcID)
	s.Require().NoError(err)
	s.Require().NotEmpty(types)
	dtID := types[0].ID()

	err = s.repo.Delete(ctx, dtID)
	s.Require().NoError(err)

	result, err := s.repo.FindByID(ctx, dtID)
	s.NoError(err)
	s.Nil(result)

	// Verify distractor types[1] (listed-info) still exists
	distractor, err := s.repo.FindByID(ctx, types[1].ID())
	s.Require().NoError(err)
	s.Require().NotNil(distractor)
	s.Equal(types[1].Name(), distractor.Name())
}
