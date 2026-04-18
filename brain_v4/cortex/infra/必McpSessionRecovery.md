# MCP SSE Session Recovery (재연결 복원성)

## Context
NeuronFS는 지속해서 자가진화를 수행하며 `neuronfs.exe`가 잦은 빌드/재시작을 거칩니다. 이 과정에서 9247 포트를 쓰는 MCP Streamable HTTP 서버의 인메모리 세션 정보(Session ID)가 소멸하게 됩니다.

## Problem
Antigravity나 VS Code 등의 MCP 클라이언트는 기존 발급된 Session ID (`YS56IE5...` 같은 식별자)를 유지한 채 `failed to reconnect` 에러를 남발하며 재시도를 반복합니다. 타임아웃이나 서버 측 연결 거부(`actively refused`)를 받으면, 현재 IDE 자체를 윈도우 창 릴로드(`Developer: Reload Window`) 하지 않는 한 정상화되지 못하는 치명적 핑퐁 단절이 발생합니다.

## Axiom / Rule (필수)
- **必McpSessionRecovery**: MCP 서버 구현부 혹은 클라이언트 랩퍼에서, 5회 이상 재연결을 시도하여도 Target Machine Actively Refused 등 연결에 실패할 경우, 강제로 기존 클라이언트 객체를 파괴(Dispose)하고 처음부터 완전한 새 세션(Handshake)을 요청하도록 복원력을 부여해야 합니다.
- **禁재시작단절금지**: 시스템 자가진화(Re-build & Restart)가 클라이언트와의 연결 끊김 영구 장애로 이어지지 않게 하십시오. Hot Reload 아키텍처 구상을 모색하십시오.

---
// 이 파일은 마스터 프롬프트 지시에 따라 AI 에이전트에 의해 자동 성장(Grow)된 코어 뉴런입니다.
