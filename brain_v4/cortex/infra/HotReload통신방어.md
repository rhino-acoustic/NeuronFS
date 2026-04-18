# HotReload통신방어

## 배경
NeuronFS는 IDE-Telegram 간의 무한 연쇄 루프(Spiral Evolution) 속에서 지속적으로 컴파일/빌드되며 시스템 자신을 갱신합니다.
기존 아키텍처에서는 Supervisor(부모 데몬) 전체를 껐다 켜면서 MCP 통신 소켓이 닫혀 핑퐁이 끊어지는 심각한 결함이 있었습니다.

## 무중단 통신 교리 (Axiom)
1. **Network 망루의 영속성**: HTTP/TCP `:9247` 망루(역방향 프록시)와 메인 Supervisor 프로세스는 코어가 교체되어도 *절대* 종료하지 않습니다. `taskkill`은 더 이상 사용하지 않습니다. 
2. **Stateless Worker 이양(Delegate)**: MCP 요청(Tool Call)이 들어오면 내부 함수(`growNeuron`, `readBrain` 등)를 메모리 주소에서 직접 호출하는 것을 엄격히 금지합니다.
3. **CLI 스폰(Spawn)**: 모든 도구 로직은 반드시 최근 빌드된 `neuronfs.exe c:\... --tool <name> <params>` 형태로 `exec.Command` 를 통해 자식 프로세스로 실행(Hot Swap)해야 합니다.
4. 이 구조를 통해 코드가 1만 번 수정되어도, 통신 세션은 영구적으로 유지되며 로직은 항시 최신 바이너리의 지배를 받게 됩니다.
