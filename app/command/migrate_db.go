package command

import (
	"stock-tool/storage"

	"gorm.io/gorm"
)

func MigrateDB(db *gorm.DB) error {
	err := db.AutoMigrate(&storage.Brand{})
	if err != nil {
		return err
	}

	err = db.AutoMigrate(&storage.Price{})
	if err != nil {
		return err
	}

	return nil
}
