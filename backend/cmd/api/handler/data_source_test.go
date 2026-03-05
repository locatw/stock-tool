package handler

import (
	"context"
	"testing"
	"time"

	api "stock-tool/api/gen"
	"stock-tool/internal/usecase"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type DataSourceUseCaseMock struct {
	mock.Mock
}

func (m *DataSourceUseCaseMock) Create(
	ctx context.Context,
	req *usecase.CreateDataSourceRequest,
) (*usecase.DataSourceResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.DataSourceResponse), args.Error(1)
}

func (m *DataSourceUseCaseMock) Get(
	ctx context.Context,
	id uuid.UUID,
) (*usecase.DataSourceResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.DataSourceResponse), args.Error(1)
}

func (m *DataSourceUseCaseMock) List(
	ctx context.Context,
) ([]*usecase.DataSourceResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*usecase.DataSourceResponse), args.Error(1)
}

func (m *DataSourceUseCaseMock) Update(
	ctx context.Context,
	req *usecase.UpdateDataSourceRequest,
) (*usecase.DataSourceResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.DataSourceResponse), args.Error(1)
}

func (m *DataSourceUseCaseMock) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

type DataSourceHandlerTestSuite struct {
	suite.Suite
	ucMock  *DataSourceUseCaseMock
	handler *DataSourceHandler
}

func TestDataSourceHandler(t *testing.T) {
	suite.Run(t, new(DataSourceHandlerTestSuite))
}

func (s *DataSourceHandlerTestSuite) SetupTest() {
	s.ucMock = new(DataSourceUseCaseMock)
	s.handler = &DataSourceHandler{uc: s.ucMock}
}

