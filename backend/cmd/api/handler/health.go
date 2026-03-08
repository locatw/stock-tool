package handler

import (
	"context"

	api "stock-tool/api/gen"
)

func (h *Handler) HealthCheck(
	_ context.Context,
	_ api.HealthCheckRequestObject,
) (api.HealthCheckResponseObject, error) {
	return api.HealthCheck200JSONResponse{Status: "ok"}, nil
}
