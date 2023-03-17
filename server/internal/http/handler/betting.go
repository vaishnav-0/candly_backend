package handler

import (
	"net/http"

	"candly/internal/betting"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)


type BettingData struct {
    Id     string 
    User string 
	Amount int64
}
// GetPools godoc
// @Summary Get all the pools.
// @Description get details of all pools.
// @Tags pools
// @Accept */*
// @Produce json
// @Success 200 {object} map[string]interface{}
//
//	@Failure		500
//
// @Router /pools/get [get]
func (h *Handlers) GetPools(c *gin.Context) {
	pools, err := betting.GetPools(h.rd)
	if err != nil {
		h.log.Error().Err(err).Msg("cannot get pools")
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, pools)

}

func (h *Handlers) GetBets(c *gin.Context) {
	bets, err := betting.GetBets(h.rd, c.Param("id"))
	if err != nil {
		h.log.Error().Err(err).Msg("cannot get bets")
		c.Status(http.StatusInternalServerError)
		return
	}
	if len(bets) == 0{
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, bets)

}

func (h *Handlers) Bet(c *gin.Context) {
	var data BettingData
	c.MustBindWith(&data, binding.JSON)
	err := betting.Bet(h.rd, data.Id, data.User, data.Amount)
	if err != nil {
		h.log.Error().Err(err).Msg("cannot bet")
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)

}
