package memstore

import (
	"github.com/go-redis/redis/v9"
)

type Config struct{

}

func NewRedisClient(config Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	return client
}
