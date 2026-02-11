package main

import (
	"fmt"
	"os"
	"stock-tool/cmd/api/cmd"
	"stock-tool/database"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

const (
	envFile = "./cmd/api/.env"
)

type envVars struct {
	DBHost  string `env:"DB_HOST" envDefault:"localhost"`
	DBPort  int    `env:"DB_PORT" envDefault:"5432"`
	DBUser  string `env:"DB_USER"`
	DBPass  string `env:"DB_PASSWORD"`
	DBName  string `env:"DB_NAME"`
	APIPort int    `env:"API_PORT" envDefault:"8080"`
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
		Password: ev.DBPass,
		DBName:   ev.DBName,
		SSLMode:  false,
	}
	rawDB := database.NewRawDB(dbConfig)
	if err := rawDB.Connect(); err != nil {
		fmt.Printf("failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer rawDB.Shutdown()

	command := cmd.NewRootCmd(rawDB.DB(), ev.APIPort)

	if err := command.Execute(); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
