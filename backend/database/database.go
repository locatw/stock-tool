package database

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

const (
	SchemaName = "stock"

	CTXKeyDBConfig = "DBConfig"
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

type RawDB struct {
	db     *sql.DB
	config Config
}

func NewRawDB(config Config) *RawDB {
	return &RawDB{db: nil, config: config}
}

func (r *RawDB) Connect() error {
	db, err := sql.Open("postgres", r.DSN())
	if err != nil {
		return err
	}

	r.db = db

	return nil
}

func (r *RawDB) Init() error {
	initialized, err := r.checkInitialized()
	if err != nil {
		return fmt.Errorf("failed to check if database is initialized: %w", err)
	} else if initialized {
		return nil
	}

	if _, err := r.db.Exec(fmt.Sprintf(`CREATE SCHEMA %s`, SchemaName)); err != nil {
		return fmt.Errorf("failed to create schema %s: %w", SchemaName, err)
	}

	return nil
}

func (r *RawDB) checkInitialized() (bool, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = '%s'", SchemaName)

	var count int
	if err := r.db.QueryRow(query).Scan(&count); err != nil {
		return false, err
	}

	return count != 0, nil
}

func (r *RawDB) Shutdown() error {
	return r.db.Close()
}

func (r *RawDB) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		r.config.Host,
		r.config.Port,
		r.config.User,
		r.config.Password,
		r.config.DBName,
		func(value bool) string {
			if value {
				return "enable"
			} else {
				return "disable"
			}
		}(r.config.SSLMode),
	)
}

func (db *RawDB) DB() *sql.DB {
	return db.db
}

type DB interface {
	Transaction(fc func(tx DB) error, opts ...*sql.TxOptions) error
	gorm() *gorm.DB
}

type postgresDB struct {
	gormDB *gorm.DB
}

func (db *postgresDB) Transaction(fc func(tx DB) error, opts ...*sql.TxOptions) error {
	return db.gormDB.Transaction(func(tx *gorm.DB) error {
		return fc(&postgresDB{gormDB: tx})
	}, opts...)
}

func (db *postgresDB) gorm() *gorm.DB {
	return db.gormDB
}

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  bool
}

func Connect(config Config) (DB, error) {
	rawDB := NewRawDB(config)
	gormDB, err := gorm.Open(postgres.Open(rawDB.DSN()), &gorm.Config{})
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
