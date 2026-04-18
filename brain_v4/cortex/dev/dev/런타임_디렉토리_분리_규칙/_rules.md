---
name: 런타임_디렉토리_분리_규칙
description: NeuronFS runtime 하위 기능별 폴더 격리 규칙 및 버전 관리 지침
---
### 🛠️ 런타임 환경 (Runtime) 폴더 구조 격리 및 관리 지침

NeuronFS v4 이후 시스템이 커지면서 `runtime/` 폴더 내에 각종 스니퍼, 더프 파일, 임시 코드가 혼재하는 "환경 오염(Environment Pollution)" 현상이 발생했습니다. 이를 원천 차단하기 위해 런타임 구역을 하위 기능별로 철저히 격리(Isolation)합니다.

#### 1. 폴더 격리 조직도 (Directory Isolation Map)
앞으로 `runtime/` 루트에는 핵심 바이너리(`neuronfs.exe`) 및 코어 진입점 스크립트만 둡니다. 기능 파일들은 반드시 다음 하위 폴더에 격리하십시오:
- **`hijackers/`**: 네트워크 스니핑 및 탈취 스크립트 (예: `hijack-launcher.mjs`, `context-hijacker.mjs`, `v13`~`v19` 레거시 브리지 등)
- **`core_agents/`**: 에이전트 구동 및 주입 스크립트 (예: `kickstart.mjs`, `multi-agent.mjs`)
- **`proto_dumps/`**: gRPC/HTTP2 패킷 분석 시 떨어지는 바이너리 덤프 (예: `proto_desc_*.bin`)
- **`coverage_reports/`**: 테스트 실행 후 떨어지는 커버리지 및 정적 분석 리포트
- **`probes/`**: 시스템 탐지 목적의 일회성 프로브 (예: `probe.mjs`)
- **`tests_logs/`**: 스웜 실행 결과나 스트레스 테스트 로그 파일들

#### 2. 파일 스냅샷 버전 관리 지침
기능을 교체할 때 **절대 이전 코드를 덮어쓰고 지우지 마십시오**.
- (X) `hijack-launcher.mjs` 안의 코드를 통째로 삭제 후 재작성
- (O) `v20-hijack-cdp.mjs` 등 **버전명(접두사)을 붙인 새로운 파일을 생성**합니다.
- 새로운 버전이 작동함을 확인하면, 이전 파일들은 `hijackers/legacy/` 같은 곳으로 아카이브하거나 버전을 유지합니다. (단, 깃헙 등에 올릴 때는 `git` 커밋으로 이력을 남깁니다.)

이 규칙에 따라 `start_v4_swarm.bat`의 경로 또한 `%RUNTIME_DIR%\hijackers\hijack-launcher.mjs`처럼 기능별 구조를 타게 됩니다.
