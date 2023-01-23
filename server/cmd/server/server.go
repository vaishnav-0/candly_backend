// @title           Candly
// @version         1.0
// @description     Candly server API.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io


// @host      localhost:3000
// @BasePath  /api


package main

import (

	// "net/http"
	"candly/internal/http"
	"candly/internal/http/handler"
	logging "candly/internal/logging"
	"context"
	"fmt"
	"time"

	// "github.com/gin-gonic/gin"
	// comm "candly/internal/communication"
	"candly/internal/config"
	"candly/internal/db"
	"candly/internal/memstore"
	"candly/internal/market"
 
)

func repeatThis(i int) func() {
	return func() {
		fmt.Println(i)
		time.AfterFunc(time.Second, repeatThis(i+1))
	}

}



func main() {
	ctx := context.Background()
	c := config.GetConfig()
	log := logging.New(c.Logging)

	// err := comm.SendSMS("test", "7356156300")
	// if err != nil {
	// 	log.Error().Err(err).Msg("Error sending sms")

	// }
	dbClient, err := db.Open(c.Db.Host, c.Db.Username, c.Db.Password, c.Db.Name)
	if err != nil {
		log.Fatal().Err(err).Msg("Database connection error")
	}
	defer dbClient.Close(ctx)

	rd := memstore.NewRedisClient(memstore.Config{})
	
	defer rd.Close()

	market.StartFetchAndStore(rd, log)

	serverHTTP := http.NewServerHTTP(http.Config{}, handler.NewHandler(dbClient, rd))
	serverHTTP.Start()

	// router.Run("0.0.0.0:8080")
}
