package utils

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/HGergo98/rt-stock-market-backend/config"
	"github.com/HGergo98/rt-stock-market-backend/models"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var (
	symbols = []string{"AAPL", "AMZN"}

	// Map of live candles for each symbol
	tempCandles = make(map[string]*models.TempCandle)

	mu sync.Mutex

	// Broadcast message to all connected clients
	broadcast = make(chan *models.BroadcastMessage)
)

func ConnectToFinhubbWS(envConfig *config.EnvConfig) *websocket.Conn {
	ws, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("wss://ws.finnhub.io?token=%s", envConfig.ApiKey), nil)
	if err != nil {
		panic(err)
	}

	for _, s := range symbols {
		msg, _ := json.Marshal(map[string]interface{}{"type": "subscribe", "symbol": s})
		ws.WriteMessage(websocket.TextMessage, msg)
	}

	return ws
}

func HandleFinhubbWSMessages(ws *websocket.Conn, db *gorm.DB) {
	finnhubMessage := &models.FinhubbWSMessage{}

	for {
		if err := ws.ReadJSON(finnhubMessage); err != nil {
			fmt.Println("Error reading message from Finhubb websocket:", err)
			continue
		}

		// Only handle trade messages
		if finnhubMessage.Type != "trade" {
			continue
		}

		for _, tradeData := range finnhubMessage.Data {
			// Process the trade data
			processTradeData(&tradeData, db)
		}
	}
}

// Process each trade data and update or create a new tempCandle
func processTradeData(tradeData *models.TradeData, db *gorm.DB) {
	// HandleFinhubbWSMessages is a goroutine, so we need to protect from data races
	mu.Lock()
	defer mu.Unlock()

	// Extract trade data
	symbol := tradeData.Symbol
	price := tradeData.Price
	timestamp := time.UnixMilli(tradeData.Timestamp)
	volume := tradeData.Volume

	// Create a new tempCandle for the symbol
	tempCandle, exists := tempCandles[symbol]

	// If the tempCandle doesn't exist, create a new one, or should be already closed
	if !exists || timestamp.After(tempCandle.CloseTime) {
		// Finalize and save the previous tempCandle
		if exists {
			// Convert the tempCandle to a candle
			candle := tempCandle.ToCandle()

			// Save the candle to the database
			if err := db.Create(candle).Error; err != nil {
				fmt.Println("Error saving candle to the database:", err)
			}

			// Broadcast
			broadcast <- &models.BroadcastMessage{
				UpdateType: models.Closed,
				Candle:     candle,
			}
		}

		// Initialize a new tempCandle when not exists
		tempCandle = &models.TempCandle{
			Symbol:     symbol,
			OpenTime:   timestamp,
			CloseTime:  timestamp.Add(time.Minute),
			OpenPrice:  price,
			ClosePrice: price,
			HighPrice:  price,
			LowPrice:   price,
			Volume:     volume,
		}
	}

	// Update the tempCandle
	tempCandle.ClosePrice = price
	tempCandle.Volume += volume
	if price < tempCandle.HighPrice {
		tempCandle.HighPrice = price
	}
	if price > tempCandle.LowPrice {
		tempCandle.LowPrice = price
	}

	// Update the tempCandles symbol
	tempCandles[symbol] = tempCandle

	// Write to the broadcast channel
	broadcast <- &models.BroadcastMessage{
		UpdateType: models.Live,
		Candle:     tempCandle.ToCandle(),
	}
}
