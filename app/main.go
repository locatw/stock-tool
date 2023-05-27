package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"stock-tool/command"
	"stock-tool/storage"

	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	db, err = storage.Init(storage.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     port,
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   "stock",
		SSLMode:  false,
		TimeZone: time.FixedZone("Asia/Tokyo", 9*60*60),
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	decimal.MarshalJSONWithoutQuotes = true
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: jquants-study COMMAND [COMMAND_ARGS]")
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "migrate":
		err := command.MigrateDB(db)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "update-stock-info":
		if len(os.Args) < 3 {
			fmt.Println("Usage: jquants-study update-stock-info DATE")
			os.Exit(1)
		}

		err := command.UpdateStockInfo(db, os.Args[2])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		os.Exit(1)
	}
}
