---
type: "Insight"
confidence: 0.98
tags: ["CRM", "SSOT", "FreeRider", "VGVR", "DataAnalysis"]
created: "2026-04-07"
---

# VGVR 프리라이더 및 명단 누락 패턴 분석 (Free-Rider Patterns)

SSOT 파이프라인에서 "출석은 2회 이상 했으나 활성 구독 내역이 없는 회원"을 추적했을 때 3가지 주요 아키텍처 패턴이 발견됨. 

## 1. The Eternal Session (장기 악용자)
- **증상**: 6개월 이상 전에 1DAY 혹은 SESSION_1 (1회권)을 결제한 기록만 존재하는데, 매월 출석부(attendance_records)에 꾸준히 참석자로 찍히는 인원. (예: 김도영)
- **원인**: 어드민/코치 앱의 UI에서 출석체크 대상 리스트를 띄울 때 '현재 잔여 횟수가 유효한가?'를 핑거프린팅하지 않고 이름만 검색해서 마킹하기 때문.
- **해결책**: 출석 앱 렌더링 단에서 `member.category === 'Black'` 인 회원은 아예 출석 버튼을 비활성화해야 함.

## 2. Mid-Month Cancellation (월 중도 취소 오탐)
- **증상**: 월 초중반에 참석을 정당하게 다수(3~4회) 진행하다가 월말에 환불/취소를 한 인원. (예: 조성임)
- **원인**: SSOT 로직(`isMemberSSOTActive`)의 무관용 취소자 원칙(`if (p.cancelled_at) return false;`) 때문. 이 원칙은 '순수 정기구독 통계액'의 거품을 빼는 데는 퍼펙트하게 작용하지만, 트래커 입장에서는 "결제도 안 하고 출석한 사람"으로 False Positive 판단을 내림.
- **해결책**: 트래커에서 "이 사람이 무임승차자인가?"를 판단하려면 SSOT와는 별개로 `cancelled_at이 출석일(attendance_date)보다 이후인가?`를 검증하는 시간차 로직이 추가되어야 완벽함.

## 3. Zero Record Ghost (결제 내역 제로)
- **증상**: `user_purchases`, `naver`, `cafe24` 그 어떤 원시 DB에도 해당 이름/전화번호가 존재하지 않는 유령 회원. (예: 김명대)
- **원인**: 현장 현금 결제 후 시스템 미등록, 가족/지인 대리 결제(Proxy Payment), 혹은 프로필의 동명이인/오타 파싱 실패.
- **해결책**: 어드민 트래커 메뉴에서 Proxy 매핑(가족계정 연결)을 지원하거나, 현금 수동 등록(MANUAL) 기능을 강력하게 제공해야 함.
