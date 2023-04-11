package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"candly/internal/http/helpers"
	"candly/pkg/utils"
)

func (m *Middlewares) AuthorizeToken() gin.HandlerFunc {

	return func(c *gin.Context) {
		if len(c.Request.Header["Authorization"]) > 0 {
			token := c.Request.Header["Authorization"][0]
			if claims, err := m.auth.VerifyUserJWT(token); err == nil {

				if utils.Contains(claims.Roles, "new") {
					c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.JSONMessage("registration not complete"))
					return
				}

				c.Set("claims", claims)
				c.Next()
			} else {

				c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.JSONMessage("invalid token"))
				return

			}

		} else {

			c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.JSONMessage("token not found"))
			return

		}

	}
}

func (m *Middlewares) AuthorizeNewUserToken() gin.HandlerFunc {

	return func(c *gin.Context) {
		if len(c.Request.Header["Authorization"]) > 0 {
			token := c.Request.Header["Authorization"][0]
			if claims, err := m.auth.VerifyNewUserJWT(token); err == nil {

				if utils.Contains(claims.Roles, "new") {
					c.Set("claims", claims)
					c.Next()

				} else {
					c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.JSONMessage("registration already completed"))
					return
				}

			} else {

				c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.JSONMessage("invalid token"))
				return

			}

		} else {

			c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.JSONMessage("token not found"))
			return

		}

	}
}
