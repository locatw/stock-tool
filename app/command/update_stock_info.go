package command

import (
	"os"
	"stock-tool/jquants"
	"stock-tool/storage"

	"gorm.io/gorm"
)

const (
	CHUNK_SIZE = 100
)

func UpdateStockInfo(db *gorm.DB, date string) error {
	client := jquants.NewClient()
	token, err := login(client)
	if err != nil {
		return err
	}

	err = updateBrands(client, token, db)
	if err != nil {
		return err
	}

	err = updatePrices(client, token, db, date)
	if err != nil {
		return err
	}

	return nil
}

func updateBrands(client *jquants.Client, token string, db *gorm.DB) error {
	brands, err := client.ListBrand(token, jquants.ListBrandRequest{})
	if err != nil {
		return err
	}

	records := convertBrands(brands)

	loopCount := len(records) / CHUNK_SIZE
	remainder := len(records) % CHUNK_SIZE
	if remainder != 0 {
		loopCount++
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		for i := 0; i < loopCount; i++ {
			startIndex := i * CHUNK_SIZE
			endIndex := startIndex + CHUNK_SIZE
			if len(records) < endIndex {
				endIndex = len(records)
			}
			chunk := records[startIndex:endIndex]

			err = storage.UpsertToBrands(db, chunk)
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

func updatePrices(client *jquants.Client, token string, db *gorm.DB, date string) error {
	targetDate, err := jquants.NewDateFromString(date)
	if err != nil {
		return err
	}

	req := jquants.NewGetDailyQuoteRequestByDate(targetDate)
	err = db.Transaction(func(tx *gorm.DB) error {
		for {
			resp, err := client.GetDailyQuotes(token, req)
			if err != nil {
				return err
			}

			records := convertPrices(resp)

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

				err = storage.UpsertToPrice(db, chunk)
				if err != nil {
					return err
				}
			}

			if resp.PaginationKey == nil {
				break
			} else {
				req.PaginationKey = resp.PaginationKey
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func login(client *jquants.Client) (string, error) {
	mailAddress := os.Getenv("JQUANTS_MAIL_ADDRESS")
	password := os.Getenv("JQUANTS_PASSWORD")
	authUserResp, err := client.AuthUser(jquants.AuthUserRequest{MailAddress: mailAddress, Password: password})
	if err != nil {
		return "", err
	}

	refreshTokenResp, err := client.RefreshToken(jquants.RefreshTokenRequest{RefreshToken: authUserResp.RefreshToken})
	if err != nil {
		return "", err
	}

	return refreshTokenResp.IDToken, nil
}

func convertBrands(fromValues jquants.ListBrandResponse) []storage.Brand {
	brands := []storage.Brand{}
	for _, from := range fromValues.Brands {
		brand := storage.Brand{
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

func convertPrices(fromValues jquants.GetDailyQuoteResponse) []storage.Price {
	prices := []storage.Price{}
	for _, from := range fromValues.DailyQuotes {
		price := storage.Price{
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

func convertDate(jquantsDate jquants.Date) storage.Date {
	return storage.Date{Year: jquantsDate.Year, Month: jquantsDate.Month, Day: jquantsDate.Day}
}
