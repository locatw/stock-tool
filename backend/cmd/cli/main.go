package main

import (
	"context"
	"fmt"
	"os"
	"stock-tool/cmd/cli/cmd"
	"stock-tool/database"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

const (
	envFile = "./cmd/cli/.env"
)

type envVars struct {
	DBHost     string `env:"DB_HOST" envDefault:"localhost"`
	DBPort     int    `env:"DB_PORT" envDefault:"5432"`
	DBUser     string `env:"DB_USER"`
	DBPassword string `env:"DB_PASSWORD"`
	DBName     string `env:"DB_NAME"`
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

	ctx := context.Background()
	ctx = context.WithValue(ctx, database.CTXKeyDBConfig, dbConfig)

	command := cmd.NewRootCmd()
	command.SetContext(ctx)

	if err := command.Execute(); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