func (s *DataSourceHandlerTestSuite) TestListDataSources() {
	now := time.Now()
	id1 := uuid.Must(uuid.NewV7())
	s.ucMock.On("List", mock.Anything).Return([]*usecase.DataSourceResponse{
		{
			ID:        id1,
			Name:      "src1",
			Enabled:   true,
			Timezone:  "UTC",
			Settings:  map[string]any{},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}, nil)

	resp, err := s.handler.ListDataSources(context.Background(), api.ListDataSourcesRequestObject{})

	expected := api.ListDataSources200JSONResponse{
		{Id: id1, Name: "src1", Enabled: true, Timezone: "UTC", Settings: map[string]any{}, CreatedAt: now, UpdatedAt: now},
	}
	s.NoError(err)
	s.Require().IsType(api.ListDataSources200JSONResponse{}, resp)
	s.True(cmp.Equal(expected, resp.(api.ListDataSources200JSONResponse)), cmp.Diff(expected, resp.(api.ListDataSources200JSONResponse)))
}

func (s *DataSourceHandlerTestSuite) TestCreateDataSource_Success() {
	now := time.Now()
	id1 := uuid.Must(uuid.NewV7())
	expectedReq := &usecase.CreateDataSourceRequest{Name: "new-src", Enabled: true, Timezone: "UTC", Settings: map[string]any{}}
	s.ucMock.On("Create", mock.Anything, expectedReq).Return(&usecase.DataSourceResponse{
		ID:        id1,
		Name:      "new-src",
		Enabled:   true,
		Timezone:  "UTC",
		Settings:  map[string]any{},
		CreatedAt: now,
		UpdatedAt: now,
	}, nil)

	body := &api.CreateDataSourceRequest{Name: "new-src", Enabled: true, Timezone: "UTC", Settings: map[string]any{}}
	resp, err := s.handler.CreateDataSource(context.Background(), api.CreateDataSourceRequestObject{Body: body})

	expected := api.CreateDataSource201JSONResponse{Id: id1, Name: "new-src", Enabled: true, Timezone: "UTC", Settings: map[string]any{}, CreatedAt: now, UpdatedAt: now}
	s.NoError(err)
	s.Require().IsType(api.CreateDataSource201JSONResponse{}, resp)
	s.True(cmp.Equal(expected, resp.(api.CreateDataSource201JSONResponse)), cmp.Diff(expected, resp.(api.CreateDataSource201JSONResponse)))
}

func (s *DataSourceHandlerTestSuite) TestCreateDataSource_ValidationError() {
	expectedReq := &usecase.CreateDataSourceRequest{Name: "src", Enabled: true, Timezone: "Bad/Zone", Settings: map[string]any{}}
	s.ucMock.On("Create", mock.Anything, expectedReq).
		Return(nil, &usecase.ValidationError{Message: "invalid timezone: Bad/Zone"})

	body := &api.CreateDataSourceRequest{Name: "src", Enabled: true, Timezone: "Bad/Zone", Settings: map[string]any{}}
	resp, err := s.handler.CreateDataSource(context.Background(), api.CreateDataSourceRequestObject{Body: body})

	expected := api.CreateDataSource422JSONResponse{Error: "invalid timezone: Bad/Zone"}
	s.NoError(err)
	s.Require().IsType(api.CreateDataSource422JSONResponse{}, resp)
	s.True(cmp.Equal(expected, resp.(api.CreateDataSource422JSONResponse)), cmp.Diff(expected, resp.(api.CreateDataSource422JSONResponse)))
}

func (s *DataSourceHandlerTestSuite) TestGetDataSource_Found() {
	now := time.Now()
	id1 := uuid.Must(uuid.NewV7())
	s.ucMock.On("Get", mock.Anything, id1).Return(&usecase.DataSourceResponse{
		ID:        id1,
		Name:      "src",
		Enabled:   true,
		Timezone:  "UTC",
		Settings:  map[string]any{},
		CreatedAt: now,
		UpdatedAt: now,
	}, nil)

	resp, err := s.handler.GetDataSource(context.Background(), api.GetDataSourceRequestObject{Id: id1})

	expected := api.GetDataSource200JSONResponse{Id: id1, Name: "src", Enabled: true, Timezone: "UTC", Settings: map[string]any{}, CreatedAt: now, UpdatedAt: now}
	s.NoError(err)
	s.Require().IsType(api.GetDataSource200JSONResponse{}, resp)
	s.True(cmp.Equal(expected, resp.(api.GetDataSource200JSONResponse)), cmp.Diff(expected, resp.(api.GetDataSource200JSONResponse)))
}

func (s *DataSourceHandlerTestSuite) TestGetDataSource_NotFound() {
	notFoundID := uuid.Must(uuid.NewV7())
	s.ucMock.On("Get", mock.Anything, notFoundID).Return(nil, nil)

	resp, err := s.handler.GetDataSource(context.Background(), api.GetDataSourceRequestObject{Id: notFoundID})

	expected := api.GetDataSource404JSONResponse{Error: "data source not found"}
	s.NoError(err)
	s.Require().IsType(api.GetDataSource404JSONResponse{}, resp)
	s.True(cmp.Equal(expected, resp.(api.GetDataSource404JSONResponse)), cmp.Diff(expected, resp.(api.GetDataSource404JSONResponse)))
}

func (s *DataSourceHandlerTestSuite) TestUpdateDataSource_Success() {
	now := time.Now()
	id1 := uuid.Must(uuid.NewV7())
	expectedReq := &usecase.UpdateDataSourceRequest{ID: id1, Name: "updated", Enabled: false, Timezone: "Asia/Tokyo", Settings: map[string]any{}}
	s.ucMock.On("Update", mock.Anything, expectedReq).Return(&usecase.DataSourceResponse{
		ID:        id1,
		Name:      "updated",
		Enabled:   false,
		Timezone:  "Asia/Tokyo",
		Settings:  map[string]any{},
		CreatedAt: now,
		UpdatedAt: now,
	}, nil)

	body := &api.UpdateDataSourceRequest{Name: "updated", Enabled: false, Timezone: "Asia/Tokyo", Settings: map[string]any{}}
	resp, err := s.handler.UpdateDataSource(context.Background(), api.UpdateDataSourceRequestObject{Id: id1, Body: body})

	expected := api.UpdateDataSource200JSONResponse{Id: id1, Name: "updated", Enabled: false, Timezone: "Asia/Tokyo", Settings: map[string]any{}, CreatedAt: now, UpdatedAt: now}
	s.NoError(err)
	s.Require().IsType(api.UpdateDataSource200JSONResponse{}, resp)
	s.True(cmp.Equal(expected, resp.(api.UpdateDataSource200JSONResponse)), cmp.Diff(expected, resp.(api.UpdateDataSource200JSONResponse)))
}

func (s *DataSourceHandlerTestSuite) TestUpdateDataSource_NotFound() {
	notFoundID := uuid.Must(uuid.NewV7())
	expectedReq := &usecase.UpdateDataSourceRequest{ID: notFoundID, Name: "x", Enabled: true, Timezone: "UTC", Settings: map[string]any{}}
	s.ucMock.On("Update", mock.Anything, expectedReq).Return(nil, nil)

	body := &api.UpdateDataSourceRequest{Name: "x", Enabled: true, Timezone: "UTC", Settings: map[string]any{}}
	resp, err := s.handler.UpdateDataSource(context.Background(), api.UpdateDataSourceRequestObject{Id: notFoundID, Body: body})

	expected := api.UpdateDataSource404JSONResponse{Error: "data source not found"}
	s.NoError(err)
	s.Require().IsType(api.UpdateDataSource404JSONResponse{}, resp)
	s.True(cmp.Equal(expected, resp.(api.UpdateDataSource404JSONResponse)), cmp.Diff(expected, resp.(api.UpdateDataSource404JSONResponse)))
}

func (s *DataSourceHandlerTestSuite) TestDeleteDataSource() {
	id1 := uuid.Must(uuid.NewV7())
	s.ucMock.On("Delete", mock.Anything, id1).Return(nil)

	resp, err := s.handler.DeleteDataSource(context.Background(), api.DeleteDataSourceRequestObject{Id: id1})

	expected := api.DeleteDataSource204Response{}
	s.NoError(err)
	s.Require().IsType(api.DeleteDataSource204Response{}, resp)
	s.True(cmp.Equal(expected, resp.(api.DeleteDataSource204Response)), cmp.Diff(expected, resp.(api.DeleteDataSource204Response)))
}
