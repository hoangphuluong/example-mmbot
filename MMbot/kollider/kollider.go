package kollider

import (
	"encoding/json"
	"sort"
	"strings"
	"time"

	"example/mmbot/trading-bot/constant"
	"example/mmbot/trading-bot/model"
	"example/mmbot/trading-bot/service"
	"example/mmbot/trading-bot/util"
	orderbookCache "example/mmbot/trading-bot/obcache"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/wesovilabs/koazee"
)

type KolliderResp struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

type KolliderBookItems map[string]float64
type KolliderOrderbookResp struct {
	Type string `json:"type"`
	Data struct {
		Symbol     string            `json:"symbol"`
		Bids       KolliderBookItems `json:"bids"`
		Asks       KolliderBookItems `json:"asks"`
		UpdateType string            `json:"update_type"`
	} `json:"data"`
}

func InitKollider() {
	// Connect ws
	c, _, err := websocket.DefaultDialer.Dial("wss://api.kollider.xyz/v1/ws/", nil)
	if err != nil {
		log.Fatal("cannot connect kollider ws", err)
	}
	defer c.Close()
	// Subscribe to pairs
	symbols := koazee.StreamOf(filterExchangePairs(Pairs, constant.Kollider)).Map(func(item model.ExchangePair) string {
		return item.Name
	}).Do().Out().Val().([]string)
	connectString := map[string]interface{}{
		"type":     "subscribe",
		"channels": []string{"orderbook_level2"},
		"symbols":  symbols,
	}
	connectByte, err := json.Marshal(connectString)
	if err != nil {
		log.WithField("exchange", "kollider").WithField("marshal_error", err).Error()
	}
	c.WriteMessage(websocket.TextMessage, connectByte)

	// Listen for messages
	done := make(chan struct{})
	go func() {
		defer close(done)
		pairs := koazee.StreamOf(filterExchangePairs(Pairs, constant.Kollider)).Do().Out().Val().([]model.ExchangePair)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.WithField("read_error", err).Error()
				return
			}

			// Process ws response and extract data
			processKolliderSocketData(message, pairs)
		}
	}()
	for {
		select {
		case <-done:
			log.WithField("exchange", "kollider").WithField("listen", "done").Info()
			return
		}
	}
}

func processKolliderSocketData(message []byte, pairs []model.ExchangePair) {
	var parsedMsg KolliderResp
	err := json.Unmarshal(message, &parsedMsg)

	if err != nil {
		switch parsedMsg.Type {
		case "level2state":
			var orderbookResp KolliderOrderbookResp
			err = json.Unmarshal(message, &orderbookResp)
			if err != nil {
				log.WithField("type", parsedMsg.Type).WithField("unmarshal_error", err).Error()
				return
			}
			// Build a cachable orderbook
			var pair model.ExchangePair
			for _, p := range pairs {
				if p.Name == orderbookResp.Data.Symbol {
					pair = p
					break
				}
			}
			updatedOrderbook := model.Orderbook{}
			cachedOrderbook := GetFromCache(constant.Kollider)
			if orderbookResp.Data.UpdateType == "snapshot" {
				updatedOrderbook = model.Orderbook{
					Symbol:            orderbookResp.Data.Symbol,
					Exchange:          constant.Kollider,
					LastUpdated:       time.Now().UnixMilli(),
					RequireUpdateTime: 120000,
					OutdateTime:       10000,
					Asks:              getBookItemsFromKolliderResp(nil, orderbookResp.Data.Asks, pair, "asks"),
					Bids:              getBookItemsFromKolliderResp(nil, orderbookResp.Data.Bids, pair, "bids"),
				}
			}

			if orderbookResp.Data.UpdateType == "delta" {
				updatedOrderbook = model.Orderbook{
					Symbol:            orderbookResp.Data.Symbol,
					Exchange:          constant.Kollider,
					LastUpdated:       time.Now().UnixMilli(),
					RequireUpdateTime: 120000,
					OutdateTime:       10000,
					Asks:              getBookItemsFromKolliderResp(cachedOrderbook.Data[orderbookResp.Data.Symbol].Asks, orderbookResp.Data.Asks, pair, "asks"),
					Bids:              getBookItemsFromKolliderResp(cachedOrderbook.Data[orderbookResp.Data.Symbol].Bids, orderbookResp.Data.Bids, pair, "bids"),
				}
			}

			// Save orderbook to cache
			cachedOrderbook.Set(orderbookResp.Data.Symbol, &updatedOrderbook)
			SetToCache(constant.Kollider, cachedOrderbook)
		default:
			log.WithField("type", parsedMsg.Type).WithField("unmarshal_error", err).Error()
			return
		}
	} else {
		log.WithField("exchange", "kollider").WithField("recv_message", parsedMsg).Info()
	}
}

func getBookItemsFromKolliderResp(cachedBookItems []model.Book, items KolliderBookItems, pair model.ExchangePair, side string) []model.Book {
	length := 10
	if length > len(items) {
		length = len(items)
	}

	i := 0
	for price, amount := range items {
		if i > length {
			break
		}

		formattedPrice := service.GetFormattedPrice(util.String2Float64(price), pair.PriceDecimals)

		if cachedBookItems == nil {
			var books []model.Book
			books = append(books, model.Book{
				Price:    formattedPrice,
				Quantity: amount,
			})
			cachedBookItems = books
		} else { // Update cached orderbook
			// Find the updated item
			i := sort.Search(len(cachedBookItems), func(i int) bool {
				if side == "asks" {
					return cachedBookItems[i].Price >= formattedPrice
				} else { // Bids
					return cachedBookItems[i].Price <= formattedPrice
				}
			})
			if amount == 0 {
				// Remove from orderbook
				if i < len(cachedBookItems) && cachedBookItems[i].Price == formattedPrice {
					remove(cachedBookItems, i)
				}
			} else {
				// Update amount at price
				if i < len(cachedBookItems) && cachedBookItems[i].Price == formattedPrice {
					cachedBookItems[i] = model.Book{
						Price:    formattedPrice,
						Quantity: amount,
					}
				} else {
					// Insert new
					cachedBookItems = append(cachedBookItems, model.Book{})
					copy(cachedBookItems[i+1:], cachedBookItems[i:])
					cachedBookItems[i] = model.Book{
						Price:    formattedPrice,
						Quantity: amount,
					}
				}
			}
		}
		i++
	}
	return cachedBookItems
}

func SetToCache(key string, value *model.ExchangeOrderBooks) {
	key = strings.ReplaceAll(key, " ", "")
	for _, itemValue := range value.Data {
		Cache.SetValue(orderbookCache.OrderBookPrefix+key, itemValue.Symbol, itemValue)
	}
}

func GetFromCache(key string) *model.ExchangeOrderBooks {
	result := model.ExchangeOrderBooks{
		Data: make(map[string]*model.Orderbook),
	}
	data := make(map[string]interface{})
	key = strings.ReplaceAll(key, " ", "") // remove all space
	Cache.GetValue(orderbookCache.OrderBookPrefix+key, data)
	for symbol, value := range data {
		var item model.Orderbook
		jsonbody, err := json.Marshal(value)
		if err != nil {
			log.WithError(err).WithField("when", "GetOrderbook json Marshal")
		}

		err = json.Unmarshal(jsonbody, &item)
		if err != nil {
			log.WithError(err).WithField("when", "GetOrderbook json Unmarshal")
		}
		result.Data[symbol] = &item
	}

	return &result
}

func remove(slice []model.Book, s int) []model.Book {
	return append(slice[:s], slice[s+1:]...) // preserves the order
}
