# NeuronFS 블로그 시리즈 #16

## 7계층 뇌 — brainstem(P0)이 cortex(P4)를 항상 이긴다

> *Act IV: 증명 — Part 5/5 | 서사적 역할: 내부 구조 공개*

---

### Subsumption Cascade

```
brainstem (P0) ←→ limbic (P1) ←→ hippocampus (P2) ←→ sensors (P3)
     ↕                                                      ↕
cortex (P4) ←→ ego (P5) ←→ prefrontal (P6)
```

**낮은 P가 높은 P를 항상 우선.** bomb은 전체 정지.

| 계층 | 역할 | 예시 |
|---|---|---|
| P0 brainstem | 절대 원칙 | 禁_보안위반 |
| P1 limbic | 감정 필터 | 분노 감지 → 톤 조절 |
| P2 hippocampus | 기억 | 에러 패턴 기록 |
| P3 sensors | 환경 | NAS 경로, 브랜드 |
| P4 cortex | 지식 | React 규칙, DB 패턴 |
| P5 ego | 성향 | 말투, 포맷 |
| P6 prefrontal | 목표 | 스프린트 계획 |

### 3-Tier Emit Flow

```
neuronfs --emit       → stdout ONLY (미리보기)
neuronfs --emit all   → IDE 파일 + _rules.md
neuronfs --inject     → GEMINI.md + _index.md + _rules.md (전체 주입)
```

### 자율 진화 (evolve.go)

```
[대화 중] processChunk → _signals/*.json (기록만)
[30분마다] 자동 스케줄러 → neuronfs --evolve (Groq 분석)
[승격 시] 🧬 NEURON EVOLVED → 텔레그램 알림
```

AI가 대화하면서 발견한 패턴 → 시그널로 기록 → Groq가 분석 → 뉴런으로 승격/폐기. **뇌가 스스로 자란다.**

---

### 다음 편 예고 → Act V 시작: **편 17: "법조계+세무법"**
