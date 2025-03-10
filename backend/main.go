package main

import (
	"github.com/HGergo98/rt-stock-market-backend/config"
	"github.com/HGergo98/rt-stock-market-backend/db"
	"github.com/HGergo98/rt-stock-market-backend/utils"
)

func main() {
	// Environmnet config
	envConfig := config.NewEnvConfig()

	// Database connection
	db := db.InitDB(envConfig, db.DBMigrator)

	// Connect to Finhubb Websockets
	finnhubWSConn := utils.ConnectToFinhubbWS(envConfig)
	defer finnhubWSConn.Close()

	// TODO: Handle incoming messages

	// TODO: Broadcast messages to all connected clients

	// TODO: Endpoint for connect to the websocket
	// TODO: Endpoint for fetching all the past candles for all the symbols
	// TODO: Endpoint for fetching all the past candles for a symbol

	// TODO: Serve the endpoints
}
