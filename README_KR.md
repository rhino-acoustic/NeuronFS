# NeuronFS: AGI를 위한 제로 트러스트 가상 파일 시스템

> **"mkdir이 벡터 DB를 이긴다."** -- 0바이트 폴더 하나가 결정론적 O(1) 벽이다. 벡터 DB는 확률적 O(n) 추측이다.

---

## 한 줄 요약

AI 에이전트의 뇌를 **OS 폴더 구조**로 물리적으로 통제한다. 프롬프트 엔지니어링도, 벡터 DB도, GPU도 필요 없다.

## 빠른 시작

```bash
# 설치
git clone https://github.com/rhino-acoustic/NeuronFS.git
cd NeuronFS/runtime && go build -o neuronfs .

# 규칙 생성 = 폴더 생성
mkdir 禁_fallback              # 禁 접두사 = 절대 금지 (Zod 룬)
# 끝. 0바이트 폴더 자체가 규칙이다.

# 컴파일 = 시스템 프롬프트 자동 생성
./neuronfs --emit cursor        # -> .cursorrules
./neuronfs --emit claude        # -> CLAUDE.md
./neuronfs --emit all           # -> 모든 AI 포맷 동시 출력 (Cursor, Windsurf, Claude Desktop...)
```

## 옵코드는 룬워드다

디아블로 2를 해봤다면 -- **NeuronFS 옵코드는 룬워드와 정확히 같은 원리로 작동한다.**

| 옵코드 | 룬 | 효과 | 예시 |
|---|---|---|---|
| 禁/ | Zod | **절대 금지** -- AI가 물리적으로 넘을 수 없는 벽 | 禁/하드코딩/ |
| 必/ | Ber | **필수 게이트** -- 반드시 통과해야 함 | 必/부서장승인/ |
| 推/ | Ist | **추천** -- 부드러운 넛지, 무시 가능 | 推/테스트코드/ |
| .axon | Jah | **텔레포트** -- 먼 뇌 영역을 연결 | 推/보험료.axon => [보험금/] |
| bomb | El | **킬 스위치** -- 영역 전체 동결 | bomb.neuron |

> *"폴더가 소켓이고, 옵코드가 룬이다. 조합이 룬워드다."*

## 폴더 이름은 시제를 가진다

| 옵코드 | 시제 | 의미 |
|---|---|---|
| 禁_ | **현재형** -- "지금 금지되어 있다" | 능동적 차단벽 |
| 必_ | **현재진행형** -- "지금 진행 중이어야 한다" | 지속적 필수 게이트 |
| 推_ | **미래형** -- "다음에 이것을 해야 한다" | 전방 추천 |
| 검증_후_보고 | **미래의도** -- "완료 후 검증할 것이다" | 선언된 미래 행동 |

## 7계층 뇌 구조

```
brain_v4/
+-- brainstem/     (P0 -- 뇌간, 절대 원칙)
+-- limbic/        (P1 -- 변연계, 감정 필터)
+-- hippocampus/   (P2 -- 해마, 기억/에러 패턴)
+-- sensors/       (P3 -- 감각, 환경 제약)
+-- cortex/        (P4 -- 피질, 지식/기술)
+-- ego/           (P5 -- 자아, 성향/톤)
\-- prefrontal/    (P6 -- 전두엽, 목표/계획)
```

**핵심:** 낮은 P가 높은 P를 항상 물리적으로 덮어쓴다. brainstem(P0)의 禁은 cortex(P4)의 모든 규칙을 이긴다.

## VFS 엔진 아키텍처

- **RouterFS (vfs_core.go)**: O(1) Copy-on-Write 라우팅
- **Boot Ignition (vfs_ignition.go)**: Argon2id KDF 브레인월렛 통합
- **Crypto Cartridge (crypto_cartridge.go)**: XChaCha20-Poly1305 RAM 기반 .jloot 복호화

## 공식 위키

> **[NeuronFS 공식 위키 바로가기](https://github.com/rhino-acoustic/NeuronFS/wiki)** (한/영 이중 언어)

- **[Getting Started](https://github.com/rhino-acoustic/NeuronFS/wiki/Getting-Started)** -- 5분 퀵스타트
- **[Act 1: 의심과 발견](https://github.com/rhino-acoustic/NeuronFS/wiki/Act-1)** (Ep.01-07)
- **[Act 2: 시련과 워게임](https://github.com/rhino-acoustic/NeuronFS/wiki/Act-2)** (Ep.08-11)
- **[Act 3: 증명과 벤치마크](https://github.com/rhino-acoustic/NeuronFS/wiki/Act-3)** (Ep.12-16)
- **[Act 4: 선언과 울트라플랜](https://github.com/rhino-acoustic/NeuronFS/wiki/Act-4)** (Ep.17-22)

## 라이선스

본 프로젝트는 **AGPL-3.0** 라이선스 하에 배포됩니다. 상용 이용 시 별도 조항이 적용됩니다. [LICENSE](LICENSE)를 참조하십시오.

---
> *Created by 박정근 (PD) -- rubisesJO777*
> *Architecture: 26 Go files, ~10,400 lines. Single binary. Zero dependencies.*