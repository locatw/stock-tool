package usecase

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"stock-tool/database"
	"stock-tool/internal/infra/repository"
	"stock-tool/internal/util/testutil"
)

type DataTypeUseCaseTestSuite struct {
	testutil.DBTest
	db        *gorm.DB
	dsRepo    *repository.DataSourceRepository
	dtypeRepo *repository.DataTypeRepository
	dsUC      *DataSourceUseCase
	dtUC      *DataTypeUseCase
}

func TestDataTypeUseCase(t *testing.T) {
	suite.Run(t, new(DataTypeUseCaseTestSuite))
}

func (s *DataTypeUseCaseTestSuite) SetupTest() {
	s.ApplyMigrations()

	db, err := database.CreateGormDB(s.GetDB())
	s.Require().NoError(err)

	s.db = db
	s.dsRepo = repository.NewDataSourceRepository(db)
	s.dtypeRepo = repository.NewDataTypeRepository(db)
	s.dsUC = NewDataSourceUseCase(s.dsRepo)
	s.dtUC = NewDataTypeUseCase(s.dtypeRepo)
}

func (s *DataTypeUseCaseTestSuite) TearDownTest() {
	s.Require().NoError(s.CleanupMigrations())
}

func (s *DataTypeUseCaseTestSuite) TestCreate() {
	ctx := context.Background()
	cmpOpts := []cmp.Option{cmpopts.IgnoreFields(DataTypeResponse{}, "ID", "DataSourceID", "CreatedAt", "UpdatedAt")}

	type testCase struct {
		name        string
		setup       func() uuid.UUID
		req         func(id uuid.UUID) *CreateDataTypeRequest
		expected    *DataTypeResponse
		expectErrAs *ValidationError
	}
	tests := []testCase{
		{
			name: "success",
			setup: func() uuid.UUID {
				src, err := s.dsUC.Create(ctx, &CreateDataSourceRequest{
					Name:     "src",
					Enabled:  true,
					Timezone: "UTC",
					Settings: map[string]any{},
				})
				s.Require().NoError(err)
				return src.ID
			},
			req: func(id uuid.UUID) *CreateDataTypeRequest {
				return &CreateDataTypeRequest{
					DataSourceID:        id,
					Name:                "daily-quotes",
					Enabled:             true,
					UpdateFrequency:     "daily",
					UpdateTimes:         []string{"18:00"},
					BackfillEnabled:     true,
					StaleTimeoutMinutes: 30,
					Settings:            map[string]any{},
				}
			},
			expected: &DataTypeResponse{
				Name:                "daily-quotes",
				Enabled:             true,
				UpdateFrequency:     "daily",
				UpdateTimes:         []string{"18:00"},
				BackfillEnabled:     true,
				StaleTimeoutMinutes: 30,
				Settings:            map[string]any{},
			},
		},
		{
			name:  "invalid frequency",
			setup: func() uuid.UUID { return uuid.Must(uuid.NewV7()) },
			req: func(id uuid.UUID) *CreateDataTypeRequest {
				return &CreateDataTypeRequest{
					DataSourceID:    id,
					Name:            "dt",
					Enabled:         true,
					UpdateFrequency: "biweekly",
					UpdateTimes:     []string{},
					Settings:        map[string]any{},
				}
			},
			expectErrAs: &ValidationError{},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			id := tc.setup()
			resp, err := s.dtUC.Create(ctx, tc.req(id))
			if tc.expectErrAs != nil {
				s.Require().Error(err)
				s.IsType(tc.expectErrAs, err)
				return
			}
			s.Require().NoError(err)
			s.True(cmp.Equal(tc.expected, resp, cmpOpts...), cmp.Diff(tc.expected, resp, cmpOpts...))
		})
	}
}

