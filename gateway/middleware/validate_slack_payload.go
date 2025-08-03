package middleware

import (
	"bytes"
	"crypto/hmac"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"os"
)

func ValidationCheckSlackPayload() gin.HandlerFunc {
	return func(c *gin.Context) {
		signingSecret := os.Getenv("SLACK_BOT_SIGNING_SECRET")
		if signingSecret == "" {
			log.Error().Msg("SLACK_BOT_SIGNING_SECRET is not set in the environment variables.")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		timestamp := c.GetHeader("X-Slack-Request-Timestamp")
		signature := c.GetHeader("X-Slack-Signature")

		if timestamp == "" || signature == "" {
			log.Warn().Msg("Missing required Slack headers.")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required headers"})
			return
		}

		reqBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Err(err).Msg("Failed to read request body.")
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		baseString := "v0:" + timestamp + ":" + string(reqBody)
		signingHash := "v0=" + generateHmacHash(signingSecret, baseString)

		if !hmac.Equal([]byte(signingHash), []byte(signature)) {
			// Debugging Line
			log.Error().Msgf("[Invalid Header] Signature mismatch: provided=%s, expected=%s", signature, signingHash)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		c.Set("body", string(reqBody))

		// Debugging Line
		log.Debug().Msgf("[Valid Header] Request signature verified: %s", signature)
		c.Next()
	}
}
