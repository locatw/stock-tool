package testutil

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/suite"
)

const (
	testMigrationDir = "../../../migrations"
)

type DBTest struct {
	suite.Suite
	pool     *dockertest.Pool
	resource *dockertest.Resource
	db       *sql.DB
}

func (s *DBTest) setupDockerTest() error {
	testDBName := "testdb"
	testDBUser := "testuser"
	testDBPassword := "testpass"

	pool, err := dockertest.NewPool("")
	if err != nil {
		return fmt.Errorf("could not construct pool: %w", err)
	}

	s.pool = pool

	err = s.pool.Client.Ping()
	if err != nil {
		return fmt.Errorf("could not connect to Docker: %w", err)
	}

	resource, err := s.pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: "postgres",
			Tag:        "17",
			Env: []string{
				fmt.Sprintf("POSTGRES_DB=%s", testDBName),
				fmt.Sprintf("POSTGRES_USER=%s", testDBUser),
				fmt.Sprintf("POSTGRES_PASSWORD=%s", testDBPassword),
				"listen_address = '*'",
			},
		},
		func(config *docker.HostConfig) {
			config.AutoRemove = true
			config.RestartPolicy = docker.RestartPolicy{
				Name: "no",
			}
		},
	)
	if err != nil {
		return fmt.Errorf("could not start resource: %w", err)
	}
	s.resource = resource

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		testDBUser,
		testDBPassword,
		s.resource.GetHostPort("5432/tcp"),
		testDBName,
	)
	pool.MaxWait = 30 * time.Second
	err = s.pool.Retry(func() error {
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			return err
		}

		s.db = db

		return s.db.Ping()
	})
	if err != nil {
		return fmt.Errorf("could not connect to database: %w", err)
	}

	return nil
}

func (s *DBTest) SetupSuite() {
	if err := s.setupDockerTest(); err != nil {
		s.T().Fatal(err)
	}
}

func (s *DBTest) TearDownSuite() {
	if s.resource != nil {
		if err := s.pool.Purge(s.resource); err != nil {
			s.T().Errorf("Could not purge resource: %v", err)
		}
	}
}

func (s *DBTest) ApplyMigrations() {
	db := s.GetDB()

	db.Exec("CREATE SCHEMA IF NOT EXISTS stock")

	driver, err := postgres.WithInstance(db, &postgres.Config{
		SchemaName: "stock",
	})
	s.Require().NoError(err)

	mig, err := migrate.NewWithDatabaseInstance("file://"+testMigrationDir, "postgres", driver)
	s.Require().NoError(err)

	s.Require().NoError(mig.Up())
}

func (s *DBTest) CleanupMigrations() error {
	db := s.GetDB()

	driver, err := postgres.WithInstance(db, &postgres.Config{
		SchemaName: "stock",
	})
	s.Require().NoError(err)

	mig, err := migrate.NewWithDatabaseInstance("file://"+testMigrationDir, "postgres", driver)
	s.Require().NoError(err)

	s.Require().NoError(mig.Down())

	db.Exec("DROP SCHEMA IF EXISTS stock")

	return nil
}

func (s *DBTest) GetDB() *sql.DB {
	return s.db
}
