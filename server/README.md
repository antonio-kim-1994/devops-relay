# DevOps Relay Server
DevOps Relay Server는 GitHub Actions 및 Slack으로부터 전달되는 배포 요청을 처리하고, ArgoCD 기반의 애플리케이션을 자동으로 동기화 및 롤아웃하는 API 중계서버입니다. 또한 서비스 상태 확인 및 Slack 알림 기능을 포함하고 있으며, AWS Secrets Manager를 활용한 보안 구성과 환경 변수 관리 기능을 제공합니다.

---
## 주요 기능
- GitHub 웹훅 요청 기반 ArgoCD 애플리케이션 동기화 및 배포 자동화
- Slack 인터랙션 기반 배포 승인 또는 반려 처리
- Kubernetes 서비스 헬스체크 및 실패 시 Slack Webhook 경고 발송
- ArgoCD REST API 기반 롤아웃 프로모션 및 중단 지원
- AWS Secrets Manager에서 보안 환경 변수를 로드 및 자동 적용

---
## 디렉토리 구조
```
.
├── server.go                          # 메인 진입점
├── config/
│   └── service_config.go              # Secrets Manager 설정 및 환경변수 적용 로직
├── handler/
│   ├── handler_github_request.go     # GitHub 요청 처리 및 ArgoCD 동기화
│   ├── handler_slack_response.go     # Slack 버튼 응답 처리
│   ├── handler_argocd.go             # ArgoCD 롤아웃 프로모션 및 중단 처리
│   ├── server_health_check.go        # 내부 서비스 헬스체크 수행
│   ├── slack_message.go              # Slack 메시지 전송 유틸리티
│   └── type_common.go                # 공통 타입 정의
├── middleware/
│   └── validate_api_request.go       # Request-Auth 헤더 기반 인증 미들웨어
```
---
## API 엔드포인트
### 1. 헬스체크
- `GET /healthz/healthcheck`  
  Gateway 자체 헬스 상태 확인
- `POST /sys/healthcheck`  
  요청된 서비스(application, namespace)에 대해 내부 클러스터 DNS로 헬스체크 수행

### 2. GitHub 동기화 요청
- `POST /update/github`  
  GitHub Actions로부터 배포 요청 수신 후 ArgoCD 애플리케이션 동기화 요청.  
  브랜치가 `prod`인 경우 Slack 배포 승인 요청 메시지 전송. 그 외에는 성공 메시지 전송.

### 3. Slack 배포 승인/반려 처리
- `POST /update/slack`  
  Slack 버튼 응답을 처리하여, ArgoCD 롤아웃을 프로모션하거나 중단합니다.  
  승인 시 사전 헬스체크 수행 후 진행되며 실패 시 Slack에 경고 메시지 전송.
---
## ArgoCD 연동
- 애플리케이션 동기화  
  `POST /api/v1/applications/{name}/sync`

- ArgoCD Rollout promote   
  `PUT /api/v1/rollouts/{namespace}/{rollout}/promote`

- ArgoCD Rollout Abort  
  `PUT /api/v1/rollouts/{namespace}/{rollout}/abort`

- Admin 인증 토큰 획득   
  `POST /api/v1/session` 요청 시 `ARGO_ADMIN_USERNAME`, `ARGO_ADMIN_PASSWORD` 사용

---

## AWS Secrets Manager 연동
설정 정보는 `/secret/devops`라는 Secret 이름에서 로드됩니다.

### 포함된 Secret Key
| Key                       | 설명                                      |
|--------------------------|-------------------------------------------|
| `REQUEST_TOKEN`          | 모든 외부 API 요청의 인증 헤더 값         |
| `ARGO_ADMIN_USERNAME`    | ArgoCD 인증용 관리자 계정 ID               |
| `PROD_ARGO_ADMIN_PASSWORD` | 운영 환경용 ArgoCD 관리자 비밀번호        |
| `DEV_ARGO_ADMIN_PASSWORD`  | 개발 환경용 ArgoCD 관리자 비밀번호        |

### 적용 방식
- `APP_ENV` 값에 따라 `prod` 또는 `dev` 비밀번호를 선택
- `SERVER_PORT`, `TIMEZONE` 등의 값은 환경변수로 덮어쓰기 가능
- Timezone 설정 시 `time.Local`에 반영됨
---
## 환경 변수
| 변수명                  | 설명                                                        |
|-------------------------|-------------------------------------------------------------|
| `APP_ENV`               | 실행 환경 구분 (dev 또는 prod)                              |
| `SERVER_PORT`           | 서비스 바인딩 포트 (기본: 8080)                             |
| `TIMEZONE`              | 로컬 시간대 설정 (기본: Asia/Seoul)                         |
| `ARGO_ADMIN_USERNAME`   | ArgoCD 관리자 계정 (Secrets Manager에서 로드됨)             |
| `ARGO_ADMIN_PASSWORD`   | ArgoCD 관리자 비밀번호 (환경에 따라 다르게 로드됨)          |
| `REQUEST_TOKEN`         | API 인증을 위한 헤더 값 (Secrets Manager에서 로드됨)        |

---
## Slack 메시지 전송
Slack Webhook을 통해 다음 알림이 전송됩니다.
- 배포 요청 메시지 (GitHub 요청 시)
- 승인/반려 결과 메시지 (Slack 버튼 클릭 시)
- 헬스체크 실패 메시지

Slack 메시지에는 서비스 이름, 브랜치, 커밋 메시지, 담당자 정보 등이 포함됩니다.

---

## 실행 예시
```bash
export APP_ENV=dev
export SERVER_PORT=8080
go run server.go
```

Secrets는 자동으로 AWS Secrets Manager에서 로드됩니다.

---
## 기타 사항
- 서비스 헬스체크는 내부 DNS 기반으로 `svc.cluster.local` 형태의 URL에 HTTP GET 요청을 보냅니다.
- Slack 버튼에는 `org/branch/app/namespace/request_type/result` 형태의 값을 포함하여 응답 처리 시 활용합니다.
- 모든 인증 요청은 `Request-Auth` 헤더를 통해 수행되며, 값은 Secrets에서 주입됩니다.
