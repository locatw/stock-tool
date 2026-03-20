package handler

import (
	"context"
	"errors"

	api "stock-tool/api/gen"
	"stock-tool/internal/domain/ingestion"
	"stock-tool/internal/usecase"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

// DataTypeUseCase defines the operations the handler delegates to the usecase layer.
type DataTypeUseCase interface {
	Create(ctx context.Context, req *usecase.CreateDataTypeRequest) (*usecase.DataTypeResponse, error)
	Get(ctx context.Context, id uuid.UUID) (*usecase.DataTypeResponse, error)
	List(ctx context.Context, dataSourceID uuid.UUID) ([]*usecase.DataTypeResponse, error)
	Update(ctx context.Context, req *usecase.UpdateDataTypeRequest) (*usecase.DataTypeResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type DataTypeHandler struct {
	uc DataTypeUseCase
}

func (h *DataTypeHandler) ListDataTypes(
	ctx context.Context,
	request api.ListDataTypesRequestObject,
) (api.ListDataTypesResponseObject, error) {
	if request.Params.DataSourceId == nil {
		return nil, errors.New("dataSourceId query parameter is required")
	}
	list, err := h.uc.List(ctx, *request.Params.DataSourceId)
	if err != nil {
		return nil, err
	}
	return api.ListDataTypes200JSONResponse(lo.Map(list, func(dt *usecase.DataTypeResponse, _ int) api.DataType {
		return toDataTypeResponse(dt)
	})), nil
}

func (h *DataTypeHandler) CreateDataType(
	ctx context.Context,
	request api.CreateDataTypeRequestObject,
) (api.CreateDataTypeResponseObject, error) {
	resp, err := h.uc.Create(ctx, &usecase.CreateDataTypeRequest{
		DataSourceID: request.Body.DataSourceId,
		Name:         request.Body.Name,
		Enabled:      request.Body.Enabled,
		Schedule: usecase.ScheduleInput{
			Type:  string(request.Body.Schedule.Type),
			Times: request.Body.Schedule.Times,
		},
		BackfillEnabled:     request.Body.BackfillEnabled,
		StaleTimeoutMinutes: request.Body.StaleTimeoutMinutes,
		Settings:            request.Body.Settings,
	})
	if err != nil {
		if msg, ok := validationErrorMessage(err); ok {
			return api.CreateDataType422JSONResponse{Error: msg}, nil
		}
		return nil, err
	}
	return api.CreateDataType201JSONResponse(toDataTypeResponse(resp)), nil
}

func (h *DataTypeHandler) GetDataType(
	ctx context.Context, request api.GetDataTypeRequestObject,
) (api.GetDataTypeResponseObject, error) {
	resp, err := h.uc.Get(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return api.GetDataType404JSONResponse{Error: "data type not found"}, nil
	}
	return api.GetDataType200JSONResponse(toDataTypeResponse(resp)), nil
}

func (h *DataTypeHandler) UpdateDataType(
	ctx context.Context, request api.UpdateDataTypeRequestObject,
) (api.UpdateDataTypeResponseObject, error) {
	resp, err := h.uc.Update(ctx, &usecase.UpdateDataTypeRequest{
		ID:      request.Id,
		Name:    request.Body.Name,
		Enabled: request.Body.Enabled,
		Schedule: usecase.ScheduleInput{
			Type:  string(request.Body.Schedule.Type),
			Times: request.Body.Schedule.Times,
		},
		BackfillEnabled:     request.Body.BackfillEnabled,
		StaleTimeoutMinutes: request.Body.StaleTimeoutMinutes,
		Settings:            request.Body.Settings,
	})
	if err != nil {
		if msg, ok := validationErrorMessage(err); ok {
			return api.UpdateDataType422JSONResponse{Error: msg}, nil
		}
		return nil, err
	}
	if resp == nil {
		return api.UpdateDataType404JSONResponse{Error: "data type not found"}, nil
	}
	return api.UpdateDataType200JSONResponse(toDataTypeResponse(resp)), nil
}

func (h *DataTypeHandler) DeleteDataType(
	ctx context.Context, request api.DeleteDataTypeRequestObject,
) (api.DeleteDataTypeResponseObject, error) {
	if err := h.uc.Delete(ctx, request.Id); err != nil {
		return nil, err
	}
	return api.DeleteDataType204Response{}, nil
}

func toDataTypeResponse(r *usecase.DataTypeResponse) api.DataType {
	return api.DataType{
		Id:                  r.ID,
		DataSourceId:        r.DataSourceID,
		Name:                r.Name,
		Enabled:             r.Enabled,
		Schedule:            toScheduleAPI(r.Schedule),
		BackfillEnabled:     r.BackfillEnabled,
		StaleTimeoutMinutes: r.StaleTimeoutMinutes,
		Settings:            r.Settings,
		CreatedAt:           r.CreatedAt,
		UpdatedAt:           r.UpdatedAt,
	}
}

func toScheduleAPI(s ingestion.Schedule) api.Schedule {
	times := lo.Map(s.Times(), func(t ingestion.TimeOfDay, _ int) string { return string(t) })
	return api.Schedule{Type: api.Daily, Times: times}
}
