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
		showUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "update-stock-info":
		if len(os.Args) < 3 {
			showUsage()
			os.Exit(1)
		}

		err := command.UpdateStockInfo(db, os.Args[2])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		fmt.Println("")
		showUsage()
		os.Exit(1)
	}
}

func showUsage() {
	fmt.Println("Usage: stock-tool COMMAND")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  update-stock-info DATE")
	fmt.Println("    Update stock information.")
	fmt.Println("    Args:")
	fmt.Println("      DATE  Update target date. Format is 'YYYY-MM-DD'.")
}
