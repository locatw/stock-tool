package handler

import (
	"errors"

	api "stock-tool/api/gen"
	"stock-tool/internal/usecase"
)

var _ api.StrictServerInterface = (*Handler)(nil)

type Handler struct {
	DataSourceHandler
}

func NewHandler(dsUC DataSourceUseCase) *Handler {
	return &Handler{
		DataSourceHandler: DataSourceHandler{uc: dsUC},
	}
}

func validationErrorMessage(err error) (string, bool) {
	var ve *usecase.ValidationError
	if errors.As(err, &ve) {
		return ve.Message, true
	}
	return "", false
}
