package model

import "time"

type ExchangeOrderBooks struct {
	Data map[string]*Orderbook
}

func (ob *ExchangeOrderBooks) Set(key string, value *Orderbook) {
	ob.Data[key] = value
}

func (ob *ExchangeOrderBooks) Get(key string) *Orderbook {
	return ob.Data[key]
}

type Orderbook struct {
	Symbol            string `json:"symbol"`
	Base              string `json:"base"`
	Quote             string `json:"quote"`
	Exchange          string `json:"exchange"`
	LastUpdated       int64  `json:"last_updated"`
	RequireUpdateTime int64  `json:"require_update_time"`
	OutdateTime       int64  `json:"outdate_time"`
	Bids              []Book `json:"bids"`
	Asks              []Book `json:"asks"`
}

func (ob *Orderbook) IsOutdated() bool {
	return ob.LastUpdated > 0 && ob.LastUpdated+ob.OutdateTime < time.Now().UnixMilli()
}

func (ob *Orderbook) IsRequireResubscribe() bool {
	return ob.LastUpdated > 0 && ob.RequireUpdateTime > 0 && ob.LastUpdated+ob.RequireUpdateTime < time.Now().UnixMilli()
}

func (ob *Orderbook) IsRequireReconnect() bool {
	return ob.LastUpdated > 0 && ob.RequireUpdateTime > 0 && ob.LastUpdated+(ob.RequireUpdateTime*2) < time.Now().UnixMilli()
}

type Book struct {
	Price    float64
	Quantity float64
}

func NewOrderbook() *Orderbook {
	return &Orderbook{
		LastUpdated: 0,
		Bids:        []Book{},
		Asks:        []Book{},
	}
}

func removePriceLevel(ob []Book, idx int) []Book {
	return append(ob[:idx], ob[idx+1:]...)
}

func insertPriceLevel(ob []Book, idx int, pl Book) []Book {
	if len(ob) == idx {
		ob = append(ob, pl)
	}
	ob = append(ob[:idx+1], ob[idx:]...)
	ob[idx] = pl
	return ob
}
