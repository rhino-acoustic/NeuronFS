# Episode 15. Brainwallet & Zero-Trace Encryption
## "암호화도 가능하네" — 세상 어디에도 키가 저장되지 않는 암호 체계

> **Language Note / 언어 안내**
>
> **[ENG]** This document explains NeuronFS's Brainwallet encryption: the master key exists only in the creator's memory. No key file, no cloud, no USB. Combined with RAM-only mounting, the brain cartridge leaves zero forensic traces on disk.
>
> **[KOR]** 본 문서는 NeuronFS의 Brainwallet 암호화를 설명합니다. 마스터 키는 오직 창조주의 기억 속에만 존재하며, 어디에도 저장되지 않습니다. RAM 전용 마운트와 결합하여 디스크에 포렌식 흔적을 0으로 남깁니다.

---

### 킥

> **"암호화도 가능하네 소프트웨어 노하우를 담을 수 있어"**
> — essential L2751

> **"더 기발하게 하고싶어 내 머리가 잊지 않는 문장"**
> — essential L4311

---

### 문제: 전문가의 지식은 어떻게 보호하는가

세법 전문가가 10년간 축적한 판례 분석과 절세 전략. 의료 전문가의 희귀 질환 진단 패턴. 이것들을 NeuronFS 뇌에 담아 판매하려면, **레시피를 훔칠 수 없는 완벽한 보호 장치**가 필요하다.

기존 방식의 허점:
```
[일반 ZIP 암호화]
비밀번호 → 하드디스크에 풀기 → 폴더 노출 → 복사 가능
→ 한 번 풀면 끝. 원본 유출 불가역적.

[클라우드 키 관리 (AWS KMS 등)]
키가 클라우드에 '저장'됨 → 서버 해킹 시 키 탈취
→ 운영자도 키에 접근 가능 → 내부자 위협
```

---

### 해결 1: 인메모리 마운트 — 디스크에 0흔적

```
[NeuronFS .jloot 카트리지]
암호화된 .jloot 파일 → 하드디스크에 풀지 않음
→ RAM에 가상 드라이브로 마운트 (OverlayFS)
→ AI가 폴더 구조를 읽고 작동
→ 탐색기로 들여다보려 하면 → OS 레벨 접근 거부
→ 프로그램 종료 → RAM에서 즉시 증발
→ 디스크 포렌식 흔적: 0 bytes
```

구현 코드 (`vfs_mount.go` + `crypto_cartridge.go`):
```
passphrase → Argon2id KDF (100ms 의도적 지연) → Master Key 도출
Master Key → XChaCha20-Poly1305 → .jloot 스트림 복호화
복호화된 데이터 → RAM 버퍼에만 존재 → RouterFS Upper Layer로 마운트
종료 시 → SecureZero(buffer) → RAM 완전 소거
```

**"결과물은 누리되, 레시피는 훔칠 수 없는 완벽한 블랙박스."**

---

### 해결 2: Brainwallet — 마스터 키는 머리 속에만

> *"내 머리가 잊지 않는 문장"* → 그 문장이 곧 마스터 키

```
사용자 입력:    "우리가 처음 만난 건 비 오는 화요일이었다"
         ↓
Argon2id:       메모리 64MB, 반복 3회, 100ms 의도적 지연
         ↓      (무차별 대입 공격 비용: GPU 1대 × 2,700년)
SHA-256:        0x7f3a9b2c...
         ↓
XChaCha20 Key:  256-bit Master Key 도출 완료
```

**세상 어디에도 키가 '저장'된 곳이 없다.**
- USB에 없다
- 클라우드에 없다
- 종이에 없다
- 파일 시스템에 없다

**창조주의 기억 자체가 마스터 키.** 머릿속의 문장을 모르면, 전 세계의 슈퍼컴퓨터를 동원해도 .jloot를 열 수 없다.

---

### 비즈니스 적용: DRM 시한폭탄

법과 규정은 매년 바뀐다. 암호화 키에 유효기간을 설정:

```
2026_개정세법_통제망.jloot
→ 유효기간: 2026.12.31 24:00
→ 만료 시: 복호화 키 파기 → 카트리지 접근 불가
→ AI 응답: "2027년형 브레인 카트리지를 갱신하십시오."
→ 고객: 매년 구독료 결제

수익 모델:
  전문가: 10년의 노하우를 .jloot로 패키징 → Brain Marketplace 등록
  고객: 연간 구독 → 최신 법률/의료 뇌를 자동 갱신
  플랫폼: 거래 수수료 15%
```

> *"10,000시간의 전문 지식이 담긴 뇌를 15MB .jloot 칩에 봉인하고, 머릿속에만 존재하는 키로 잠근다. 이것이 Brain Marketplace의 DRM이다."*

---

[Back to Act 3](Act-3) | [Ep.14](Episode-14-Multimodal-Solved-by-OS) | [Ep.16](Episode-16-P0-Brainstem-Always-Overrides-P4-Cortex)
