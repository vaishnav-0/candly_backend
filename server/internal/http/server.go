package http

import (
	"fmt"

	"github.com/gin-gonic/gin"
	// swaggerfiles "github.com/swaggo/files"
	// ginSwagger "github.com/swaggo/gin-swagger"

	// _ "github.com/thnkrn/go-gin-clean-arch/cmd/api/docs"
	"candly/internal/config"
	"candly/internal/http/handler"

	// middleware "github.com/thnkrn/go-gin-clean-arch/pkg/api/middleware"
	_ "candly/cmd/server/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type ServerHTTP struct {
	engine *gin.Engine
}

type Config struct {
	Mode config.Mode
}

func NewServerHTTP(conf Config, handlers *handler.Handlers) *ServerHTTP {
	if conf.Mode == config.Production {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	engine.Use(gin.Recovery())
	// Use logger from Gin
	engine.Use(gin.Logger())

	// Swagger docs
	// engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// Request JWT
	// engine.POST("/login", middleware.LoginHandler)

	// Auth middleware

	//swagger
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := engine.Group("/api")

	pool := api.Group("/pool")
	{
		pool.GET("", handlers.GetPools)
		pool.GET("/:id", handlers.GetBets)
		pool.POST("/bet", handlers.Bet)
	}

	return &ServerHTTP{engine: engine}
}

func (sh *ServerHTTP) Start(port int) {
	sh.engine.Run(":" + fmt.Sprint(port))
}