func (s *DataTypeUseCaseTestSuite) TestGet() {
	ctx := context.Background()
	cmpOpts := []cmp.Option{cmpopts.IgnoreFields(DataTypeResponse{}, "ID", "DataSourceID", "CreatedAt", "UpdatedAt")}

	type testCase struct {
		name     string
		setup    func() uuid.UUID
		expected *DataTypeResponse
	}
	tests := []testCase{
		{
			name: "found",
			setup: func() uuid.UUID {
				src, err := s.dsUC.Create(ctx, &CreateDataSourceRequest{
					Name:     "src",
					Enabled:  true,
					Timezone: "UTC",
					Settings: map[string]any{},
				})
				s.Require().NoError(err)
				_, err = s.dtUC.Create(ctx, &CreateDataTypeRequest{
					DataSourceID:    src.ID,
					Name:            "dt-other",
					Enabled:         false,
					UpdateFrequency: "weekly",
					UpdateTimes:     []string{},
					Settings:        map[string]any{},
				})
				s.Require().NoError(err)
				dt, err := s.dtUC.Create(ctx, &CreateDataTypeRequest{
					DataSourceID:    src.ID,
					Name:            "dt",
					Enabled:         true,
					UpdateFrequency: "daily",
					UpdateTimes:     []string{},
					Settings:        map[string]any{},
				})
				s.Require().NoError(err)
				return dt.ID
			},
			expected: &DataTypeResponse{
				Name:            "dt",
				Enabled:         true,
				UpdateFrequency: "daily",
				UpdateTimes:     []string{},
				Settings:        map[string]any{},
			},
		},
		{
			name:     "not found",
			setup:    func() uuid.UUID { return uuid.Must(uuid.NewV7()) },
			expected: nil,
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			id := tc.setup()
			resp, err := s.dtUC.Get(ctx, id)
			s.Require().NoError(err)
			if tc.expected == nil {
				s.Nil(resp)
				return
			}
			s.Require().NotNil(resp)
			s.True(cmp.Equal(tc.expected, resp, cmpOpts...), cmp.Diff(tc.expected, resp, cmpOpts...))
		})
	}
}

func (s *DataTypeUseCaseTestSuite) TestListDataTypes() {
	ctx := context.Background()
	cmpOpts := []cmp.Option{
		cmpopts.IgnoreFields(DataTypeResponse{}, "ID", "DataSourceID", "CreatedAt", "UpdatedAt"),
		cmpopts.SortSlices(func(a, b *DataTypeResponse) bool { return a.Name < b.Name }),
	}

	src1, err := s.dsUC.Create(ctx, &CreateDataSourceRequest{
		Name:     "src1",
		Enabled:  true,
		Timezone: "UTC",
		Settings: map[string]any{},
	})
	s.Require().NoError(err)
	src2, err := s.dsUC.Create(ctx, &CreateDataSourceRequest{
		Name:     "src2",
		Enabled:  true,
		Timezone: "UTC",
		Settings: map[string]any{},
	})
	s.Require().NoError(err)

	_, err = s.dtUC.Create(ctx, &CreateDataTypeRequest{
		DataSourceID:    src1.ID,
		Name:            "dt1",
		Enabled:         true,
		UpdateFrequency: "daily",
		UpdateTimes:     []string{},
		Settings:        map[string]any{},
	})
	s.Require().NoError(err)
	_, err = s.dtUC.Create(ctx, &CreateDataTypeRequest{
		DataSourceID:    src1.ID,
		Name:            "dt2",
		Enabled:         false,
		UpdateFrequency: "weekly",
		UpdateTimes:     []string{"09:00"},
		Settings:        map[string]any{},
	})
	s.Require().NoError(err)
	_, err = s.dtUC.Create(ctx, &CreateDataTypeRequest{
		DataSourceID:    src2.ID,
		Name:            "dt3",
		Enabled:         true,
		UpdateFrequency: "daily",
		UpdateTimes:     []string{},
		Settings:        map[string]any{},
	})
	s.Require().NoError(err)

	list, err := s.dtUC.List(ctx, src1.ID)

	s.Require().NoError(err)
	expected := []*DataTypeResponse{
		{Name: "dt1", Enabled: true, UpdateFrequency: "daily", UpdateTimes: []string{}, Settings: map[string]any{}},
		{Name: "dt2", Enabled: false, UpdateFrequency: "weekly", UpdateTimes: []string{"09:00"}, Settings: map[string]any{}},
	}
	s.True(cmp.Equal(expected, list, cmpOpts...), cmp.Diff(expected, list, cmpOpts...))
}

