package orderbookCache

import (
	"encoding/json"

	"github.com/go-redis/redis/v7"
	"github.com/sirupsen/logrus"
)

const (
	OrderBookPrefix = "order_book_"
)

var initialized = false

type Cache struct {
	client *redis.Client
	log    *logrus.Entry
}

func NewCache(redisUri string, log *logrus.Entry) *Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     redisUri,
		Password: "", // not set password
		DB:       0,  // use default DB
	})

	_, err := client.Ping().Result()

	if err != nil {
		log.Fatal(err)
	}

	return &Cache{client: client, log: log}
}

func (c *Cache) SetValue(cacheKey, field string, model interface{}) error {
	serialized, err := json.Marshal(model)

	if err != nil {
		return err
	}
	return c.client.HSet(cacheKey, field, string(serialized)).Err()
}

func (c *Cache) GetValue(cacheKey string, result map[string]interface{}) error {
	serialized, err := c.client.HGetAll(cacheKey).Result()

	if err != nil {
		return err
	}

	for key, value := range serialized {
		var item interface{}
		err = json.Unmarshal([]byte(value), &item)
		result[key] = &item
	}

	return err
}

func (c *Cache) GetValueWithSymbol(cacheKey, symbol string, result interface{}) error {
	serialized, err := c.client.HGet(cacheKey, symbol).Result()

	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(serialized), &result)

	return err
}

func (c *Cache) DelKey(cacheKey string) error {
	return c.client.Del(cacheKey).Err()
}

func (c *Cache) Init() error {
	if !initialized {
		initialized = true
	}

	return nil
}
