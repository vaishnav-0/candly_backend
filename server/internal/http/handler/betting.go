package handler

import (
	"candly/internal/auth"
	"candly/internal/betting"
	"candly/internal/http/helpers"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-redis/redis/v9"
	"github.com/rs/zerolog"

	_ "github.com/swaggo/swag/example/basic/web"
)

type BettingData struct {
	Id     string
	User   string
	Amount int64
}

// GetStringByInt example
//
//	@Summary		Get pools
//	@Description	get the details of all the pools
//	@ID				get-pools
//	@Produce		json	
//	@Success		200		{object}	market.PoolData
//	@Failure		400
//	@Router			/pool [get]
func GetPools(rd *redis.Client, log *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		pools, err := betting.GetPools(rd)

		if err != nil {
			log.Error().Err(err).Msg("cannot get pools")
			c.Status(http.StatusInternalServerError)
			return
		}

		c.JSON(http.StatusOK, pools)

	}
}

func GetBets(rd *redis.Client, log *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		bets, err := betting.GetBets(rd, c.Param("id"))
		if err != nil {
			log.Error().Err(err).Msg("cannot get bets")
			c.Status(http.StatusInternalServerError)
			return
		}
		if len(bets) == 0 {
			c.Status(http.StatusNotFound)
			return
		}

		c.JSON(http.StatusOK, bets)

	}
}

func Bet(rd *redis.Client, log *zerolog.Logger) gin.HandlerFunc {

	return func(c *gin.Context) {
		var data BettingData
		c.MustBindWith(&data, binding.JSON)

		cl, _ := c.Get("claims")
		claims := cl.(*auth.JwtUserClaims)

		err := betting.Bet(rd, data.Id, claims.User, data.Amount)
		if err == betting.PoolNotFoundError {
			c.JSON(http.StatusBadRequest, helpers.JSONMessage("pool not found"))
			return
		}
		if err != nil {
			log.Error().Err(err).Msg("cannot bet")
			c.Status(http.StatusInternalServerError)
			return
		}

		c.Status(http.StatusOK)

	}
}
