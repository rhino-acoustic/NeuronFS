---
description: "governance_consts.go: 모든 거버넌스 매직넘버의 SSOT. 값 변경은 이 파일만."
---
# Governance Constants (SSOT)

## 정의된 상수
| Const | 값 | 용도 |
|-------|-----|------|
| MergeThreshold | 0.6 | grow/dedup 유사도 병합 기준 |
| MaxEpisodes | 10 | session_log 순환 버퍼 |
| PruneDays | 3 | 推 뉴런 dormant 일수 |
| SessionLogCap | 3 | session_log 최대 파일 수 |
| DefaultEmotionIntensity | 0.6 | limbic 감정 기본 강도 |

## 원칙
- 코드에 숫자를 직접 쓰지 않는다
- DCI-08이 하드코딩을 자동 감찰
