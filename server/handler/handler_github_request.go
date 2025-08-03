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
)

const argoUrl = "http://argocd-server.argocd.svc.cluster.local"

func HandleGithubRequest(c *gin.Context) {
	var s ServiceInfo
	if err := c.ShouldBindJSON(&s); err != nil {
		log.Error().Err(err).Msgf("HandleGithubRequest | failed to bind service info")
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed to get service info",
			"status":  "failed",
		})
		return
	}

	token, err := getArgoCDAdminToken()
	if err != nil {
		log.Error().Err(err).Msg("SyncApplication | failed to get ArgoCD admin token")
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to get ArgoCD admin token",
			"error":   fmt.Sprintf("%v", err),
			"status":  "failed",
		})
		return
	}

	path := fmt.Sprintf("api/v1/applications/%s/sync", s.ApplicationName)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", argoUrl, path), nil)
	if err != nil {
		log.Error().Err(err).Msg("SyncApplication | failed to create request")
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to create request",
			"error":   fmt.Sprintf("%v", err),
			"status":  "failed",
		})
		return
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("SyncApplication | failed to send request")
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to send request",
			"error":   fmt.Sprintf("%v", err),
			"status":  "failed",
		})
		return
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("SyncApplication | failed to read response body")
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to read response body",
			"error":   fmt.Sprintf("%v", err),
			"status":  "failed",
		})
		return
	}

	log.Info().Msgf("SyncApplication | sync request to argocd succeeded - Application: %s, Namespace: %s ", s.ApplicationName, s.ApplicationNamespace)

	healthCheckResult, h := serviceHealthCheck(s.ApplicationName, s.ApplicationNamespace)
	if !healthCheckResult {
		err := sendHealthCheckFailMessage(s.ApplicationName, s.SlackWebhookUrl, *h)
		log.Err(err).Msg("SyncApplication | Failed to check health check")
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "server health check failed",
			"error":   fmt.Sprintf("%v", err),
			"status":  "failed",
		})
		return
	}

	switch s.Branch {
	case "prod":
		err := sendDeployRequestMessage(s)
		if err != nil {
			log.Error().Err(err).Msg("SyncApplication | Failed to send deploy request")
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to send deploy request",
				"error":   fmt.Sprintf("%v", err),
				"status":  "failed",
			})
			return
		}
	default:
		err := sendUpdateSuccessMessage(s)
		if err != nil {
			log.Error().Err(err).Msg("SyncApplication | Failed to send update success message")
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to send update success message",
				"error":   fmt.Sprintf("%v", err),
				"status":  "failed",
			})
			return
		}
	}
	return
}

func getArgoCDAdminToken() (string, error) {
	url := fmt.Sprintf("%s/api/v1/session", argoUrl)
	username := os.Getenv("ARGO_ADMIN_USERNAME")
	password := os.Getenv("ARGO_ADMIN_PASSWORD")

	payload := map[string]string{
		"username": username,
		"password": password,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", errors.New(fmt.Sprintf("getAdminToken | failed to marshal payload to json: %s", err))
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", errors.New(fmt.Sprintf("getAdminToken | failed to create request: %s", err))
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.New(fmt.Sprintf("getAdminToken | failed to send request: %s", err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New(fmt.Sprintf("getAdminToken | failed to read response body: %s", err))
	}

	log.Debug().Msgf("getAdminToken | response status: %s, response body: %s", string(resp.Status), string(body))

	var response map[string]string
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", errors.New(fmt.Sprintf("getAdminToken | failed to unmarshal response body: %s", err))
	}

	token, ok := response["token"]
	if !ok {
		return "", errors.New(fmt.Sprintf("getAdminToken | failed to get token from response: %s", err))
	}

	return token, nil
}
