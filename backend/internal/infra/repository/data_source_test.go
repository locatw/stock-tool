package repository

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/suite"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"stock-tool/internal/domain/ingestion"
	"stock-tool/internal/util/idp"
	"stock-tool/internal/util/testutil"
)

var dataSrcCmpOpts = cmp.Options{
	cmp.AllowUnexported(ingestion.DataSource{}),
	cmp.Comparer(func(a, b *time.Location) bool {
		if a == nil && b == nil {
			return true
		}
		if a == nil || b == nil {
			return false
		}
		return a.String() == b.String()
	}),
	cmp.Comparer(func(a, b time.Time) bool {
		return a.Round(time.Microsecond).Equal(b.Round(time.Microsecond))
	}),
}

type DataSourceRepositoryTestSuite struct {
	testutil.DBTest
	repo *DataSourceRepository
	db   *gorm.DB
}

func TestDataSourceRepository(t *testing.T) {
	suite.Run(t, new(DataSourceRepositoryTestSuite))
}

func (s *DataSourceRepositoryTestSuite) SetupTest() {
	s.ApplyMigrations()

	db, err := s.RawDB().CreateGormDB()
	s.Require().NoError(err)

	s.db = db
	s.repo = NewDataSourceRepository(db)
}

func (s *DataSourceRepositoryTestSuite) TearDownTest() {
	s.Require().NoError(s.CleanupMigrations())
}

func (s *DataSourceRepositoryTestSuite) seedDataSource() uuid.UUID {
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
			Schedule:            datatypes.NewJSONType(scheduleJSON{Type: "daily", Times: []string{"18:00", "20:00"}}),
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
			Schedule:            datatypes.NewJSONType(scheduleJSON{Type: "daily", Times: []string{"09:00", "12:00"}}),
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

func (s *DataSourceRepositoryTestSuite) TestFindByID() {
	ctx := context.Background()
	tests := []struct {
		name    string
		setup   func() (uuid.UUID, *ingestion.DataSource)
		wantNil bool
	}{
		{
			name: "found",
			setup: func() (uuid.UUID, *ingestion.DataSource) {
				s.seedDataSource()
				src, err := ingestion.NewDataSource(ctx, "another-source", false, "UTC", map[string]any{})
				s.Require().NoError(err)
				_, err = s.repo.Create(ctx, src)
				s.Require().NoError(err)
				return src.ID(), src
			},
			wantNil: false,
		},
		{
			name: "not found",
			setup: func() (uuid.UUID, *ingestion.DataSource) {
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
			s.True(cmp.Equal(*expected, *result, dataSrcCmpOpts...), cmp.Diff(*expected, *result, dataSrcCmpOpts...))
		})
	}
}

func (s *DataSourceRepositoryTestSuite) TestCreate() {
	fixedID := uuid.MustParse("01961f1a-89c4-7641-b052-4dca477a457a")
	ctx := idp.WithFixedID(context.Background(), fixedID)

	src, err := ingestion.NewDataSource(ctx, "test-source", true, "UTC", map[string]any{"key": "val"})
	s.Require().NoError(err)
	created, err := s.repo.Create(ctx, src)

	s.Require().NoError(err)
	loc, _ := time.LoadLocation("UTC")
	expected := ingestion.NewDataSourceDirectly(
		fixedID,
		"test-source",
		true,
		loc,
		map[string]any{"key": "val"},
		created.CreatedAt(),
		created.UpdatedAt(),
	)
	s.True(cmp.Equal(*expected, *created, dataSrcCmpOpts...), cmp.Diff(*expected, *created, dataSrcCmpOpts...))
}

func (s *DataSourceRepositoryTestSuite) TestList() {
	ctx := context.Background()
	s.seedDataSource()

	// Create a second source
	src, err := ingestion.NewDataSource(ctx, "another-source", false, "US/Eastern", map[string]any{})
	s.Require().NoError(err)
	_, err = s.repo.Create(ctx, src)
	s.Require().NoError(err)

	list, err := s.repo.List(ctx)

	s.Require().NoError(err)
	s.Require().Len(list, 2)
	names := lo.Map(list, func(ds *ingestion.DataSource, _ int) string { return ds.Name() })
	sort.Strings(names)
	s.Equal([]string{"another-source", "j-quants"}, names)
}

func (s *DataSourceRepositoryTestSuite) TestUpdate() {
	ctx := context.Background()
	seededID := s.seedDataSource()

	// distractor: should remain unchanged after the update
	anotherSrc, err := ingestion.NewDataSource(ctx, "another-source", false, "UTC", map[string]any{})
	s.Require().NoError(err)
	anotherCreated, err := s.repo.Create(ctx, anotherSrc)
	s.Require().NoError(err)

	found, err := s.repo.FindByID(ctx, seededID)
	s.Require().NoError(err)

	s.Require().NoError(found.Update("j-quants-updated", false, "US/Eastern", map[string]any{"api_version": "v3"}))
	err = s.repo.Update(ctx, found)
	s.Require().NoError(err)

	result, err := s.repo.FindByID(ctx, found.ID())
	s.Require().NoError(err)
	loc, _ := time.LoadLocation("US/Eastern")
	expected := ingestion.NewDataSourceDirectly(
		found.ID(),
		"j-quants-updated",
		false,
		loc,
		map[string]any{"api_version": "v3"},
		found.CreatedAt(),
		result.UpdatedAt(),
	)
	s.True(cmp.Equal(*expected, *result, dataSrcCmpOpts...), cmp.Diff(*expected, *result, dataSrcCmpOpts...))

	// Verify distractor was not affected
	anotherResult, err := s.repo.FindByID(ctx, anotherCreated.ID())
	s.Require().NoError(err)
	s.Require().NotNil(anotherResult)
	s.Equal("another-source", anotherResult.Name())
}

func (s *DataSourceRepositoryTestSuite) TestDelete() {
	ctx := context.Background()
	seededID := s.seedDataSource()

	// distractor: should survive the delete
	anotherSrc, err := ingestion.NewDataSource(ctx, "another-source", false, "UTC", map[string]any{})
	s.Require().NoError(err)
	anotherCreated, err := s.repo.Create(ctx, anotherSrc)
	s.Require().NoError(err)

	found, err := s.repo.FindByID(ctx, seededID)
	s.Require().NoError(err)

	err = s.repo.Delete(ctx, found.ID())
	s.Require().NoError(err)

	result, err := s.repo.FindByID(ctx, found.ID())
	s.NoError(err)
	s.Nil(result)

	var count int64
	s.Require().NoError(s.db.Model(&DataType{}).Where("data_source_id = ?", found.ID()).Count(&count).Error)
	s.Equal(int64(0), count)

	// Verify distractor still exists
	anotherResult, err := s.repo.FindByID(ctx, anotherCreated.ID())
	s.Require().NoError(err)
	s.Require().NotNil(anotherResult)
	s.Equal("another-source", anotherResult.Name())
}
