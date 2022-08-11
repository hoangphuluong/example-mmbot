package model

import (
	"time"
)

type PairStatus string

const (
	EnabledPair  PairStatus = "enabled"
	DisabledPair PairStatus = "disabled"
	RemovedPair  PairStatus = "removed"
)

type ExchangePair struct {
	ID                 int         `json:"id" gorm:"primary_key"`
	BaseAssetSymbolId  string      `json:"baseAssetSymbolId"`
	QuoteAssetSymbolId string      `json:"quoteAssetSymbolId"`
	BaseAssetSymbol    AssetSymbol `json:"baseAssetSymbol" gorm:"foreignKey:BaseAssetSymbolId" validate:"-"`
	QuoteAssetSymbol   AssetSymbol `json:"quoteAssetSymbol" gorm:"foreignKey:QuoteAssetSymbolId" validate:"-"`
	Base               string      `json:"-" gorm:"-" validate:"-"`
	Quote              string      `json:"-" gorm:"-" validate:"-"`
	Name               string      `json:"name"`
	WSName             string      `json:"wsName"`
	ExchangeID         int         `json:"exchangeId"`
	Status             PairStatus  `json:"status"`
	AmountDecimals     int         `json:"amountDecimals"`
	PriceDecimals      int         `json:"priceDecimals"`
	MinNotional        float64     `json:"minNotional"`
	MinAmount          float64     `json:"minAmount"`
	MaxAmount          float64     `json:"maxAmount"`
	ExpiredDate        time.Time   `json:"expiredDate"`
	CreatedAt          time.Time   `json:"createdAt"`
	UpdatedAt          time.Time   `json:"-"`
	Exchange           Exchange    `json:"exchangeInfo" gorm:"foreignkey:ExchangeID" validate:"-"`
}

func (ExchangePair) TableName() string {
	return "exchange_pair"
}
