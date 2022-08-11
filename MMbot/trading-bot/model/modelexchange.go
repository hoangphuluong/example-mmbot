package model

import (
	"time"
)

type TradeType string

const (
	Spot    TradeType = "spot"
	Futures TradeType = "futures"
	Margin  TradeType = "margin"
	Options TradeType = "options"
)

type Exchange struct {
	ID            int            `json:"id" gorm:"primary_key"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	TradeType     TradeType      `json:"tradeType"`
	CreatedAt     time.Time      `json:"createdAt"`
	ExchangePairs []ExchangePair `json:"pairs" gorm:"foreignkey:ExchangeID" validate:"-"`
}

func (Exchange) TableName() string {
	return "exchange"
}
