package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"os"
	"time"
)

func GatewayHealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}

func ServerHealthCheck(c *gin.Context) {
	var h HealthCheckRequest
	if err := c.ShouldBindJSON(&h); err != nil {
		log.Error().Err(err).Msgf("ServerHealthCheck | failed to bind service info")
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed to get service info",
			"status":  "failed",
		})
		return
	}

	url, err := getTargetServerURL(h.ApplicationName, h.Org, h.Branch)
	if err != nil {
		log.Error().Err(err).Msgf("ServerHealthCheck | failed to get target server")
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed to get target server endpoint",
			"status":  "failed",
		})
		return
	}

	log.Info().Msgf("ServerHealthCheck | target url: %s", url)

	body, status, err := sendHealthcheckRequest(h, url)
	if err != nil {
		log.Error().Err(err).Msgf("ServerHealthCheck | failed to send healthcheck request")
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed to send healthcheck request",
			"status":  "failed",
		})
		return
	}

	type result struct {
		Message string `json:"message"`
	}
	var r result

	if err := json.Unmarshal(body, &r); err != nil {
		log.Error().Err(err).Msgf("ServerHealthCheck | failed to unmarshal body")
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed to unmarshal body",
			"status":  "failed",
		})
		return
	}

	c.JSON(status, gin.H{
		"result": r.Message,
		"status": "success",
	})
	return
}

func sendHealthcheckRequest(h HealthCheckRequest, url string) ([]byte, int, error) {
	path := "sys/healthcheck"
	data, err := json.Marshal(h)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", url, path), bytes.NewBuffer(data))
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Request-Auth", os.Getenv("REQUEST_TOKEN"))

	log.Info().
		Str("url", url).
		Interface("request", h).
		Msgf("Sending POST request to %s", url)

	resp, err := client.Do(req)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msgf("sendHealthcheckRequest | failed to read response body")
		return nil, http.StatusInternalServerError, err
	}

	log.Info().Msgf("sendHealthcheckRequest | Successfully sent request to %s", url)
	return body, resp.StatusCode, nil
}
