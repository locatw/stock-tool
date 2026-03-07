package handler

import (
	"errors"

	api "stock-tool/api/gen"
	"stock-tool/internal/usecase"
)

var _ api.StrictServerInterface = (*Handler)(nil)

type Handler struct {
	DataSourceHandler
	DataTypeHandler
}

func NewHandler(dsUC DataSourceUseCase, dtUC DataTypeUseCase) *Handler {
	return &Handler{
		DataSourceHandler: DataSourceHandler{uc: dsUC},
		DataTypeHandler:   DataTypeHandler{uc: dtUC},
	}
}

func validationErrorMessage(err error) (string, bool) {
	var ve *usecase.ValidationError
	if errors.As(err, &ve) {
		return ve.Message, true
	}
	return "", false
}
