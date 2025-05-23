package command

import (
	"errors"
	"stock-tool/database"
	"stock-tool/jquants"
)

const (
	CHUNK_SIZE = 100
)

type UpdateStockInfoCommand struct {
	jquantsClient *jquants.Client
	db            database.DB
}

func NewUpdateStockInfoCommand(client *jquants.Client, db database.DB) *UpdateStockInfoCommand {
	return &UpdateStockInfoCommand{
		jquantsClient: client,
		db:            db,
	}
}

func (c *UpdateStockInfoCommand) Execute(date string) error {
	err := c.jquantsClient.Login()
	if err != nil {
		return err
	}

	err = c.updateBrands()
	if err != nil {
		return err
	}

	err = c.updatePrices(date)
	if err != nil {
		return err
	}

	return nil
}

func (c *UpdateStockInfoCommand) updateBrands() error {
	resp, err := c.jquantsClient.ListBrands(jquants.ListBrandRequest{})
	if err != nil {
		return err
	}

	var brands *jquants.ListBrandResponseBody
	switch body := resp.Body.(type) {
	case jquants.ListBrandResponseBody:
		brands = &body
	case jquants.ErrorResponseBody:
		return errors.New(body.Message)
	}

	records := convertBrands(brands)

	loopCount := len(records) / CHUNK_SIZE
	remainder := len(records) % CHUNK_SIZE
	if remainder != 0 {
		loopCount++
	}
	err = c.db.Transaction(func(tx database.DB) error {
		for i := 0; i < loopCount; i++ {
			startIndex := i * CHUNK_SIZE
			endIndex := startIndex + CHUNK_SIZE
			if len(records) < endIndex {
				endIndex = len(records)
			}
			chunk := records[startIndex:endIndex]

			err = database.UpsertToBrands(tx, chunk)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *UpdateStockInfoCommand) updatePrices(date string) error {
	targetDate, err := jquants.NewDateFromString(date)
	if err != nil {
		return err
	}

	req := jquants.NewGetDailyQuoteRequestByDate(targetDate)
	err = c.db.Transaction(func(tx database.DB) error {
		for {
			resp, err := c.jquantsClient.GetDailyQuotes(req)
			if err != nil {
				return err
			}

			var prices *jquants.GetDailyQuoteResponseBody
			switch body := resp.Body.(type) {
			case jquants.GetDailyQuoteResponseBody:
				prices = &body
			case jquants.ErrorResponseBody:
				return errors.New(body.Message)
			}

			records := convertPrices(prices)

			loopCount := len(records) / CHUNK_SIZE
			remainder := len(records) % CHUNK_SIZE
			if remainder != 0 {
				loopCount++
			}
			for i := 0; i < loopCount; i++ {
				startIndex := i * CHUNK_SIZE
				endIndex := startIndex + CHUNK_SIZE
				if len(records) < endIndex {
					endIndex = len(records)
				}
				chunk := records[startIndex:endIndex]

				err = database.UpsertToPrice(tx, chunk)
				if err != nil {
					return err
				}
			}

			if prices.PaginationKey == nil {
				break
			} else {
				req.PaginationKey = prices.PaginationKey
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func convertBrands(fromValues *jquants.ListBrandResponseBody) []database.Brand {
	brands := []database.Brand{}
	for _, from := range fromValues.Brands {
		brand := database.Brand{
			Date:               convertDate(from.Date),
			Code:               from.Code,
			CompanyName:        from.CompanyName,
			CompanyNameEnglish: from.CompanyNameEnglish,
			Sector17Code:       from.Sector17Code,
			Sector17CodeName:   from.Sector17CodeName,
			Sector33Code:       from.Sector33Code,
			Sector33CodeName:   from.Sector33CodeName,
			ScaleCategory:      from.ScaleCategory,
			MarketCode:         from.MarketCode,
			MarketCodeName:     from.MarketCodeName,
		}

		brands = append(brands, brand)
	}

	return brands
}

func convertPrices(fromValues *jquants.GetDailyQuoteResponseBody) []database.Price {
	prices := []database.Price{}
	for _, from := range fromValues.DailyQuotes {
		price := database.Price{
			Date:             convertDate(from.Date),
			Code:             from.Code,
			Open:             from.Open,
			High:             from.High,
			Low:              from.Low,
			Close:            from.Close,
			Volume:           from.Volume,
			TurnoverValue:    from.TurnoverValue,
			AdjustmentFactor: from.AdjustmentFactor,
			AdjustmentOpen:   from.AdjustmentOpen,
			AdjustmentHigh:   from.AdjustmentHigh,
			AdjustmentLow:    from.AdjustmentLow,
			AdjustmentClose:  from.AdjustmentClose,
			AdjustmentVolume: from.AdjustmentVolume,
		}

		prices = append(prices, price)
	}

	return prices
}

func convertDate(jquantsDate jquants.Date) database.Date {
	return database.Date{Year: jquantsDate.Year, Month: jquantsDate.Month, Day: jquantsDate.Day}
}
