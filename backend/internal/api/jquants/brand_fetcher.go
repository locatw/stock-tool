package jquants

import (
	"context"
	"fmt"
	"time"
)

type BrandFetcher struct {
	client *Client
}

func NewBrandFetcher(client *Client) *BrandFetcher {
	return &BrandFetcher{client: client}
}

func (f *BrandFetcher) FetchBrands(ctx context.Context, code *string, date *time.Time) ([]byte, error) {
	if !f.client.IsAuthorized() {
		if err := f.client.Login(); err != nil {
			return nil, fmt.Errorf("failed to login: %w", err)
		}
	}

	var jqDate *Date
	if date != nil {
		d := NewDateFromTime(*date)
		jqDate = &d
	}

	resp, err := f.client.ListBrands(ListBrandRequest{
		Code: code,
		Date: jqDate,
	})
	if err != nil {
		return nil, err
	}

	return resp.RawBody, nil
}
