package middleware

import (
		"net/http"

	"github.com/gin-gonic/gin"
)

func (m *Middlewares) AuthorizeToken() gin.HandlerFunc {
	
	
	return func(c *gin.Context) {
		if len(c.Request.Header["Authorization"]) > 0 {
			token := c.Request.Header["Authorization"][0]
			if claims, err := m.auth.VerifyJWT(token); err == nil {
				c.Set("claims", claims)
				c.Next()
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
				return
			}

		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "token not found"})
			return
		}

	}
}
