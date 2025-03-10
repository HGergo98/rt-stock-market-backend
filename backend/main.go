package main

import (
	"fmt"
	"net/http"

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

	// Handle incoming messages
	go utils.HandleFinhubbWSMessages(finnhubWSConn, db)

	// Broadcast messages to all connected clients
	go utils.BroadcastUpdates()

	// Endpoint for connect to the websocket
	http.HandleFunc("/ws", utils.WSHandler)

	// Endpoint for fetching all the past candles for all the symbols
	http.HandleFunc("/stocks-history", func(w http.ResponseWriter, r *http.Request) {
		utils.StocksHistoryHandler(w, r, db)
	})

	// Endpoint for fetching all the past candles for a symbol
	http.HandleFunc("/stocks-candles", func(w http.ResponseWriter, r *http.Request) {
		utils.CandlesHandler(w, r, db)
	})

	// Serve the endpoints
	http.ListenAndServe(fmt.Sprintf(":%s", envConfig.ServerPort), nil)
}
