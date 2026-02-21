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

func CreateGormDB(db *sql.DB) (*gorm.DB, error) {
	return gorm.Open(
		postgres.New(postgres.Config{Conn: db}),
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				TablePrefix: SchemaName + ".", // Use schema name as table prefix
			},
		})
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
	gormDB, err := gorm.Open(postgres.Open(rawDB.DSN()), &gorm.Config{
		// TODO: same in CreateGormDB()
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: SchemaName + ".", // Use schema name as table prefix
		},
	})
	if err != nil {
		return nil, err
	}

	return &postgresDB{gormDB: gormDB}, nil
}
