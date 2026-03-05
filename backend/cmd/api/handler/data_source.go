package handler

import (
	"context"

	api "stock-tool/api/gen"
	"stock-tool/internal/usecase"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

type DataSourceUseCase interface {
	Create(ctx context.Context, req *usecase.CreateDataSourceRequest) (*usecase.DataSourceResponse, error)
	Get(ctx context.Context, id uuid.UUID) (*usecase.DataSourceResponse, error)
	List(ctx context.Context) ([]*usecase.DataSourceResponse, error)
	Update(ctx context.Context, req *usecase.UpdateDataSourceRequest) (*usecase.DataSourceResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type DataSourceHandler struct {
	uc DataSourceUseCase
}

func (h *DataSourceHandler) ListDataSources(
	ctx context.Context, _ api.ListDataSourcesRequestObject,
) (api.ListDataSourcesResponseObject, error) {
	list, err := h.uc.List(ctx)
	if err != nil {
		return nil, err
	}
	return api.ListDataSources200JSONResponse(lo.Map(list, func(s *usecase.DataSourceResponse, _ int) api.DataSource {
		return toDataSourceResponse(s)
	})), nil
}

func (h *DataSourceHandler) CreateDataSource(
	ctx context.Context, request api.CreateDataSourceRequestObject,
) (api.CreateDataSourceResponseObject, error) {
	resp, err := h.uc.Create(ctx, &usecase.CreateDataSourceRequest{
		Name:     request.Body.Name,
		Enabled:  request.Body.Enabled,
		Timezone: request.Body.Timezone,
		Settings: request.Body.Settings,
	})
	if err != nil {
		if msg, ok := validationErrorMessage(err); ok {
			return api.CreateDataSource422JSONResponse{Error: msg}, nil
		}
		return nil, err
	}
	return api.CreateDataSource201JSONResponse(toDataSourceResponse(resp)), nil
}

func (h *DataSourceHandler) GetDataSource(
	ctx context.Context, request api.GetDataSourceRequestObject,
) (api.GetDataSourceResponseObject, error) {
	resp, err := h.uc.Get(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return api.GetDataSource404JSONResponse{Error: "data source not found"}, nil
	}
	return api.GetDataSource200JSONResponse(toDataSourceResponse(resp)), nil
}

func (h *DataSourceHandler) UpdateDataSource(
	ctx context.Context, request api.UpdateDataSourceRequestObject,
) (api.UpdateDataSourceResponseObject, error) {
	resp, err := h.uc.Update(ctx, &usecase.UpdateDataSourceRequest{
		ID:       request.Id,
		Name:     request.Body.Name,
		Enabled:  request.Body.Enabled,
		Timezone: request.Body.Timezone,
		Settings: request.Body.Settings,
	})
	if err != nil {
		if msg, ok := validationErrorMessage(err); ok {
			return api.UpdateDataSource422JSONResponse{Error: msg}, nil
		}
		return nil, err
	}
	if resp == nil {
		return api.UpdateDataSource404JSONResponse{Error: "data source not found"}, nil
	}
	return api.UpdateDataSource200JSONResponse(toDataSourceResponse(resp)), nil
}

func (h *DataSourceHandler) DeleteDataSource(
	ctx context.Context, request api.DeleteDataSourceRequestObject,
) (api.DeleteDataSourceResponseObject, error) {
	if err := h.uc.Delete(ctx, request.Id); err != nil {
		return nil, err
	}
	return api.DeleteDataSource204Response{}, nil
}

func toDataSourceResponse(r *usecase.DataSourceResponse) api.DataSource {
	return api.DataSource{
		Id:        r.ID,
		Name:      r.Name,
		Enabled:   r.Enabled,
		Timezone:  r.Timezone,
		Settings:  r.Settings,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
