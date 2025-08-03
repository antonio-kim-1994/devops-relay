package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"net/http"
)

func HandleSlackResponse(c *gin.Context) {
	var r SlackResponse
	if err := c.ShouldBindJSON(&r); err != nil {
		log.Err(err).Msg("HandleSlackResponse | failed to bind request body")
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed to get service info",
			"status":  "failed",
		})
		return
	}

	switch r.Button.Result {
	case "approve":
		healthCheckResult, h := serviceHealthCheck(r.Button.ApplicationName, r.Button.ApplicationNamespace)
		if !healthCheckResult {
			log.Error().Msg("HandleSlackResponse | failed to send health check fail message")
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to send health check fail message",
				"status":  "failed",
			})
			err := sendHealthCheckFailMessage(r.Button.ApplicationName, r.ResponseURL, *h)
			if err != nil {
				log.Error().Err(err).Msg("HandleSlackResponse | failed to send health check fail message")
				return
			}
		}
		err := promoteApplication(fmt.Sprintf("%s-rollout", r.Button.ApplicationName), r.Button.ApplicationNamespace)
		if err != nil {
			log.Error().Err(err).Msgf("HandleSlackResponse | failed to promote application: %s", r.Button.ApplicationName)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to promote application",
				"status":  "failed",
			})
			return
		}

		reply := slackResponseForm{
			url:           r.ResponseURL,
			msg:           generateSlackTextBlock(fmt.Sprintf(":white_check_mark: *운영 배포 승인* | *%s* 사용자에 의해 *%s* 배포가 승인되었습니다.", r.User.Name, r.Button.ApplicationName)),
			replaceOption: true,
		}

		err = reply.sendResponseToSlack()
		if err != nil {
			log.Err(err).Msgf("HandleSlackResponse | failed to send result message to slack: %s", r.Button.ApplicationName)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to send result message to slack",
				"status":  "failed",
			})
			return
		}
		return
	case "reject":
		err := abortApplication(fmt.Sprintf("%s-rollout", r.Button.ApplicationName), r.Button.ApplicationNamespace)
		if err != nil {
			log.Error().Err(err).Msgf("HandleSlackResponse | failed to abort application: %s", r.Button.ApplicationName)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to abort application",
				"status":  "failed",
			})
			return
		}

		reply := slackResponseForm{
			url:           r.ResponseURL,
			msg:           generateSlackTextBlock(fmt.Sprintf(":no_entry: *운영 배포 반려* | *%s* 사용자에 의해 *%s* 배포가 반려되었습니다.", r.User.Name, r.Button.ApplicationName)),
			replaceOption: true,
		}

		err = reply.sendResponseToSlack()
		if err != nil {
			log.Err(err).Msgf("HandleSlackResponse | failed to send result message to slack: %s", r.Button.ApplicationName)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to send result message to slack",
				"status":  "failed",
			})
			return
		}
		return
	}
}
