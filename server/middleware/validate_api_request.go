package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
)

func ValidateApiRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := os.Getenv("REQUEST_TOKEN")

		authHeader := c.Request.Header.Get("Request-Auth")
		if authHeader != token {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "Unauthorized request. Check your request.",
			})
			log.Error().Msgf("Unauthorized request | User-Agent: %s, x-Forwarded-Proto: %s, x-Forwarded-For: %s, x-Forwarded-host: %s",
				c.GetHeader("user-agent"),
				c.GetHeader("X-Forwarded-Proto"),
				c.GetHeader("X-Forwarded-For"),
				c.GetHeader("X-Forwarded-Host"),
			)
			return
		}
		c.Next()
	}
}
