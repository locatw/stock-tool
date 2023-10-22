package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
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

type HasModelDBTestSuite interface {
	TableModels() []interface{}
}

type TestingDBSuite interface {
	suite.TestingSuite

	setDBSuite(dbSuite TestingDBSuite)
}

type DBTestSuite struct {
	suite.Suite

	config Config
	db     DB

	dbSuite TestingDBSuite
}

func (s *DBTestSuite) SetupSuite() {
	s.config = s.loadDBConfig()
	s.setupDB()
}

func (s *DBTestSuite) SetupTest() {
	s.createAllTables()
}

func (s *DBTestSuite) TearDownTest() {
	s.dropAllTables()
}

func (s *DBTestSuite) loadDBConfig() Config {
	curDir, err := os.Getwd()
	s.Require().Nil(err)

	dotEnvPath := filepath.Join(curDir, "..", ".env")

	err = godotenv.Load(dotEnvPath)
	s.Require().Nil(err)

	host := os.Getenv("TEST_DB_HOST")
	s.Require().NotEmpty(host)

	port, err := strconv.Atoi(os.Getenv("TEST_DB_PORT"))
	s.Require().Nil(err)

	user := os.Getenv("TEST_DB_USER")
	s.Require().NotEmpty(user)

	password := os.Getenv("TEST_DB_PASSWORD")
	s.Require().NotEmpty(password)

	return Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   "stock-test",
		SSLMode:  false,
		TimeZone: time.FixedZone("Asia/Tokyo", 9*60*60),
	}
}

func (s *DBTestSuite) setupDB() {
	db, err := Connect(s.config)
	s.Require().Nil(err)

	s.db = db

	s.dropAllTables()
}

func (s *DBTestSuite) createAllTables() {
	var tables []interface{}
	if hasModelDBTestSuite, ok := s.dbSuite.(HasModelDBTestSuite); ok {
		tables = hasModelDBTestSuite.TableModels()
	}

	for _, table := range tables {
		err := s.db.(*postgresDB).gormDB.AutoMigrate(table)
		s.Require().Nil(err)
	}
}

func (s *DBTestSuite) dropAllTables() {
	tables := s.listTables()

	result := s.db.(*postgresDB).gormDB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", strings.Join(tables, ", ")))
	s.Require().Nil(result.Error)
}

func (s *DBTestSuite) listTables() []string {
	rows, err := s.db.(*postgresDB).gormDB.Raw("SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname = 'public'").Rows()
	s.Require().Nil(err)
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			s.Require().Nil(err)
		}

		tables = append(tables, name)
	}

	return tables
}

func (s *DBTestSuite) setDBSuite(dbSuite TestingDBSuite) {
	s.dbSuite = dbSuite
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

func Run(t *testing.T, dbTestSuite TestingDBSuite) {
	dbTestSuite.setDBSuite(dbTestSuite)

	suite.Run(t, dbTestSuite)
}

func Test_DBTestSuite(t *testing.T) {
	Run(t, new(DBTestSuite))
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

	err := UpsertToBrands(s.db, brands)

	s.Nil(err)

	var actualBrands []Brand
	result := s.db.(*postgresDB).gormDB.Find(&actualBrands)
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

	err := UpsertToPrice(s.db, prices)

	s.Nil(err)

	var actualPrices []Price
	result := s.db.(*postgresDB).gormDB.Find(&actualPrices)
	s.Nil(result.Error)

	s.Equal(1, len(actualPrices))
	for i, actual := range actualPrices {
		diffOpts := cmpopts.IgnoreFields(Price{}, "CreatedAt", "UpdatedAt")
		s.AssertPartialEqual(prices[i], actual, diffOpts)
	}
}
