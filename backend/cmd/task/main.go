package main

import (
	"context"
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/samber/do"

	"stock-tool/cmd/task/cmd"
	"stock-tool/database"
	"stock-tool/internal/api/jquants"
	"stock-tool/internal/infra/repository"
	"stock-tool/internal/infra/storage"
)

const (
	envFile = "./cmd/task/.env"
)

type envVars struct {
	JQuantsMailAddress string `env:"JQUANTS_MAIL_ADDRESS"`
	JQuantsPassword    string `env:"JQUANTS_PASSWORD"`
	DBHost             string `env:"DB_HOST" envDefault:"localhost"`
	DBPort             int    `env:"DB_PORT" envDefault:"5432"`
	DBUser             string `env:"DB_USER"`
	DBPassword         string `env:"DB_PASSWORD"`
	DBName             string `env:"DB_NAME"`
	S3Endpoint         string `env:"S3_ENDPOINT" envDefault:"http://localhost:8333"`
	S3Bucket           string `env:"S3_BUCKET"`
	S3AccessKey        string `env:"S3_ACCESS_KEY"`
	S3SecretKey        string `env:"S3_SECRET_KEY"`
	S3Region           string `env:"S3_REGION" envDefault:"ap-northeast-1"`
	S3ForcePathStyle   bool   `env:"S3_FORCE_PATH_STYLE" envDefault:"true"`
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
	dbConfig := database.Config{
		Host:     ev.DBHost,
		Port:     ev.DBPort,
		User:     ev.DBUser,
		Password: ev.DBPassword,
		DBName:   ev.DBName,
		SSLMode:  false,
	}
	db := database.NewRawDB(dbConfig)
	if err := db.Connect(); err != nil {
		fmt.Printf("failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, database.CTXKeyDBConfig, dbConfig)

	injector := do.New()
	do.Provide(injector, func(i *do.Injector) (*jquants.Client, error) {
		return jquants.NewClient(ev.JQuantsMailAddress, ev.JQuantsPassword), nil
	})
	do.Provide(injector, func(i *do.Injector) (*jquants.BrandFetcher, error) {
		client := do.MustInvoke[*jquants.Client](i)
		return jquants.NewBrandFetcher(client), nil
	})
	do.Provide(injector, func(i *do.Injector) (*database.RawDB, error) {
		return db, nil
	})
	do.Provide(injector, func(i *do.Injector) (*repository.ExtractTaskRepository, error) {
		rawDB := do.MustInvoke[*database.RawDB](i)
		db, err := rawDB.CreateGormDB()
		if err != nil {
			return nil, fmt.Errorf("failed to create Gorm DB: %w", err)
		}
		return repository.NewExtractTaskRepository(db), nil
	})
	do.Provide(injector, func(i *do.Injector) (*storage.S3Client, error) {
		return storage.NewS3Client(storage.S3Config{
			Endpoint:       ev.S3Endpoint,
			Bucket:         ev.S3Bucket,
			AccessKey:      ev.S3AccessKey,
			SecretKey:      ev.S3SecretKey,
			Region:         ev.S3Region,
			ForcePathStyle: ev.S3ForcePathStyle,
		}), nil
	})

	command := cmd.NewRootCmd(injector)
	command.SetContext(ctx)

	if err := command.Execute(); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
