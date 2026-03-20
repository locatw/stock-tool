package handler

import (
	"context"
	"testing"
	"time"

	api "stock-tool/api/gen"
	"stock-tool/internal/domain/ingestion"
	"stock-tool/internal/usecase"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type DataTypeUseCaseMock struct {
	mock.Mock
}

func (m *DataTypeUseCaseMock) Create(
	ctx context.Context,
	req *usecase.CreateDataTypeRequest,
) (*usecase.DataTypeResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.DataTypeResponse), args.Error(1)
}

func (m *DataTypeUseCaseMock) Get(
	ctx context.Context,
	id uuid.UUID,
) (*usecase.DataTypeResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.DataTypeResponse), args.Error(1)
}

func (m *DataTypeUseCaseMock) List(
	ctx context.Context,
	dataSourceID uuid.UUID,
) ([]*usecase.DataTypeResponse, error) {
	args := m.Called(ctx, dataSourceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*usecase.DataTypeResponse), args.Error(1)
}

func (m *DataTypeUseCaseMock) Update(
	ctx context.Context,
	req *usecase.UpdateDataTypeRequest,
) (*usecase.DataTypeResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.DataTypeResponse), args.Error(1)
}

func (m *DataTypeUseCaseMock) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

type DataTypeHandlerTestSuite struct {
	suite.Suite
	ucMock  *DataTypeUseCaseMock
	handler *DataTypeHandler
}

func TestDataTypeHandler(t *testing.T) {
	suite.Run(t, new(DataTypeHandlerTestSuite))
}

func (s *DataTypeHandlerTestSuite) SetupTest() {
	s.ucMock = new(DataTypeUseCaseMock)
	s.handler = &DataTypeHandler{uc: s.ucMock}
}

func (s *DataTypeHandlerTestSuite) mustSchedule(sched ingestion.Schedule, err error) ingestion.Schedule {
	s.Require().NoError(err)
	return sched
}

func (s *DataTypeHandlerTestSuite) TestListDataTypes() {
	now := time.Now()
	dsID := uuid.Must(uuid.NewV7())
	dtID := uuid.Must(uuid.NewV7())
	s.ucMock.On("List", mock.Anything, dsID).Return([]*usecase.DataTypeResponse{
		{
			ID: dtID, DataSourceID: dsID, Name: "dt1", Enabled: true,
			Schedule: s.mustSchedule(ingestion.NewDailySchedule([]ingestion.TimeOfDay{"18:00"})),
			Settings: map[string]any{}, CreatedAt: now, UpdatedAt: now,
		},
	}, nil)

	resp, err := s.handler.ListDataTypes(context.Background(), api.ListDataTypesRequestObject{
		Params: api.ListDataTypesParams{DataSourceId: lo.ToPtr(dsID)},
	})

	times := []string{"18:00"}
	expected := api.ListDataTypes200JSONResponse{
		{
			Id: dtID, DataSourceId: dsID, Name: "dt1", Enabled: true,
			Schedule: api.Schedule{Type: api.Daily, Times: &times},
			Settings: map[string]any{}, CreatedAt: now, UpdatedAt: now,
		},
	}
	s.NoError(err)
	s.Require().IsType(api.ListDataTypes200JSONResponse{}, resp)
	s.True(cmp.Equal(expected, resp.(api.ListDataTypes200JSONResponse)), cmp.Diff(expected, resp.(api.ListDataTypes200JSONResponse)))
}

