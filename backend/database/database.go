package database

import (
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

const SchemaName = "stock"

type ctxKey string

const CTXKeyDBConfig ctxKey = "DBConfig"

type HeadColumns struct {
	ID uint `gorm:"primarykey"`
}

type TailColumns struct {
	CreatedAt time.Time      `gorm:"not null"`
	UpdatedAt time.Time      `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// RawDB is the application's single entry point to the PostgreSQL database.
// It owns the lifecycle of the underlying connection and exposes it to
// infrastructure components that need direct SQL or GORM access.
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

	if err := db.Ping(); err != nil {
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

func (r *RawDB) CreateGormDB() (*gorm.DB, error) {
	return gorm.Open(
		postgres.New(postgres.Config{Conn: r.db}),
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				TablePrefix: SchemaName + ".",
			},
		})
}

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  bool
}
