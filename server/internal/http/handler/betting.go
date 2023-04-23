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

// GetPools
//
//	@Summary		Get pools
//	@Description	get the details of all the pools
//	@ID				get-pools
//	@Tags			pool
//	@Produce		json
//	@Success		200		{object}	[]market.PoolData
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

// GetBets
//
//	@Summary		Get bets
//	@Description	Get the details of bets for a given pool
//	@ID				get-bets
//	@Tags			pool
//	@Produce		json
//
// @Param 			pool_id   path string true "pool ID"
//
//	@Success		200		{object}	betting.BetData  "The json contains statistics with stat: prefix and user bet amounts"
//	@Failure		500
//	@Failure		404		{object} 	helpers.HTTPMessage
//	@Router			/pool/{pool_id} [get]
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

// Bet
//
//	@Summary		Bet
//	@Description	Bet an amount on a pool
//	@ID				bet
//	@Tags			pool
//  @Param	PoolData  body BettingData 		true	"Pool data"
//	@Success		200
//	@Failure		400		{object}  helpers.HTTPMessage
//	@Router			/pool/bet [post]
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
