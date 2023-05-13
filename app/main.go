package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"stock-tool/jquants"
	"stock-tool/storage"

	"github.com/joho/godotenv"
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
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: jquants-study COMMAND")
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "fetch-brands":
		client := jquants.NewClient()
		token, err := login(client)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fetchBrands(client, token)
	case "load-brands":
		brands, err := jquants.LoadBrands()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		records := convertBrands(brands)
		err = storage.UpsertToBrands(db, records)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func login(client *jquants.Client) (string, error) {
	mailAddress := os.Getenv("JQUANTS_MAIL_ADDRESS")
	password := os.Getenv("JQUANTS_PASSWORD")
	authUserResp, err := client.AuthUser(jquants.AuthUserRequest{MailAddress: mailAddress, Password: password})
	if err != nil {
		return "", err
	}

	refreshTokenResp, err := client.RefreshToken(jquants.RefreshTokenRequest{RefreshToken: authUserResp.RefreshToken})
	if err != nil {
		return "", err
	}

	return refreshTokenResp.IDToken, nil
}

func writeToFile(fileName string, value interface{}) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	return encoder.Encode(value)
}

func fetchBrands(client *jquants.Client, token string) error {
	resp, err := client.ListBrand(token, jquants.ListBrandRequest{})
	if err != nil {
		return err
	}

	err = writeToFile("brands.json", resp)
	if err != nil {
		return err
	}

	return nil
}

func convertBrands(fromValues jquants.ListBrandResponse) []storage.Brand {
	brands := []storage.Brand{}
	for _, from := range fromValues.Brands {
		brand := storage.Brand{
			Date:               convertDate(from.Date),
			Code:               from.Code,
			CompanyName:        from.CompanyName,
			CompanyNameEnglish: from.CompanyNameEnglish,
			Sector17Code:       from.Sector17Code,
			Sector17CodeName:   from.Sector17CodeName,
			Sector33Code:       from.Sector33Code,
			Sector33CodeName:   from.Sector33CodeName,
			ScaleCategory:      from.ScaleCategory,
			MarketCode:         from.MarketCode,
			MarketCodeName:     from.MarketCodeName,
		}

		brands = append(brands, brand)
	}

	return brands
}

func convertDate(jquantsDate jquants.Date) storage.Date {
	return storage.Date{Year: jquantsDate.Year, Month: jquantsDate.Month, Day: jquantsDate.Day}
}
