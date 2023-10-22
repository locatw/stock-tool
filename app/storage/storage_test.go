package storage

import (
	"fmt"
	"testing"

	"stock-tool/internal/testutils"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestDateScan(t *testing.T) {
	input := "2023-01-02"

	actual := Date{}
	err := actual.Scan(input)

	assert.Nil(t, err)
	assert.Equal(t, Date{Year: 2023, Month: 1, Day: 2}, actual)
}

func TestDateValue(t *testing.T) {
	date := Date{Year: 2023, Month: 1, Day: 2}

	actual, err := date.Value()

	assert.Nil(t, err)
	assert.Equal(t, "2023-01-02", actual)
}

type DBTestSuite struct {
	testutils.DBTestSuite
}

func (s *DBTestSuite) TableModels() []interface{} {
	return []interface{}{
		&Brand{},
		&Price{},
	}
}

func (s *DBTestSuite) AssertPartialEqual(expected any, actual any, diffOpts cmp.Option) bool {
	if cmp.Equal(expected, actual, diffOpts) {
		return true
	}

	diff := cmp.Diff(expected, actual, diffOpts)
	return s.Fail(
		fmt.Sprintf(
			"Not equal: \n"+"expected: %v\n"+"actual  : %v%v",
			expected,
			actual,
			diff,
		),
	)
}

func Test_DBTestSuite(t *testing.T) {
	testutils.Run(t, new(DBTestSuite))
}

func (s *DBTestSuite) Test_UpsertToBrands() {
	brands := []Brand{
		{
			Date:               Date{Year: 2023, Month: 1, Day: 1},
			Code:               "brand1",
			CompanyName:        "株式会社A",
			CompanyNameEnglish: "A Inc.",
			Sector17Code:       "99",
			Sector17CodeName:   "その他",
			Sector33Code:       "9999",
			Sector33CodeName:   "その他",
			ScaleCategory:      "-",
			MarketCode:         "0109",
			MarketCodeName:     "その他",
		},
	}

	err := UpsertToBrands(s.GormDB, brands)

	s.Nil(err)

	var actualBrands []Brand
	result := s.GormDB.Find(&actualBrands)
	s.Nil(result.Error)

	s.Equal(1, len(actualBrands))
	for i, actual := range actualBrands {
		diffOpts := cmpopts.IgnoreFields(Brand{}, "CreatedAt", "UpdatedAt")
		s.AssertPartialEqual(brands[i], actual, diffOpts)
	}
}

func (s *DBTestSuite) Test_UpsertToPrice() {
	prices := []Price{
		{
			Date:             Date{Year: 2023, Month: 1, Day: 1},
			Code:             "brand1",
			Open:             decimal.NewFromFloat(100.0),
			High:             decimal.NewFromFloat(110.0),
			Low:              decimal.NewFromFloat(80.0),
			Close:            decimal.NewFromFloat(90.0),
			Volume:           decimal.NewFromFloat(10000.0),
			TurnoverValue:    decimal.NewFromFloat(1000.0),
			AdjustmentFactor: decimal.NewFromFloat(2.0),
			AdjustmentOpen:   decimal.NewFromFloat(200.0),
			AdjustmentHigh:   decimal.NewFromFloat(220.0),
			AdjustmentLow:    decimal.NewFromFloat(160.0),
			AdjustmentClose:  decimal.NewFromFloat(180.0),
			AdjustmentVolume: decimal.NewFromFloat(20000.0),
		},
	}

	err := UpsertToPrice(s.GormDB, prices)

	s.Nil(err)

	var actualPrices []Price
	result := s.GormDB.Find(&actualPrices)
	s.Nil(result.Error)

	s.Equal(1, len(actualPrices))
	for i, actual := range actualPrices {
		diffOpts := cmpopts.IgnoreFields(Price{}, "CreatedAt", "UpdatedAt")
		s.AssertPartialEqual(prices[i], actual, diffOpts)
	}
}
