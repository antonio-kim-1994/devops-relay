# DevOps Relay Gateway

**DevOps Relay Gateway**는 GitHub Actions, Slack 요청을 받아 조직 내부 배포 시스템으로 요청을 라우팅하고, Datadog 로그 수집까지 처리하는 중간 게이트웨이 역할을 수행하는 Go 기반 마이크로서비스입니다.

---
## 주요 기능
- **Health Check 프록시**: 외부에서 전달된 서비스 상태 요청을 내부 서버로 중계
- **GitHub Action 요청 중계**: 배포 요청 수신 및 내부 서버 동기화
- **Slack 버튼 응답 처리**: 배포 승인/반려 요청 처리 및 Slack 메시지 응답 전송
- **Datadog 로그 수집**: 배포 메타데이터를 Datadog에 기록
- **보안 인증**: API 토큰 및 Slack 서명 검증 기능 내장
- **AWS Secrets Manager 기반 환경설정 자동 로딩**
---
## 기술 스택

| 영역               | 기술                                                                 |
|--------------------|----------------------------------------------------------------------|
| 언어               | Go (Golang)                                                          |
| 웹 프레임워크      | [Gin](https://github.com/gin-gonic/gin)                             |
| 로깅               | [Zerolog](https://github.com/rs/zerolog)                             |
| 클라우드           | AWS Secrets Manager, Datadog                                          |
| 외부 연동          | GitHub Webhook, Slack Interactive Message                            |

---

## 디렉토리 구조

```
gateway/
├── config/                    # AWS Secrets 기반 구성 로딩
│   └── service_config.go
├── handler/                   # 주요 엔드포인트 핸들러
│   ├── handler_github_request.go
│   ├── handler_slack_payload.go
│   ├── handler_server_endpoint.go
│   ├── server_health_check.go
│   ├── datadog_log_ingestion.go
│   └── type_common.go
├── middleware/                # 요청 유효성 검증 미들웨어
│   ├── validate_api_request.go
│   ├── validate_slack_payload.go
│   └── generate_hmac.go
├── server.go                  # 메인 엔트리 포인트
```
---
## API 엔드포인트

### 헬스 체크

| Method | Endpoint                       | 설명                             |
|--------|--------------------------------|----------------------------------|
| GET    | `/healthz/healthcheck`         | Gateway 자체 헬스체크             |
| POST   | `/sys/healthcheck`             | 내부 서비스 헬스체크 요청 중계    |
**응답**
```json
{
  "application_name": "svc-name",
  "org": "org-a",
  "branch": "dev"
}
```
---
### GitHub 배포 요청 중계

| Method | Endpoint                 | 설명                                  |
|--------|--------------------------|---------------------------------------|
| POST   | `/v2/github/update`      | GitHub Action 요청 수신 및 내부 전파 |

> 인증 필요: `Authorization: Bearer <AUTH_TOKEN>`

**응답**
```json
{
  "message": "my-app | Sync success.",
  "status": "success"
}
```

---

### Slack 배포 승인/반려 처리

| Method | Endpoint                 | 설명                                 |
|--------|--------------------------|--------------------------------------|
| POST   | `/v2/slack/deploy`       | Slack 버튼 액션 응답 처리           |

> 서명 검증 수행: `X-Slack-Signature`, `X-Slack-Request-Timestamp`
---
## 인증 및 보안

- 모든 요청은 다음을 기반으로 검증됩니다:
    - `Authorization` 헤더 (`AUTH_TOKEN`)
    - Slack 요청 서명 (`SLACK_BOT_SIGNING_SECRET`)
    - 내부 서버 간 통신에는 `REQUEST_TOKEN` 헤더 사용

- 모든 보안 정보는 **AWS Secrets Manager**의 `/secret/devops` 에서 로딩되며, 다음 항목 포함:
    - `SLACK_BOT_SIGNING_SECRET`
    - `AUTH_TOKEN`
    - `REQUEST_TOKEN`
    - `DATADOG_API_KEY`
    - `DATADOG_SITE`

---

## 실행 방법

### 1. 필요한 환경 구성

- AWS IAM 권한: SecretsManager 읽기 권한 필요
- SecretsManager 예시
```json
{
  "SLACK_BOT_SIGNING_SECRET": "xxx",
  "AUTH_TOKEN": "xxx",
  "REQUEST_TOKEN": "xxx",
  "DATADOG_API_KEY": "xxx",
  "DATADOG_SITE": "datadoghq.com"
}
```
### 2. 실행
```bash
go run server.go
```
환경변수로 포트 지정 가능 (`SERVER_PORT`), 기본값은 `8080`.

---
## Datadog 로그 예시
배포 성공 시 다음 형식으로 Datadog Logs 전송:
```json
{
  "message": "[dev] myapp service deployed.",
  "tags": "env:dev",
  "date": "2025-08-03",
  "org": "org-a",
  "repository": "repo-name",
  "branch": "dev",
  "commit": "feat: add x logic"
}
```
---

## 지원 조직 및 서버 경로
현재 허용된 조직: `org-a`, `org-b`

| Branch   | 대상 URL                                      |
|----------|-----------------------------------------------|
| `prod`   | `https://prod-devops-relay.devnio.co.kr`      |
| others   | `https://dev-devops-relay.devnio.co.kr`       |