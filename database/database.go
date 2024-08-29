package database

import (
	"github.com/yeboahd24/subscription-stripe/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Init(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the models
	err = db.AutoMigrate(&models.CustomUser{}, &models.Product{}, &models.Subscription{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
