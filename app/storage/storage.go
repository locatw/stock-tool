package storage

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

	Date               Date   `gorm:"not null;index"`
	Code               string `gorm:"not null;index"`
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

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  bool
	TimeZone *time.Location
}

func Init(config Config) (*gorm.DB, error) {
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

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&Brand{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
