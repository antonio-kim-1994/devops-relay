package handler

import "fmt"

var relayServers = map[string]addresses{
	"aws": {
		dev:  "https://dev-devops-relay.devnio.co.kr",
		prod: "https://prod-devops-relay.devnio.co.kr",
	},
}

func getTargetServerURL(appName, org, branch string) (string, error) {
	// 지정 org가 아닌 경우 error 반환
	orgs := []string{"org-a", "org-b"}
	isValidOrg := false

	for _, o := range orgs {
		if org == o {
			isValidOrg = true
			break
		}
	}

	if !isValidOrg {
		return "", fmt.Errorf("getTargetServerURL | invalid organization: %s", org)
	}

	// Github Action API로 전달된 조직 정보 검증
	relayEndpoint, exist := relayServers[org]
	if !exist {
		return "", fmt.Errorf("getTargetServerURL | unknown organization: %s", org)
	}

	// 브랜치에 따른 타겟 설정
	var target string
	if branch == "prod" {
		target = relayEndpoint.prod
	} else {
		target = relayEndpoint.dev
	}

	// 타겟이 비어있는지 확인
	if target == "" {
		return "", fmt.Errorf("getTargetServerURL | empty target endpoint for org: %s, branch: %s", org, branch)
	}

	return target, nil
}
