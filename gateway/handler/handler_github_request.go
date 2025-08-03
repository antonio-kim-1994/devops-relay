package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"os"
	"time"
)

type addresses struct {
	dev  string
	prod string
}

func GithubRequestHandler(c *gin.Context) {
	var s ServiceInfo
	if err := c.ShouldBindJSON(&s); err != nil {
		log.Error().Err(err).Msgf("failed to bind service info")
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed to get service info",
			"status":  "failed",
		})
		return
	}

	url, err := getTargetServerURL(s.ApplicationName, s.Org, s.Branch)
	if err != nil {
		log.Error().Err(err).Msgf("failed to get target server")
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed to get target server endpoint",
			"error":   fmt.Sprintf("%v", err),
			"status":  "failed",
		})
		return
	}

	log.Info().Msgf("GithubRequestHandler | target url: %s", url)

	err = sendGithubRequestInfo(&s, url)
	if err != nil {
		log.Error().Err(err).Msgf("failed to send service info")
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed to send service info",
			"status":  "failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("%s | Sync success.", s.ApplicationName),
		"status":  "success",
	})

	// Datadog Deploy Histry 저장
	err = sendDeployInfoToDatadog(&s)
	if err != nil {
		log.Error().Err(err).Msgf("failed to send service info")
	}
	return
}

func sendGithubRequestInfo(s *ServiceInfo, url string) error {
	path := "update/github"
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", url, path), bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Request-Auth", os.Getenv("REQUEST_TOKEN"))

	log.Info().
		Str("url", url).
		Interface("request", s).
		Msgf("Sending POST request to %s", url)

	resp, err := client.Do(req)
	if err != nil {
		return errors.New(fmt.Sprintf("sendSlackResponse | failed to send request to %s", url))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New(fmt.Sprintf("sendGithubRequestInfo | failed to read response from %s", url))
	}

	log.Info().Msgf("sendGithubRequestInfo | response: %s", string(body))
	return nil
}
