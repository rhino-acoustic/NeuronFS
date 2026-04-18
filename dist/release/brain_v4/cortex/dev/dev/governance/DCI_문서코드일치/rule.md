---
description: "Axis 3 DCI: go test로 주석/const/코드 일치를 자동 검증. governance_benchmark_test.go TestDCI_Constants"
---
# DCI (Document-Code Integrity)

## 검증 축
- DCI-01~01d: governance_consts.go의 const 값 검증
- DCI-02~07: 실제 동작이 const와 일치하는지 기능 검증
- DCI-08: 소스코드에 매직넘버 하드코딩 감찰

## 실행
```
go test -run TestDCI -v -count=1
```

## SSOT
governance_consts.go → 유일한 값 정의처
코드 → const 참조만 허용
테스트 → const/패턴 자동 검증

## 숫자 체계 (검증 완료)
| 기준 | 설명 |
|------|------|
| GEMINI header (FiredNeurons) | active만 (counter+dopa>0) |
| 영역 요약 | active만, brainstem 제외 |
| brain 명령 | 모든 non-dormant 폴더 (구조 포함) |

이 차이는 **의도적 설계** — DCI로 검증 불필요.