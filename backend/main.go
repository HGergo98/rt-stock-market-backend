package main

import (
	"github.com/HGergo98/rt-stock-market-backend/config"
	"github.com/HGergo98/rt-stock-market-backend/db"
)

func main() {
	// Environmnet config
	envConfig := config.NewEnvConfig()

	// Database connection
	db := db.InitDB(envConfig, db.DBMigrator)
}