func (s *DataTypeHandlerTestSuite) TestCreateDataType_Success() {
	now := time.Now()
	dsID := uuid.Must(uuid.NewV7())
	dtID := uuid.Must(uuid.NewV7())
	expectedReq := &usecase.CreateDataTypeRequest{
		DataSourceID: dsID, Name: "dt", Enabled: true,
		Schedule: usecase.ScheduleInput{Type: "daily", Times: []string{"18:00"}},
		Settings: map[string]any{},
	}
	sched := s.mustSchedule(ingestion.NewDailySchedule([]ingestion.TimeOfDay{"18:00"}))
	s.ucMock.On("Create", mock.Anything, expectedReq).Return(
		&usecase.DataTypeResponse{
			ID: dtID, DataSourceID: dsID, Name: "dt", Enabled: true,
			Schedule: sched,
			Settings: map[string]any{}, CreatedAt: now, UpdatedAt: now,
		}, nil)

	times := []string{"18:00"}
	body := &api.CreateDataTypeRequest{
		DataSourceId: dsID, Name: "dt", Enabled: true,
		Schedule: api.Schedule{Type: api.Daily, Times: &times},
		Settings: map[string]any{},
	}
	resp, err := s.handler.CreateDataType(context.Background(), api.CreateDataTypeRequestObject{Body: body})

	expectedTimes := []string{"18:00"}
	expected := api.CreateDataType201JSONResponse{
		Id: dtID, DataSourceId: dsID, Name: "dt", Enabled: true,
		Schedule: api.Schedule{Type: api.Daily, Times: &expectedTimes},
		Settings: map[string]any{}, CreatedAt: now, UpdatedAt: now,
	}
	s.NoError(err)
	s.Require().IsType(api.CreateDataType201JSONResponse{}, resp)
	s.True(cmp.Equal(expected, resp.(api.CreateDataType201JSONResponse)), cmp.Diff(expected, resp.(api.CreateDataType201JSONResponse)))
}

func (s *DataTypeHandlerTestSuite) TestCreateDataType_ValidationError() {
	dsID := uuid.Must(uuid.NewV7())
	times := []string{}
	expectedReq := &usecase.CreateDataTypeRequest{
		DataSourceID: dsID, Name: "dt", Enabled: true,
		Schedule: usecase.ScheduleInput{Type: "daily", Times: []string{}},
		Settings: map[string]any{},
	}
	s.ucMock.On("Create", mock.Anything, expectedReq).Return(nil, &usecase.ValidationError{Message: "times must not be empty"})

	body := &api.CreateDataTypeRequest{
		DataSourceId: dsID, Name: "dt", Enabled: true,
		Schedule: api.Schedule{Type: api.Daily, Times: &times},
		Settings: map[string]any{},
	}
	resp, err := s.handler.CreateDataType(context.Background(), api.CreateDataTypeRequestObject{Body: body})

	expected := api.CreateDataType422JSONResponse{Error: "times must not be empty"}
	s.NoError(err)
	s.Require().IsType(api.CreateDataType422JSONResponse{}, resp)
	s.True(cmp.Equal(expected, resp.(api.CreateDataType422JSONResponse)), cmp.Diff(expected, resp.(api.CreateDataType422JSONResponse)))
}

func (s *DataTypeHandlerTestSuite) TestGetDataType_Found() {
	now := time.Now()
	dtID := uuid.Must(uuid.NewV7())
	dsID := uuid.Must(uuid.NewV7())
	s.ucMock.On("Get", mock.Anything, dtID).Return(
		&usecase.DataTypeResponse{
			ID: dtID, DataSourceID: dsID, Name: "dt", Enabled: true,
			Schedule: s.mustSchedule(ingestion.NewDailySchedule([]ingestion.TimeOfDay{"18:00"})),
			Settings: map[string]any{}, CreatedAt: now, UpdatedAt: now,
		}, nil)

	resp, err := s.handler.GetDataType(context.Background(), api.GetDataTypeRequestObject{Id: dtID})

	times := []string{"18:00"}
	expected := api.GetDataType200JSONResponse{
		Id: dtID, DataSourceId: dsID, Name: "dt", Enabled: true,
		Schedule: api.Schedule{Type: api.Daily, Times: &times},
		Settings: map[string]any{}, CreatedAt: now, UpdatedAt: now,
	}
	s.NoError(err)
	s.Require().IsType(api.GetDataType200JSONResponse{}, resp)
	s.True(cmp.Equal(expected, resp.(api.GetDataType200JSONResponse)), cmp.Diff(expected, resp.(api.GetDataType200JSONResponse)))
}

func (s *DataTypeHandlerTestSuite) TestGetDataType_NotFound() {
	notFoundID := uuid.Must(uuid.NewV7())
	s.ucMock.On("Get", mock.Anything, notFoundID).Return(nil, nil)

	resp, err := s.handler.GetDataType(context.Background(), api.GetDataTypeRequestObject{Id: notFoundID})

	expected := api.GetDataType404JSONResponse{Error: "data type not found"}
	s.NoError(err)
	s.Require().IsType(api.GetDataType404JSONResponse{}, resp)
	s.True(cmp.Equal(expected, resp.(api.GetDataType404JSONResponse)), cmp.Diff(expected, resp.(api.GetDataType404JSONResponse)))
}

