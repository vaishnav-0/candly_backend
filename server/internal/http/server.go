package http

import (
	"github.com/gin-gonic/gin"
	// swaggerfiles "github.com/swaggo/files"
	// ginSwagger "github.com/swaggo/gin-swagger"

	// _ "github.com/thnkrn/go-gin-clean-arch/cmd/api/docs"
	"candly/internal/http/handler"
	// middleware "github.com/thnkrn/go-gin-clean-arch/pkg/api/middleware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "candly/cmd/server/docs"
)

type ServerHTTP struct {
	engine *gin.Engine
}

type Config struct {
}

func NewServerHTTP(config Config, handlers *handler.Handlers) *ServerHTTP {
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
	{
		api.GET("test", handlers.GetPools)
	}

	// api.GET("users/:id", userHandler.FindByID)
	// api.POST("users", userHandler.Save)
	// api.DELETE("users/:id", userHandler.Delete)

	return &ServerHTTP{engine: engine}
}

func (sh *ServerHTTP) Start() {
	sh.engine.Run(":3000")
}
