# 아키텍처 결정 사항 (변경 금지)

## 오토파일럿
- CLI 응답 → hlCDPInject → IDE 직접 주입 (텔레그램 아님)
- 마스터 프롬프트 v2 항상 포함
- 순환: CLI→CDP→Antigravity→EVOLVE→반복

## 전사 분류
- WorkDir = TempDir (GEMINI.md 간섭 방지)
- relaxed JSON 파싱

## emit
- 5-Tier 구조: Bootstrap→Index→Rules→SubRules→Codemap
- 매시간 크론 실행

## 보호
- brain_v4 전체 git 추적 (565+ 파일)
- brainstem/_health.md 자동 생성 (Verification-on-Resume)

## 빌드
- ide_integration.go에 main/fileExists/svLog 없음 (main.go에서 제공)
- svPatchAntigravityShortcuts, svRestoreConversation은 stub

## CLI 실행 (multi_agent.go)
- --yolo 필수 (자동 승인)
- --allowed-mcp-server-names neuronfs (MCP 연결 허용)
- --sandbox 사용 금지 (boolean flag trap → sandbox=true로 파싱)
- stdin pipe로 프롬프트 주입
