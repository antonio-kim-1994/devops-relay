package handler

import (
	"errors"
	"fmt"
	"github.com/slack-go/slack"
	"time"
)

func sendHealthCheckFailMessage(serviceName, slackWebhookUrl string, h HealthCheck, replaceOption ...bool) error {
	if slackWebhookUrl == "" {
		return errors.New("slack webhook url is empty")
	}

	// Message Block
	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", ":warning: *Health Check 실패* :warning:", false, false),
				nil,
				nil,
			),
			slack.NewDividerBlock(),
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", "업데이트 예정 서비스의 Health Check를 실패했습니다.\n서비스 배포를 위해 관리자(@devops)에 문의하세요.", false, false),
				nil,
				nil,
			),
			slack.NewSectionBlock(
				slack.NewTextBlockObject(
					"mrkdwn",
					fmt.Sprintf("> *대상 서비스*: %s\n> *Health Check 횟수*: `%d 회`\n> *Health Check 간격*: `%d 초`", serviceName, h.limits, int(h.interval/time.Second)),
					false,
					false,
				),
				nil,
				nil,
			),
		},
	}

	replaceOriginal := false
	if len(replaceOption) > 0 {
		replaceOriginal = replaceOption[0]
	}

	msg := slack.WebhookMessage{Blocks: &blocks, ReplaceOriginal: replaceOriginal}

	err := slack.PostWebhook(slackWebhookUrl, &msg)
	if err != nil {
		return fmt.Errorf("sendHealthCheckFailMessage | failed to post slack webhook: %w", err)
	}

	return nil
}

func sendDeployRequestMessage(s ServiceInfo) error {
	if s.SlackWebhookUrl == "" {
		return errors.New("slack webhook url is empty")
	}

	repoUrl := fmt.Sprintf("*서비스:*\n*<https://github.com/%s/%s|%s/%s>*", s.Org, s.Repo, s.Org, s.Repo)
	operator := fmt.Sprintf("*담당자:*\n@%s", s.Operator)
	branch := fmt.Sprintf("*업데이트 브랜치:*\n`%s`", s.Branch)
	date := fmt.Sprintf("*업데이트 일시:*\n%s", s.Date)
	commit := fmt.Sprintf("*업데이트 내용*\n%s", s.CommitMessage)
	approveBtn := fmt.Sprintf("%s/%s/%s/%s/deploy/approve", s.Org, s.Branch, s.ApplicationName, s.ApplicationNamespace)
	rejectBtn := fmt.Sprintf("%s/%s/%s/%s/deploy/reject", s.Org, s.Branch, s.ApplicationName, s.ApplicationNamespace)

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", "🚀 *Production 배포 승인 요청* 🚀", false, false),
				nil,
				nil,
			),
			slack.NewDividerBlock(),
			slack.NewSectionBlock(nil, []*slack.TextBlockObject{
				slack.NewTextBlockObject("mrkdwn", repoUrl, false, false),
				slack.NewTextBlockObject("mrkdwn", operator, false, false),
			}, nil),
			slack.NewSectionBlock(
				nil,
				[]*slack.TextBlockObject{
					slack.NewTextBlockObject("mrkdwn", branch, false, false),
					slack.NewTextBlockObject("mrkdwn", date, false, false),
				},
				nil,
			),
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", commit, false, false),
				nil,
				nil,
			),
			slack.NewActionBlock("action_block", // Action 블록 ID
				slack.NewButtonBlockElement("approve", approveBtn,
					slack.NewTextBlockObject("plain_text", "승인", true, false),
				).WithStyle("primary"),
				slack.NewButtonBlockElement("deny", rejectBtn,
					slack.NewTextBlockObject("plain_text", "반려", true, false),
				).WithStyle("danger"),
			),
			slack.NewContextBlock("context_block",
				slack.NewTextBlockObject("mrkdwn", ":warning: *승인 버튼 클릭 시 신규 서비스가 배포됩니다.*\n:pushpin: *배포 과정에 장애가 발생한 경우 DevOps 팀에 문의주시기 바랍니다.*", false, false),
			),
		},
	}

	msg := slack.WebhookMessage{Blocks: &blocks}
	err := slack.PostWebhook(s.SlackWebhookUrl, &msg)
	if err != nil {
		return err
	}
	return nil
}

func sendUpdateSuccessMessage(s ServiceInfo) error {
	repoUrl := fmt.Sprintf("*서비스:*\n*<https://github.com/%s/%s|%s/%s>*", s.Org, s.Repo, s.Org, s.Repo)
	operator := fmt.Sprintf("*담당자:*\n@%s", s.Operator)
	commit := fmt.Sprintf("*업데이트 내용*\n%s", s.CommitMessage)

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("🚀 *`%s` 서비스 배포 완료* 🚀", s.Branch), false, false),
				nil,
				nil,
			),
			slack.NewDividerBlock(),
			slack.NewSectionBlock(nil, []*slack.TextBlockObject{
				slack.NewTextBlockObject("mrkdwn", repoUrl, false, false),
				slack.NewTextBlockObject("mrkdwn", operator, false, false),
			}, nil),
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", commit, false, false),
				nil,
				nil,
			),
			slack.NewContextBlock("context_block",
				slack.NewTextBlockObject("mrkdwn", ":pushpin: *배포 과정에 장애가 발생한 경우 DevOps 팀에 문의주시기 바랍니다.*", false, false),
			),
		},
	}

	msg := slack.WebhookMessage{Blocks: &blocks}
	err := slack.PostWebhook(s.SlackWebhookUrl, &msg)
	if err != nil {
		return err
	}
	return nil
}

type slackResponseForm struct {
	url           string
	msg           slack.Blocks
	replaceOption bool
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
