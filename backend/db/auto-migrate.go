package db

import (
	"github.com/HGergo98/rt-stock-market-backend/models"
	"gorm.io/gorm"
)

func DBMigrator(db *gorm.DB) error {
	return db.AutoMigrate(&models.Candle{})
}
