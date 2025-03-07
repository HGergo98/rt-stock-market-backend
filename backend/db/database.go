package db

import (
	"fmt"
	"log"

	"github.com/HGergo98/rt-stock-market-backend/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(config *config.EnvConfig, DBMigrator func(db *gorm.DB) error) *gorm.DB {
	uri := fmt.Sprintf("host=%s user=%s dbname=%s password=%s sslmode=%s port=5432",
		config.DBHost, config.DBUser, config.DBName, config.DBPassword, config.DBSSLMode)

	db, err := gorm.Open(postgres.Open(uri), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	log.Println("Connected to database")

	if err := DBMigrator(db); err != nil {
		log.Fatalf("Unable to migrate: %v", err)
	}

	return db
}
