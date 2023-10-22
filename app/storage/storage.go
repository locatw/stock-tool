package storage

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type HeadColumns struct {
	ID uint `gorm:"primarykey"`
}

type TailColumns struct {
	CreatedAt time.Time      `gorm:"not null"`
	UpdatedAt time.Time      `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Date struct {
	Year  int
	Month int
	Day   int
}

// @see sql.Scanner
func (d *Date) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal string date value: ", value))
	}

	_, err := fmt.Sscanf(str, "%04d-%02d-%02d", &d.Year, &d.Month, &d.Day)
	if err != nil {
		return err
	}

	return nil
}

// @see sql.Valuer
func (d Date) Value() (driver.Value, error) {
	return d.Format(), nil
}

func (d *Date) Format() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, d.Month, d.Day)
}

type Brand struct {
	HeadColumns

	// TODO: to composite index, Date and Code.
	Date               Date   `gorm:"not null;index"`
	Code               string `gorm:"not null;unique;index"`
	CompanyName        string `gorm:"not null"`
	CompanyNameEnglish string `gorm:"not null"`
	Sector17Code       string `gorm:"not null"`
	Sector17CodeName   string `gorm:"not null"`
	Sector33Code       string `gorm:"not null"`
	Sector33CodeName   string `gorm:"not null"`
	ScaleCategory      string `gorm:"not null"`
	MarketCode         string `gorm:"not null"`
	MarketCodeName     string `gorm:"not null"`

	TailColumns
}

type Price struct {
	HeadColumns

	Date             Date            `gorm:"not null;uniqueIndex:idx_prices_date_code"`
	Code             string          `gorm:"not null;uniqueIndex:idx_prices_date_code"`
	Open             decimal.Decimal `gorm:"type:numeric(11,1);null"`
	High             decimal.Decimal `gorm:"type:numeric(11,1);null"`
	Low              decimal.Decimal `gorm:"type:numeric(11,1);null"`
	Close            decimal.Decimal `gorm:"type:numeric(11,1);null"`
	Volume           decimal.Decimal `gorm:"type:numeric(11,1);null"`
	TurnoverValue    decimal.Decimal `gorm:"type:numeric(20,1);null"`
	AdjustmentFactor decimal.Decimal `gorm:"type:numeric(11,1);null"`
	AdjustmentOpen   decimal.Decimal `gorm:"type:numeric(11,1);null"`
	AdjustmentHigh   decimal.Decimal `gorm:"type:numeric(11,1);null"`
	AdjustmentLow    decimal.Decimal `gorm:"type:numeric(11,1);null"`
	AdjustmentClose  decimal.Decimal `gorm:"type:numeric(11,1);null"`
	AdjustmentVolume decimal.Decimal `gorm:"type:numeric(11,1);null"`

	TailColumns
}

type DB interface {
	Transaction(fc func(tx DB) error, opts ...*sql.TxOptions) error
	gorm() *gorm.DB
}

type postgresDB struct {
	gormDB *gorm.DB
}

func (db *postgresDB) gorm() *gorm.DB {
	return db.gormDB
}

func (db *postgresDB) Transaction(fc func(tx DB) error, opts ...*sql.TxOptions) error {
	return db.gormDB.Transaction(func(tx *gorm.DB) error {
		return fc(&postgresDB{gormDB: tx})
	}, opts...)
}

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  bool
	TimeZone *time.Location
}

func Connect(config Config) (DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
		func(value bool) string {
			if value {
				return "enable"
			} else {
				return "disable"
			}
		}(config.SSLMode),
		config.TimeZone.String(),
	)

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &postgresDB{gormDB: gormDB}, nil
}

func UpsertToBrands(db DB, records []Brand) error {
	schema, err := schema.Parse(&Brand{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	updateColumns := []string{}
	ignoreColumns := []string{"id", "created_at"}
	for _, field := range schema.Fields {
		ignore := false
		for _, ignoreColumn := range ignoreColumns {
			if field.DBName == ignoreColumn {
				ignore = true
				break
			}
		}
		if ignore {
			continue
		}

		updateColumns = append(updateColumns, field.DBName)
	}

	db.gorm().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "code"}},
		DoUpdates: clause.AssignmentColumns(updateColumns),
	}).Create(records)

	return nil
}

func UpsertToPrice(db DB, records []Price) error {
	schema, err := schema.Parse(&Price{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	updateColumns := []string{}
	ignoreColumns := []string{"id", "created_at"}
	for _, field := range schema.Fields {
		ignore := false
		for _, ignoreColumn := range ignoreColumns {
			if field.DBName == ignoreColumn {
				ignore = true
				break
			}
		}
		if ignore {
			continue
		}

		updateColumns = append(updateColumns, field.DBName)
	}

	db.gorm().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "date"}, {Name: "code"}},
		DoUpdates: clause.AssignmentColumns(updateColumns),
	}).Create(records)

	return nil
}
