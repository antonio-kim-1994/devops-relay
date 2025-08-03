package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type HealthCheck struct {
	limits   int
	interval time.Duration
}

func CommonHealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
	return
}

func ServerHealthCheck(c *gin.Context) {
	var h HealthCheckRequest
	if err := c.ShouldBindJSON(&h); err != nil {
		log.Error().Err(err).Msg("ServerHealthCheck | failed to bind healthcheck data to json")
	}
	c.JSON(200, gin.H{
		"message": fmt.Sprintf("Service Name: %s, Org: %s, Branch: %s", h.ApplicationName, h.Org, h.Branch),
	})
	return
}

func serviceHealthCheck(appName, namespace string) (bool, *HealthCheck) {
	// HealthCheck Default Config
	h := HealthCheck{
		limits:   25,
		interval: 5 * time.Second,
	}

	healthCheckLimits := os.Getenv("HEALTH_CHECK_LIMITS")
	healthCheckInterval := os.Getenv("HEALTH_CHECK_INTERVAL")

	// HEALTH_CHECK_LIMITS 파싱
	if healthCheckLimits != "" {
		if parsedLimits, err := strconv.Atoi(healthCheckLimits); err == nil {
			h.limits = parsedLimits
		} else {
			log.Info().Msgf("serviceHealthCheck | No HEALTH_CHECK_LIMITS value served. Set Default value(%d)", h.limits)
		}
	}

	// HEALTH_CHECK_INTERVAL 파싱
	if healthCheckInterval != "" {
		if parsedInterval, err := strconv.Atoi(healthCheckInterval); err == nil {
			h.interval = time.Duration(parsedInterval) * time.Second
		} else {
			log.Info().Msgf("serviceHealthCheck | No HEALTH_CHECK_INTERVAL value served. Set Default Value(%d)", h.interval)
		}
	}

	// 기본값 설정
	var url string
	switch appName {
	// Web Front
	case "homepage-front", "cms-front", "mydata-front", "pms-front", "mydata-cms-front":
		url = fmt.Sprintf("http://%s-preview.%s.svc.cluster.local/", appName, namespace)
	// Back-end
	default:
		url = fmt.Sprintf("http://%s-preview.%s.svc.cluster.local/healthz/healthcheck", appName, namespace)
	}

	count := 0
	for {
		resp, err := http.Get(url)
		if err != nil {
			log.Info().Msgf("serviceHealthCheck | Status Code: %d | [%s] health check retry [%d]", resp.StatusCode, appName, count)
		} else {
			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
				log.Info().Msgf("serviceHealthCheck | Status Code: %d | [%s] health check success", resp.StatusCode, appName)
				resp.Body.Close()
				return true, &h
			}
			resp.Body.Close()
		}

		count++

		// Healthcheck 실패 메시지 전송을 위해 healthcheck 설정값 return
		if count >= h.limits {
			log.Error().Msgf("serviceHealthCheck | [%s] health check fail", appName)
			return false, &h
		}
		time.Sleep(h.interval)
	}
}
