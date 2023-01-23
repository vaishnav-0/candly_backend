package handler

import (
	"github.com/gin-gonic/gin"
	"candly/internal/market"
)


// @Summary    	Show active pools 
// @Description   Get the active pools
// @Tags         pools
// @Produce      json
// @Success      200  {object}  market.
// @Failure      400  {object}  httputil.HTTPError
// @Failure      404  {object}  httputil.HTTPError
// @Failure      500  {object}  httputil.HTTPError
// @Router       /accounts/{id} [get]

func (h *Handlers) GetPools(c *gin.Context) {
	market.GetLatestCandleData(market.Pools[0].Symbol, "1m")	
}	
