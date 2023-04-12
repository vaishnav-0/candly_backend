package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/semihalev/gin-stats"

	// _ "github.com/thnkrn/go-gin-clean-arch/cmd/api/docs"
	"candly/internal/auth"
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
	Mode          config.Mode
	SwaggerAPIKey string
}

type Dep struct {
	Db   *pgxpool.Pool
	Rd   *redis.Client
	Log  *zerolog.Logger
	Auth *auth.Auth
}

func NewServerHTTP(conf Config, dep *Dep) *ServerHTTP {
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
	engine.GET("/swagger/*any", middleware.AuthorizeAPIKey(conf.SwaggerAPIKey),
		ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := engine.Group("/api")

	auth := api.Group("/auth")
	{
		auth.POST("/validate", handler.VerifyOTP(dep.Auth, dep.Log))
		auth.POST("/generateOTP", handler.GenerateOTP(dep.Auth, dep.Log))
		auth.POST("/register", middleware.AuthorizeNewUserToken(dep.Auth), handler.RegisterUser(dep.Auth, dep.Log))
	}

	authorized := api.Group("/")

	// Auth middleware
	authorized.Use(middleware.AuthorizeToken(dep.Auth))
	{
		pool := authorized.Group("/pool")
		{
			pool.GET("", handler.GetPools(dep.Rd, dep.Log))
			pool.GET("/:id", handler.GetBets(dep.Rd, dep.Log))
			pool.POST("/bet", handler.Bet(dep.Rd, dep.Log))
		}

	}

	return &ServerHTTP{engine: engine}
}

func (sh *ServerHTTP) Start(port int) {
	sh.engine.Run(":" + fmt.Sprint(port))
}
