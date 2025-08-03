# DevOps Relay Platform

DevOps Relay는 GitHub Actions 및 Slack과 연동된 배포 자동화 플랫폼입니다.

- `gateway/`: 외부 이벤트 수신 및 내부 시스템으로의 요청 중계
- `server/`: ArgoCD 기반의 애플리케이션 배포 제어 및 Slack 인터랙션 처리

---

## 구성 모듈

### 1. gateway/

- 역할:  
  GitHub 또는 Slack에서 발생하는 이벤트를 수신하고, 각 팀별 서버 또는 중앙 배포 서버로 요청을 전달합니다.
- 기능:
    - GitHub Action 이벤트 중계
    - Slack 버튼 응답 처리
    - 서비스 헬스체크 중계
- 위치: [`/gateway`](./gateway)

### 2. server/

- 역할:  
  ArgoCD API를 통해 실질적인 애플리케이션 동기화, 롤아웃, 배포 승인/반려 등을 수행합니다.
- 기능:
    - ArgoCD 연동 (Sync/Promote/Abort)
    - Slack 메시지 자동화
    - Health Check 및 Slack 경고
- 위치: [`/server`](./server)