func (s *DataTypeHandlerTestSuite) TestUpdateDataType_Success() {
	now := time.Now()
	dtID := uuid.Must(uuid.NewV7())
	dsID := uuid.Must(uuid.NewV7())
	expectedReq := &usecase.UpdateDataTypeRequest{
		ID: dtID, Name: "updated", Enabled: false,
		Schedule: usecase.ScheduleInput{Type: "daily", Times: []string{"09:00", "15:00"}},
		Settings: map[string]any{},
	}
	sched := s.mustSchedule(ingestion.NewDailySchedule([]ingestion.TimeOfDay{"09:00", "15:00"}))
	s.ucMock.On("Update", mock.Anything, expectedReq).Return(
		&usecase.DataTypeResponse{
			ID: dtID, DataSourceID: dsID, Name: "updated", Enabled: false,
			Schedule: sched,
			Settings: map[string]any{}, CreatedAt: now, UpdatedAt: now,
		}, nil)

	times := []string{"09:00", "15:00"}
	body := &api.UpdateDataTypeRequest{
		Name: "updated", Enabled: false,
		Schedule: api.Schedule{Type: api.Daily, Times: &times},
		Settings: map[string]any{},
	}
	resp, err := s.handler.UpdateDataType(context.Background(), api.UpdateDataTypeRequestObject{Id: dtID, Body: body})

	expectedTimes := []string{"09:00", "15:00"}
	expected := api.UpdateDataType200JSONResponse{
		Id: dtID, DataSourceId: dsID, Name: "updated", Enabled: false,
		Schedule: api.Schedule{Type: api.Daily, Times: &expectedTimes},
		Settings: map[string]any{}, CreatedAt: now, UpdatedAt: now,
	}
	s.NoError(err)
	s.Require().IsType(api.UpdateDataType200JSONResponse{}, resp)
	s.True(cmp.Equal(expected, resp.(api.UpdateDataType200JSONResponse)), cmp.Diff(expected, resp.(api.UpdateDataType200JSONResponse)))
}

func (s *DataTypeHandlerTestSuite) TestUpdateDataType_NotFound() {
	notFoundID := uuid.Must(uuid.NewV7())
	expectedReq := &usecase.UpdateDataTypeRequest{
		ID: notFoundID, Name: "x", Enabled: true,
		Schedule: usecase.ScheduleInput{Type: "daily", Times: []string{"09:00"}},
		Settings: map[string]any{},
	}
	s.ucMock.On("Update", mock.Anything, expectedReq).Return(nil, nil)

	times := []string{"09:00"}
	body := &api.UpdateDataTypeRequest{
		Name: "x", Enabled: true,
		Schedule: api.Schedule{Type: api.Daily, Times: &times},
		Settings: map[string]any{},
	}
	resp, err := s.handler.UpdateDataType(context.Background(), api.UpdateDataTypeRequestObject{Id: notFoundID, Body: body})

	expected := api.UpdateDataType404JSONResponse{Error: "data type not found"}
	s.NoError(err)
	s.Require().IsType(api.UpdateDataType404JSONResponse{}, resp)
	s.True(cmp.Equal(expected, resp.(api.UpdateDataType404JSONResponse)), cmp.Diff(expected, resp.(api.UpdateDataType404JSONResponse)))
}

func (s *DataTypeHandlerTestSuite) TestDeleteDataType() {
	dtID := uuid.Must(uuid.NewV7())
	s.ucMock.On("Delete", mock.Anything, dtID).Return(nil)

	resp, err := s.handler.DeleteDataType(context.Background(), api.DeleteDataTypeRequestObject{Id: dtID})

	expected := api.DeleteDataType204Response{}
	s.NoError(err)
	s.Require().IsType(api.DeleteDataType204Response{}, resp)
	s.True(cmp.Equal(expected, resp.(api.DeleteDataType204Response)), cmp.Diff(expected, resp.(api.DeleteDataType204Response)))
}
