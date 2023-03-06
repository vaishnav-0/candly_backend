package handler

import (
	"github.com/gin-gonic/gin"
	// "candly/internal/market"
)



// HealthCheck godoc
// @Summary Get all the pools.
// @Description get details of all pools.
// @Tags pools
// @Accept */*
// @Produce json
// @Success 200 {object} map[string]interface{}
//	@Failure		500	
// @Router /pools/get [get]
func (h *Handlers) GetPools(c *gin.Context) {
	// market.GetLatestCandleData(market.Pools[0].Symbol, "1m")	
}	
