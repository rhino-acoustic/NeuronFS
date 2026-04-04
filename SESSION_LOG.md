# 🧠 NeuronFS 세션 백업 로그

> **Antigravity 세션 휘발 버그 대응용 영구 기록**

## 2026-04-04 T11:47 세션 (대화 ID: 6aa264bc)

### 문제: bat 실행 시 Antigravity 대화 히스토리 저장 안 됨

### 근본 원인 2가지 발견 및 수정 완료

#### 1. NODE_OPTIONS → v4-hook.cjs 주입이 세션 저장 오염
- `start_v4_swarm.bat` 라인 29에서 `NODE_OPTIONS=--require v4-hook.cjs` 설정
- v4-hook.cjs가 Antigravity 내부의 http2.connect, https.request, fetch를 패치
- 이 패치가 `cloudcode-pa.googleapis.com` gRPC 스트림을 인터셉트하면서 세션 저장 API도 오염
- **수정**: NODE_OPTIONS 라인 주석처리. MCP 서버가 이미 대체 기능 제공.

#### 2. bat 파일 CP949 인코딩 → IDE 편집 도구 교착
- bat 파일이 CP949(한국어 윈도우 기본)로 저장됨
- Antigravity의 replace_file_content 도구가 UTF-8로 처리 시도 → TargetContent 매칭 실패 → 교착/지연
- 이중 인코딩으로 한글 주석 영구 손상됨
- **수정**: bat 파일 전체를 올바른 한글 + UTF-8로 재작성

### 추가 수정
- `settings.json`: `files.autoGuessEncoding: true` + `files.encoding: utf8` 추가
- `.vscode/settings.json`: 워크스페이스 레벨 설정도 추가
- 뉴런 교정: `hippocampus/에러패턴/bat_CP949_편집금지`, `열린파일_쓰기지연`

### 현재 상태
- bat 파일: UTF-8, NODE_OPTIONS 비활성화 완료
- Antigravity: CLEAN 환경으로 실행됨 (hook/proxy 없음)
- **다음 단계**: bat 재실행하여 대화 히스토리 보존 확인

---
*이전 로그는 start_v4_swarm.bat.bak에 원본 보존됨*
