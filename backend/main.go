package main

import (
	"fmt"
	"os"
	"strconv"

	"stock-tool/command"
	"stock-tool/database"
	"stock-tool/jquants"

	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
)

var (
	db database.DB
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

	db, err = database.Connect(database.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     port,
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   "stock",
		SSLMode:  false,
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

	cmdName := os.Args[1]
	switch cmdName {
	case "update-stock-info":
		if len(os.Args) < 3 {
			showUsage()
			os.Exit(1)
		}

		mailAddress := os.Getenv("JQUANTS_MAIL_ADDRESS")
		password := os.Getenv("JQUANTS_PASSWORD")
		client := jquants.NewClient(mailAddress, password)

		err := command.NewUpdateStockInfoCommand(client, db).Execute(os.Args[2])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown command: %s\n", cmdName)
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
