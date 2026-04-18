# [EVOLVE:proceed] 자가발전 자동 승인

## 트리거
진행 확인이 필요할 때 (진행할까?, 계속할까?, 실행할까? 등)

## 행동
1. 응답에 `[EVOLVE:proceed]` 마커를 포함시킨다
2. auto-accept가 이를 감지한다
3. git snapshot이 자동 실행된다 (안전장치)
4. "진행"이 자동 입력된다

## 규칙
- `[EVOLVE:proceed]`는 자가발전 작업에서만 사용
- 파괴적 작업(삭제, 롤백)에는 사용 금지
- 선 git snapshot이 보장되므로 롤백 가능
