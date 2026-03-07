package main

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	oapimiddleware "github.com/oapi-codegen/echo-middleware"
	"github.com/samber/do"

	api "stock-tool/api/gen"
	"stock-tool/cmd/api/handler"
	"stock-tool/database"
	"stock-tool/internal/infra/repository"
	"stock-tool/internal/usecase"
)

const envFile = "./cmd/api/.env"

type envVars struct {
	DBHost     string `env:"DB_HOST" envDefault:"localhost"`
	DBPort     int    `env:"DB_PORT" envDefault:"5432"`
	DBUser     string `env:"DB_USER"`
	DBPassword string `env:"DB_PASSWORD"`
	DBName     string `env:"DB_NAME"`
	Port       string `env:"PORT" envDefault:"8080"`
}

var ev envVars

func init() {
	_, err := os.Stat(envFile)
	if err == nil {
		if err := godotenv.Load(envFile); err != nil {
			fmt.Printf("failed to load .env file: %v\n", err)
			os.Exit(1)
		}
	} else if !os.IsNotExist(err) {
		fmt.Printf("failed to check env file existence: %v\n", err)
		os.Exit(1)
	}

	ev, err = env.ParseAs[envVars]()
	if err != nil {
		fmt.Printf("failed to parse environment variables: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	injector := do.New()

	do.Provide(injector, func(i *do.Injector) (*database.RawDB, error) {
		db := database.NewRawDB(database.Config{
			Host:     ev.DBHost,
			Port:     ev.DBPort,
			User:     ev.DBUser,
			Password: ev.DBPassword,
			DBName:   ev.DBName,
			SSLMode:  false,
		})
		if err := db.Connect(); err != nil {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}
		return db, nil
	})

	do.Provide(injector, func(i *do.Injector) (*repository.DataSourceRepository, error) {
		rawDB := do.MustInvoke[*database.RawDB](i)
		gormDB, err := database.CreateGormDB(rawDB.DB())
		if err != nil {
			return nil, fmt.Errorf("failed to create Gorm DB: %w", err)
		}
		return repository.NewDataSourceRepository(gormDB), nil
	})

	do.Provide(injector, func(i *do.Injector) (*usecase.DataSourceUseCase, error) {
		repo := do.MustInvoke[*repository.DataSourceRepository](i)
		return usecase.NewDataSourceUseCase(repo), nil
	})

	do.Provide(injector, func(i *do.Injector) (*repository.DataTypeRepository, error) {
		rawDB := do.MustInvoke[*database.RawDB](i)
		gormDB, err := database.CreateGormDB(rawDB.DB())
		if err != nil {
			return nil, fmt.Errorf("failed to create Gorm DB: %w", err)
		}
		return repository.NewDataTypeRepository(gormDB), nil
	})

	do.Provide(injector, func(i *do.Injector) (*usecase.DataTypeUseCase, error) {
		repo := do.MustInvoke[*repository.DataTypeRepository](i)
		return usecase.NewDataTypeUseCase(repo), nil
	})

	do.Provide(injector, func(i *do.Injector) (*handler.Handler, error) {
		dsUC := do.MustInvoke[*usecase.DataSourceUseCase](i)
		dtUC := do.MustInvoke[*usecase.DataTypeUseCase](i)
		return handler.NewHandler(dsUC, dtUC), nil
	})

	h := do.MustInvoke[*handler.Handler](injector)

	swagger, err := api.GetSwagger()
	if err != nil {
		fmt.Printf("failed to load OpenAPI spec: %v\n", err)
		os.Exit(1)
	}
	swagger.Servers = nil

	e := echo.New()
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			c.Logger().Infof("uri=%s status=%d", v.URI, v.Status)
			return nil
		},
	}))
	e.Use(middleware.Recover())
	e.Use(oapimiddleware.OapiRequestValidator(swagger))

	api.RegisterHandlers(e, api.NewStrictHandler(h, nil))

	e.Logger.Fatal(e.Start(":" + ev.Port))
}
