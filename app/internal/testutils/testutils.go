package testutils

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  bool
	TimeZone *time.Location
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

	GormDB *gorm.DB
	config Config

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
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		s.config.Host,
		s.config.Port,
		s.config.User,
		s.config.Password,
		s.config.DBName,
		func(value bool) string {
			if value {
				return "enable"
			} else {
				return "disable"
			}
		}(s.config.SSLMode),
		s.config.TimeZone.String(),
	)

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	s.Require().Nil(err)

	s.GormDB = gormDB

	s.dropAllTables()
}

func (s *DBTestSuite) createAllTables() {
	var tables []interface{}
	if hasModelDBTestSuite, ok := s.dbSuite.(HasModelDBTestSuite); ok {
		tables = hasModelDBTestSuite.TableModels()
	}

	for _, table := range tables {
		err := s.GormDB.AutoMigrate(table)
		s.Require().Nil(err)
	}
}

func (s *DBTestSuite) dropAllTables() {
	tables := s.listTables()

	result := s.GormDB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", strings.Join(tables, ", ")))
	s.Require().Nil(result.Error)
}

func (s *DBTestSuite) listTables() []string {
	rows, err := s.GormDB.Raw("SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname = 'public'").Rows()
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

func Run(t *testing.T, dbTestSuite TestingDBSuite) {
	dbTestSuite.setDBSuite(dbTestSuite)

	suite.Run(t, dbTestSuite)
}
