package usecase

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"stock-tool/internal/infra/repository"
	"stock-tool/internal/util/testutil"
)

type DataSourceUseCaseTestSuite struct {
	testutil.DBTest
	db   *gorm.DB
	repo *repository.DataSourceRepository
	uc   *DataSourceUseCase
}

func TestDataSourceUseCase(t *testing.T) {
	suite.Run(t, new(DataSourceUseCaseTestSuite))
}

func (s *DataSourceUseCaseTestSuite) SetupTest() {
	s.ApplyMigrations()

	db, err := s.RawDB().CreateGormDB()
	s.Require().NoError(err)

	s.db = db
	s.repo = repository.NewDataSourceRepository(db)
	s.uc = NewDataSourceUseCase(s.repo)
}

func (s *DataSourceUseCaseTestSuite) TearDownTest() {
	s.Require().NoError(s.CleanupMigrations())
}

func (s *DataSourceUseCaseTestSuite) TestCreate() {
	ctx := context.Background()
	cmpOpts := []cmp.Option{cmpopts.IgnoreFields(DataSourceResponse{}, "ID", "CreatedAt", "UpdatedAt")}

	type testCase struct {
		name        string
		req         *CreateDataSourceRequest
		expected    *DataSourceResponse
		expectErrAs *ValidationError
	}
	tests := []testCase{
		{
			name: "success",
			req: &CreateDataSourceRequest{
				Name:     "test-source",
				Enabled:  true,
				Timezone: "Asia/Tokyo",
				Settings: map[string]any{"key": "val"},
			},
			expected: &DataSourceResponse{
				Name:     "test-source",
				Enabled:  true,
				Timezone: "Asia/Tokyo",
				Settings: map[string]any{"key": "val"},
			},
		},
		{
			name: "invalid timezone",
			req: &CreateDataSourceRequest{
				Name:     "test",
				Enabled:  true,
				Timezone: "Invalid/Zone",
				Settings: map[string]any{},
			},
			expectErrAs: &ValidationError{},
		},
		{
			name: "duplicate name",
			req: &CreateDataSourceRequest{
				Name:     "test-source",
				Enabled:  false,
				Timezone: "UTC",
				Settings: map[string]any{},
			},
			expectErrAs: &ValidationError{},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			resp, err := s.uc.Create(ctx, tc.req)
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

func (s *DataSourceUseCaseTestSuite) TestGet() {
	ctx := context.Background()
	cmpOpts := []cmp.Option{cmpopts.IgnoreFields(DataSourceResponse{}, "ID", "CreatedAt", "UpdatedAt")}

	type testCase struct {
		name     string
		setup    func() uuid.UUID
		expected *DataSourceResponse
	}
	tests := []testCase{
		{
			name: "found",
			setup: func() uuid.UUID {
				_, err := s.uc.Create(ctx, &CreateDataSourceRequest{
					Name:     "src-other",
					Enabled:  false,
					Timezone: "UTC",
					Settings: map[string]any{},
				})
				s.Require().NoError(err)
				created, err := s.uc.Create(ctx, &CreateDataSourceRequest{
					Name:     "src",
					Enabled:  true,
					Timezone: "UTC",
					Settings: map[string]any{},
				})
				s.Require().NoError(err)
				return created.ID
			},
			expected: &DataSourceResponse{
				Name:     "src",
				Enabled:  true,
				Timezone: "UTC",
				Settings: map[string]any{},
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
			resp, err := s.uc.Get(ctx, id)
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

func (s *DataSourceUseCaseTestSuite) TestList() {
	ctx := context.Background()
	cmpOpts := []cmp.Option{
		cmpopts.IgnoreFields(DataSourceResponse{}, "ID", "CreatedAt", "UpdatedAt"),
		cmpopts.SortSlices(func(a, b *DataSourceResponse) bool { return a.Name < b.Name }),
	}

	_, err := s.uc.Create(ctx, &CreateDataSourceRequest{
		Name: "src1", Enabled: true, Timezone: "UTC", Settings: map[string]any{},
	})
	s.Require().NoError(err)
	_, err = s.uc.Create(ctx, &CreateDataSourceRequest{
		Name: "src2", Enabled: false, Timezone: "US/Eastern", Settings: map[string]any{},
	})
	s.Require().NoError(err)

	list, err := s.uc.List(ctx)

	s.Require().NoError(err)
	expected := []*DataSourceResponse{
		{Name: "src1", Enabled: true, Timezone: "UTC", Settings: map[string]any{}},
		{Name: "src2", Enabled: false, Timezone: "US/Eastern", Settings: map[string]any{}},
	}
	s.True(cmp.Equal(expected, list, cmpOpts...), cmp.Diff(expected, list, cmpOpts...))
}

func (s *DataSourceUseCaseTestSuite) TestUpdate() {
	ctx := context.Background()
	cmpOpts := []cmp.Option{cmpopts.IgnoreFields(DataSourceResponse{}, "ID", "CreatedAt", "UpdatedAt")}

	type testCase struct {
		name        string
		setup       func() uuid.UUID
		req         func(id uuid.UUID) *UpdateDataSourceRequest
		expected    *DataSourceResponse
		expectErrAs *ValidationError
		postCheck   func()
	}
	var srcOtherID uuid.UUID
	tests := []testCase{
		{
			name: "success",
			setup: func() uuid.UUID {
				srcOther, err := s.uc.Create(ctx, &CreateDataSourceRequest{
					Name:     "src-other",
					Enabled:  false,
					Timezone: "UTC",
					Settings: map[string]any{},
				})
				s.Require().NoError(err)
				srcOtherID = srcOther.ID
				created, err := s.uc.Create(ctx, &CreateDataSourceRequest{
					Name:     "src",
					Enabled:  true,
					Timezone: "UTC",
					Settings: map[string]any{},
				})
				s.Require().NoError(err)
				return created.ID
			},
			req: func(id uuid.UUID) *UpdateDataSourceRequest {
				return &UpdateDataSourceRequest{
					ID:       id,
					Name:     "src-updated",
					Enabled:  false,
					Timezone: "Asia/Tokyo",
					Settings: map[string]any{"new": "setting"},
				}
			},
			expected: &DataSourceResponse{
				Name:     "src-updated",
				Enabled:  false,
				Timezone: "Asia/Tokyo",
				Settings: map[string]any{"new": "setting"},
			},
			postCheck: func() {
				resp, err := s.uc.Get(ctx, srcOtherID)
				s.Require().NoError(err)
				s.Require().NotNil(resp)
				s.Equal("src-other", resp.Name)
			},
		},
		{
			name:  "not found",
			setup: func() uuid.UUID { return uuid.Must(uuid.NewV7()) },
			req: func(id uuid.UUID) *UpdateDataSourceRequest {
				return &UpdateDataSourceRequest{
					ID:       id,
					Name:     "x",
					Enabled:  true,
					Timezone: "UTC",
					Settings: map[string]any{},
				}
			},
			expected: nil,
		},
		{
			name: "invalid timezone",
			setup: func() uuid.UUID {
				created, err := s.uc.Create(ctx, &CreateDataSourceRequest{
					Name:     "src",
					Enabled:  true,
					Timezone: "UTC",
					Settings: map[string]any{},
				})
				s.Require().NoError(err)
				return created.ID
			},
			req: func(id uuid.UUID) *UpdateDataSourceRequest {
				return &UpdateDataSourceRequest{
					ID:       id,
					Name:     "x",
					Enabled:  true,
					Timezone: "Bad/Zone",
					Settings: map[string]any{},
				}
			},
			expectErrAs: &ValidationError{},
		},
		{
			name: "duplicate name",
			setup: func() uuid.UUID {
				_, err := s.uc.Create(ctx, &CreateDataSourceRequest{
					Name:     "existing-src",
					Enabled:  true,
					Timezone: "UTC",
					Settings: map[string]any{},
				})
				s.Require().NoError(err)
				created, err := s.uc.Create(ctx, &CreateDataSourceRequest{
					Name:     "rename-target",
					Enabled:  true,
					Timezone: "UTC",
					Settings: map[string]any{},
				})
				s.Require().NoError(err)
				return created.ID
			},
			req: func(id uuid.UUID) *UpdateDataSourceRequest {
				return &UpdateDataSourceRequest{
					ID:       id,
					Name:     "existing-src",
					Enabled:  true,
					Timezone: "UTC",
					Settings: map[string]any{},
				}
			},
			expectErrAs: &ValidationError{},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			id := tc.setup()
			resp, err := s.uc.Update(ctx, tc.req(id))
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

func (s *DataSourceUseCaseTestSuite) TestDelete() {
	ctx := context.Background()

	srcOther, err := s.uc.Create(ctx, &CreateDataSourceRequest{
		Name:     "src-other",
		Enabled:  false,
		Timezone: "UTC",
		Settings: map[string]any{},
	})
	s.Require().NoError(err)

	created, err := s.uc.Create(ctx, &CreateDataSourceRequest{
		Name:     "src",
		Enabled:  true,
		Timezone: "UTC",
		Settings: map[string]any{},
	})
	s.Require().NoError(err)

	err = s.uc.Delete(ctx, created.ID)
	s.Require().NoError(err)

	resp, err := s.uc.Get(ctx, created.ID)
	s.NoError(err)
	s.Nil(resp)

	// Verify distractor src-other still exists
	otherResp, err := s.uc.Get(ctx, srcOther.ID)
	s.Require().NoError(err)
	s.Require().NotNil(otherResp)
	s.Equal("src-other", otherResp.Name)
}
