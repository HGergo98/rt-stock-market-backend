package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

	// Map all connected clients and their symbols
	clientsConn = make(map[*websocket.Conn]string)
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

// Sned candle updates to all connected clients for every 1 second, unless it's closed candle
func BroadcastUpdates() {
	// Set the broadcast interval to 1 second
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var latesUpdate *models.BroadcastMessage

	for {
		select {
		case update := <-broadcast:
			// Watch for new updates from the broadcast channel
			// If the update is a closed candle, send it to all connected clients immediately
			if update.UpdateType == models.Closed {
				// Send the closed candle to all connected clients
				broadcastToClients(update)
			} else {
				// Replace temp updates
				latesUpdate = update
			}
		case <-ticker.C:
			// Broadcast the latest update to all connected clients
			if latesUpdate != nil {
				// Send the closed candle to all connected clients
				broadcastToClients(latesUpdate)
			}
			latesUpdate = nil
		}
	}
}

func broadcastToClients(update *models.BroadcastMessage) {
	jsonUpdate, _ := json.Marshal(update)

	// Send the update to all connected clients subscribed to the symbol
	for clientConn, symbol := range clientsConn {
		// If the client is subscribed to the symbol
		if symbol == update.Candle.Symbol {
			// Send the update to the client
			err := clientConn.WriteMessage(websocket.TextMessage, jsonUpdate)
			if err != nil {
				log.Println("Error writing message to client:", err)
				clientConn.Close()
				delete(clientsConn, clientConn)
			}
		}
	}
}

// --- HANDLERS ---
func WSHandler(w http.ResponseWriter, r *http.Request) {
	// Upgare incoming request to a websocket connection
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to websocket:", err)
	}

	// Close the connection when the function returns & unregister the client when they disconnect
	defer conn.Close()
	defer func() {
		delete(clientsConn, conn)
		log.Println("Client disconnected")
	}()

	// Register the client
	for {
		_, symbol, err := conn.ReadMessage()
		clientsConn[conn] = string(symbol)
		log.Println("Client connected with symbol:", string(symbol))

		if err != nil {
			log.Println("Error reading message from client:", err)
			break
		}
	}
}
