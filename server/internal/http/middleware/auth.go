package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"candly/internal/http/helpers"
	"candly/pkg/utils"
	"candly/internal/auth"
)

func AuthorizeToken(a *auth.Auth) gin.HandlerFunc {

	return func(c *gin.Context) {
		if len(c.Request.Header["Authorization"]) > 0 {
			token := c.Request.Header["Authorization"][0]
			if claims, err := a.VerifyUserJWT(token); err == nil {

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

func AuthorizeNewUserToken(a *auth.Auth) gin.HandlerFunc {

	return func(c *gin.Context) {
		if len(c.Request.Header["Authorization"]) > 0 {
			token := c.Request.Header["Authorization"][0]
			if claims, err := a.VerifyNewUserJWT(token); err == nil {

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

func AuthorizeAPIKey(key string) gin.HandlerFunc { 

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization") 
		if len(authHeader) == 0 {                  // Check if it's set
			c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.JSONMessage("Authorization header is required"))
			return
		}
		if authHeader != key { // Check if the API Key is correct
			c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.JSONMessage("unauthorized"))
			return
		}
		c.Next()
	}
}
