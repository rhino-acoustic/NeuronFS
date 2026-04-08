# [ENG] Jloot VFS Architecture
# [KOR] Jloot VFS 아키텍처 명세

# Jloot VFS 엔진 그랜드 오픈 (Phase 2 완성)

## 📌 핵심 요약
초기 구상에 불과했던 **"물리적 폴더 구조를 암호화된 ZIP 카트리지 형태로 묶고 투명하게 단일 구조망으로 라우팅한다"**는 개념을 마침내 완전한 OS 수준의 Go 코드로 구현을 끝마쳤습니다. 이로써 AI 에이전트는 무한대로 증식가능한 하드디스크가 아닌, 기억 카트리지를 "장착/해제" 할 수 있는 메모리 독립적 존재로 거듭납니다.

---

## 🛠️ 주요 변경 사항

### 1. 🔑 **Brainwallet 점화 시퀀스 (Zero Drops)**
- **CLI Boot**: `term.ReadPassword`를 이용한 마스킹된 브레인월렛 프롬프트 입력창 확보 (`vfs_ignition.go`). 
- **인메모리 해독**: 저장된 Mnemonic을 `Argon2id`로 파생시킨 후, `XChaCha20`으로 압축된 `.jloot` 파일 자체를 디스크에 단 한 번도 풀지 않고 **100% RAM 공간 (bytes.Reader) 상에서 해제**합니다 (`crypto_cartridge.go`).
- **보안성 확보**: 카트리지 데이터는 사용자의 런타임 메모리에만 떠 있다가 전원이 꺼지면 소멸됩니다.

### 2. 🛡️ **쓰기 샌드박싱 (Copy-on-Write)**
- 과거 지식(Lower, Cartridge)에 대해 활성화(fire), 복구(rollback) 등 **디스크 쓰기(Write)** 요청이 올 경우 `neuron_crud.go`의 라우터가 이를 검열합니다.
- 물리적인 딥 파일시스템(ZIP 내부)을 삭제하려는 시도를 차단하고, 자동으로 UpperDir(물리 디스크) 경로에 폴더 깊은 곳까지 `os.MkdirAll`을 뚫고 들어가 **섀도우 복제본**을 기록합니다.
- `vfsReadDir`는 이 둘을 동시에 반환하며, `scanBrain`은 최댓값 뉴런을 항상 우선 탐색(O(1) Route)하므로 하위 구조망이 완벽히 덮어씌워집니다.

---

## 🏗️ 업데이트된 아키텍처 구조

```mermaid
graph TD
    A[Mnemonic Input] -->|Argon2id| B(32B Master Key)
    B -->|XChaCha20| C{crypto_cartridge.go}
    D[base.jloot File] --> C
    C -->|Extract purely in RAM| E[bytes.Reader Payload]
    E -->|zip.NewReader| F[Virtual Lower Directory]
    G[Physical UI/HDD] -->|O(1) Route| H[Virtual Upper Directory]
    F -->|vfs_core.go| I((Global VFS Shadowing Router))
    H -->|Copy-on-Write / Sandboxing| I
```

## 🚀 다음 지향점
> [!NOTE] 
> 현재 엔진 코어 구현이 일단락됨에 따라, 초안으로 정리해 뒀던 블로그 에필로그(`blog_epilogue.md`)와 잠재력 문서(`blog_23_the_100_potentials.md`)들을 GitHub Wiki 또는 정식 백서(Whitepaper)로 배포/완성하는 단계를 검토할 수 있습니다.
