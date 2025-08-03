package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type slackResponseForm struct {
	url           string
	msg           slack.Blocks
	replaceOption bool
}

func SlackResponseHandler(c *gin.Context) {
	// Payload Parsing
	payload, err := parsePayload(c)
	if err != nil {
		log.Err(err)
		// Slack Callback을 위해 200 응답. 200 응답 외의 응답은 서비스 장애로 인식한다.
		c.JSON(http.StatusOK, gin.H{
			"message": "failed to parse payload",
			"status":  "failed",
		})
		return
	}

	// Button Value parsing
	splitPayload := strings.Split(payload.ActionCallback.BlockActions[0].Value, "/")
	log.Debug().Msgf("Slack Button Value: %+v", splitPayload)

	r := SlackResponse{
		ResponseURL: payload.ResponseURL,
		User: User{
			Name: payload.User.Name,
			ID:   payload.User.ID,
		},
		Button: ButtonValue{
			Org:                  splitPayload[0],
			Branch:               splitPayload[1],
			ApplicationName:      splitPayload[2],
			ApplicationNamespace: splitPayload[3],
			RequestType:          splitPayload[4],
			Result:               splitPayload[5],
		},
	}

	// Get Target Server URL
	url, err := getTargetServerURL(r.Button.ApplicationName, r.Button.Org, r.Button.Branch)
	if err != nil {
		// Slack Callback을 위해 200 응답. 200 응답 외의 응답은 서비스 장애로 인식한다.
		log.Error().Err(err).Msgf("failed to get target server")
		c.JSON(http.StatusOK, gin.H{
			"message": "failed to get target server endpoint",
			"status":  "failed",
		})
		return
	}

	log.Info().Msgf("SlackResponseHandler | target server: %s", url)

	// Slack Response
	// Server에서 처리 후 Slack 응답을 전송하기에는 환경이 분리되어 있어 처리에 시간 소요.
	// Gateway에서 선제적으로 응답 후 server에서 처리
	var reply slackResponseForm
	switch r.Button.Result {
	case "approve":
		reply = slackResponseForm{
			url:           r.ResponseURL,
			msg:           generateSlackTextBlock(fmt.Sprintf(":white_check_mark: *운영 배포 승인* | *%s* 사용자에 의해 *%s* 배포가 승인되었습니다.", r.User.Name, r.Button.ApplicationName)),
			replaceOption: true,
		}

		err = reply.sendResponseToSlack()
		if err != nil {
			log.Err(err)
			return
		}
	case "reject":
		reply = slackResponseForm{
			url: r.ResponseURL,
			msg: generateSlackTextBlock(fmt.Sprintf(":no_entry: *운영 배포 반려* | *%s* 사용자에 의해 *%s* 배포가 반려되었습니다.", r.User.Name, r.Button.ApplicationName)),
		}

		err = reply.sendResponseToSlack()
		if err != nil {
			log.Err(err)
			return
		}
	}

	// Relay Server로 데이터 전송
	err = sendSlackResponseToServer(&r, url)
	if err != nil {
		log.Error().Err(err).Msgf("failed to send service info")
		return
	}
	return
}

func parsePayload(c *gin.Context) (*slack.InteractionCallback, error) {
	// body data parse part
	// x-www-form-urlencoded parsing용
	rawPayload := c.PostForm("payload")

	if rawPayload == "" {
		return nil, errors.New("no payload received")
	}

	var payload slack.InteractionCallback
	err := json.Unmarshal([]byte(rawPayload), &payload)
	if err != nil {
		return nil, errors.New("failed to unmarshalling payload")
	}

	return &payload, nil
}

func sendSlackResponseToServer(s *SlackResponse, url string) error {
	path := "update/slack"
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
	log.Info().Msgf("Successfully sent request to %s", url)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.New(fmt.Sprintf("sendSlackResponse | failed to read response from %s", url))
	}

	log.Info().Msgf("sendGithubRequestInfo | response: %s", string(body))
	return nil
}

func generateSlackTextBlock(text string) slack.Blocks {
	return slack.Blocks{
		BlockSet: []slack.Block{
			slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", text, false, false), nil, nil),
		},
	}
}

func (s slackResponseForm) sendResponseToSlack() error {
	err := slack.PostWebhook(s.url, &slack.WebhookMessage{Blocks: &s.msg, ReplaceOriginal: s.replaceOption})
	if err != nil {
		return err
	}
	return nil
}
