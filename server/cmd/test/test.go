package main

import (
	"candly/internal/config"
	"candly/internal/http/handler"
	// errCustom "candly/internal/errors"
	httpServer "candly/internal/http"
	logging "candly/internal/logging"

	"github.com/go-redis/redis"

	// com "candly/internal/communication"
	// "candly/internal/market"
	"fmt"
)

func main() {
	c := config.GetConfig()
	log := logging.New(c.Logging)
	// data, err := market.GetLatestCandleData("BTCUSDT", "1m")
	// if err != nil {
	// 	if errors.Is(err, errCustom.ErrorFatal) {
	// 		log.Fatal().Err(err).Msg("Failed to get Market data")
	// 	}
	// 	log.Error().Err(err).Msg("Error")
	// }
	// err := com.SendSMS("tesst", "7356156300")
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
	if err != nil {
		log.Error().Err(err).Msg("Error")

	}

	httpServer.NewServerHTTP(nil, handler.NewHandler())

	fmt.Printf("%+v", data)
}