func (s *DataTypeUseCaseTestSuite) TestUpdate() {
	ctx := context.Background()
	cmpOpts := []cmp.Option{cmpopts.IgnoreFields(DataTypeResponse{}, "ID", "DataSourceID", "CreatedAt", "UpdatedAt")}

	type testCase struct {
		name        string
		setup       func() uuid.UUID
		req         func(id uuid.UUID) *UpdateDataTypeRequest
		expected    *DataTypeResponse
		expectErrAs *ValidationError
		postCheck   func()
	}
	var dtOtherID uuid.UUID
	tests := []testCase{
		{
			name: "success",
			setup: func() uuid.UUID {
				src, err := s.dsUC.Create(ctx, &CreateDataSourceRequest{
					Name:     "src",
					Enabled:  true,
					Timezone: "UTC",
					Settings: map[string]any{},
				})
				s.Require().NoError(err)
				dtOther, err := s.dtUC.Create(ctx, &CreateDataTypeRequest{
					DataSourceID:    src.ID,
					Name:            "dt-other",
					Enabled:         false,
					UpdateFrequency: "weekly",
					UpdateTimes:     []string{},
					Settings:        map[string]any{},
				})
				s.Require().NoError(err)
				dtOtherID = dtOther.ID
				dt, err := s.dtUC.Create(ctx, &CreateDataTypeRequest{
					DataSourceID:    src.ID,
					Name:            "dt",
					Enabled:         true,
					UpdateFrequency: "daily",
					UpdateTimes:     []string{"18:00"},
					Settings:        map[string]any{},
				})
				s.Require().NoError(err)
				return dt.ID
			},
			req: func(id uuid.UUID) *UpdateDataTypeRequest {
				return &UpdateDataTypeRequest{
					ID:              id,
					Name:            "dt-updated",
					Enabled:         false,
					UpdateFrequency: "weekly",
					UpdateTimes:     []string{"09:00"},
					Settings:        map[string]any{"x": "y"},
				}
			},
			expected: &DataTypeResponse{
				Name:            "dt-updated",
				Enabled:         false,
				UpdateFrequency: "weekly",
				UpdateTimes:     []string{"09:00"},
				Settings:        map[string]any{"x": "y"},
			},
			postCheck: func() {
				resp, err := s.dtUC.Get(ctx, dtOtherID)
				s.Require().NoError(err)
				s.Require().NotNil(resp)
				s.Equal("dt-other", resp.Name)
			},
		},
		{
			name:  "not found",
			setup: func() uuid.UUID { return uuid.Must(uuid.NewV7()) },
			req: func(id uuid.UUID) *UpdateDataTypeRequest {
				return &UpdateDataTypeRequest{
					ID:              id,
					Name:            "x",
					UpdateFrequency: "daily",
					UpdateTimes:     []string{},
					Settings:        map[string]any{},
				}
			},
			expected: nil,
		},
		{
			name:  "invalid frequency",
			setup: func() uuid.UUID { return uuid.Must(uuid.NewV7()) },
			req: func(id uuid.UUID) *UpdateDataTypeRequest {
				return &UpdateDataTypeRequest{
					ID:              id,
					Name:            "x",
					UpdateFrequency: "biweekly",
					UpdateTimes:     []string{},
					Settings:        map[string]any{},
				}
			},
			expectErrAs: &ValidationError{},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			id := tc.setup()
			resp, err := s.dtUC.Update(ctx, tc.req(id))
			if tc.expectErrAs != nil {
				s.Require().Error(err)
				s.IsType(tc.expectErrAs, err)
				return
			}
			s.Require().NoError(err)
			if tc.expected == nil {
				s.Nil(resp)
				return
			}
			s.Require().NotNil(resp)
			s.True(cmp.Equal(tc.expected, resp, cmpOpts...), cmp.Diff(tc.expected, resp, cmpOpts...))
			if tc.postCheck != nil {
				tc.postCheck()
			}
		})
	}
}

func (s *DataTypeUseCaseTestSuite) TestDeleteDataType() {
	ctx := context.Background()

	src, err := s.dsUC.Create(ctx, &CreateDataSourceRequest{
		Name:     "src",
		Enabled:  true,
		Timezone: "UTC",
		Settings: map[string]any{},
	})
	s.Require().NoError(err)
	dtOther, err := s.dtUC.Create(ctx, &CreateDataTypeRequest{
		DataSourceID:    src.ID,
		Name:            "dt-other",
		Enabled:         false,
		UpdateFrequency: "weekly",
		UpdateTimes:     []string{},
		Settings:        map[string]any{},
	})
	s.Require().NoError(err)
	dt, err := s.dtUC.Create(ctx, &CreateDataTypeRequest{
		DataSourceID:    src.ID,
		Name:            "dt",
		Enabled:         true,
		UpdateFrequency: "daily",
		UpdateTimes:     []string{},
		Settings:        map[string]any{},
	})
	s.Require().NoError(err)

	err = s.dtUC.Delete(ctx, dt.ID)
	s.Require().NoError(err)

	resp, err := s.dtUC.Get(ctx, dt.ID)
	s.NoError(err)
	s.Nil(resp)

	// Verify distractor dt-other still exists
	otherResp, err := s.dtUC.Get(ctx, dtOther.ID)
	s.Require().NoError(err)
	s.Require().NotNil(otherResp)
	s.Equal("dt-other", otherResp.Name)
}
