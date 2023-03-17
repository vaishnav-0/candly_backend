package memstore

import (
	"github.com/go-redis/redis/v9"
	"golang.org/x/net/context"
)

type Config struct {
	Host string `env:"REDIS_HOST,notEmpty"`
	Port string `env:"REDIS_PORT,notEmpty"`
}

func NewRedisClient(config Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Host + ":" + config.Port,
		Password: "",
		DB:       0,
	})
	_, err := client.Ping(context.Background()).Result()
	return client, err
}

func GetHash(store *redis.Client, id string) (map[string]string, error) {

	ctx := context.Background()
	res, err := store.HGetAll(ctx, id).Result()

	if err != nil {
		return nil, err
	}
	return res, nil
}
