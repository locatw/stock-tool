package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"stock-tool/internal/api/jquants"
	"strings"

	"github.com/samber/lo"
)

type FetchDataTaskUseCase struct {
	jqClient *jquants.Client
}

func NewFetchDataTaskUseCase(jqClient *jquants.Client) *FetchDataTaskUseCase {
	return &FetchDataTaskUseCase{
		jqClient: jqClient,
	}
}

func (uc *FetchDataTaskUseCase) FetchData(ctx context.Context, req *FetchDataRequest) (*FetchDataResponse, error) {
	switch req.Source {
	case "jquants":
		if err := uc.jqClient.Login(); err != nil {
			return nil, err
		}

		switch req.DataType {
		case "brand":
			resp, err := uc.jqClient.ListBrands(jquants.ListBrandRequest{
				Code: req.Code,
				Date: lo.TernaryF(
					req.StartDate != nil,
					func() *jquants.Date { return lo.ToPtr(jquants.NewDateFromTime(*req.StartDate)) },
					func() *jquants.Date { return nil },
				),
			})
			if err != nil {
				return nil, err
			}

			var brands *jquants.ListBrandResponseBody
			switch body := resp.Body.(type) {
			case jquants.ListBrandResponseBody:
				brands = &body
			case jquants.ErrorResponseBody:
				return nil, errors.New(body.Message)
			}

			content, err := json.MarshalIndent(brands, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("failed to marshal brands: %w", err)
			}

			f, err := os.Create(strings.Replace(req.DestURL, "file://", "", 1))
			if err != nil {
				return nil, fmt.Errorf("failed to create file: %w", err)
			}
			defer f.Close()

			if _, err := f.Write(content); err != nil {
				return nil, fmt.Errorf("failed to write to file: %w", err)
			}
		default:
			return nil, fmt.Errorf("unsupported type: %s.%s", req.Source, req.DataType)
		}
	default:
		return nil, fmt.Errorf("unsupported source: %s", req.Source)
	}

	return &FetchDataResponse{}, nil
}
