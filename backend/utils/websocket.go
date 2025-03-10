package utils

import (
	"encoding/json"
	"fmt"

	"github.com/HGergo98/rt-stock-market-backend/config"
	"github.com/gorilla/websocket"
)

var (
	symbols = []string{"AAPL", "AMZN"}
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
