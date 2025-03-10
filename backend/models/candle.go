package models

import "time"

// Candle struct represebt a single OHLC (Open, High, Low, Close) candle
type Candle struct {
	Symbol    string    `json:"symbol"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Timestamp time.Time `json:"timestamp"`
}

// TempCandle struct represents a temporary candle data that is used to build the final candle
type TempCandle struct {
	Symbol     string
	OpenTime   time.Time
	CloseTime  time.Time
	OpenPrice  float64
	ClosePrice float64
	HighPrice  float64
	LowPrice   float64
	Volume     int64
}

// TradeData struct represents a single trade data
type TradeData struct {
	Close     []string `json:"c"`
	Price     float64  `json:"p"`
	Symbol    string   `json:"s"`
	Timestamp int64    `json:"t"`
	Volume    int64    `json:"v"`
}

// Stucture of the data that comes from FH websocket
type FinhubbWSMessage struct {
	Data []TradeData `json:"data"`
	Type string      `json:"type"` // ping | trade
}

// Converts TempCandle to Candle
func (tc *TempCandle) ToCandle() *Candle {
	return &Candle{
		Symbol:    tc.Symbol,
		Open:      tc.OpenPrice,
		High:      tc.HighPrice,
		Low:       tc.LowPrice,
		Close:     tc.ClosePrice,
		Timestamp: tc.CloseTime,
	}
}

// Data to write to the connected clients
type UpdateType string

const (
	Live   UpdateType = "live"   // live candles, still open
	Closed UpdateType = "closed" // past candles, already closed
)

type BroadcastMessage struct {
	UpdateType UpdateType `json:"updateType"` // live | closed
	Candle     *Candle    `json:"candle"`
}
