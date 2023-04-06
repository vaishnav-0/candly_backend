package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/semihalev/gin-stats"

	// _ "github.com/thnkrn/go-gin-clean-arch/cmd/api/docs"
	"candly/internal/config"
	"candly/internal/http/handler"
	"candly/internal/http/middleware"

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

func NewServerHTTP(conf Config, handlers *handler.Handlers, middlwares *middleware.Middlewares) *ServerHTTP {
	if conf.Mode == config.Production {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	engine.Use(gin.Recovery())
	// Use logger from Gin
	engine.Use(gin.Logger())

	engine.Use(stats.RequestStats())

	engine.GET("/stats", func(c *gin.Context) {
		c.JSON(http.StatusOK, stats.Report())
	})

	//swagger
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := engine.Group("/api")

	auth := api.Group("/auth")
	{
		auth.POST("/validate", handlers.VerifyOTP)
		auth.POST("/generateOTP", handlers.GenerateOTP)
	}

	authorized := api.Group("/")

	// Auth middleware
	authorized.Use(middlwares.AuthorizeToken())
	{
		pool := authorized.Group("/pool")
		{
			pool.GET("", handlers.GetPools)
			pool.GET("/:id", handlers.GetBets)
			pool.POST("/bet", handlers.Bet)
		}

	}

	return &ServerHTTP{engine: engine}
}

func (sh *ServerHTTP) Start(port int) {
	sh.engine.Run(":" + fmt.Sprint(port))
}
