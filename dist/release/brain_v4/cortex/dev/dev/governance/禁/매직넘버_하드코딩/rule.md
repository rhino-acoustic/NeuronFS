---
description: "코드에 0.6, 10, 3 등 매직넘버를 직접 쓰면 DCI-08 테스트가 실패. governance_consts.go 참조 필수."
---
# 禁 매직넘버 하드코딩

neuron_crud.go, lifecycle.go, brain.go, emit_bootstrap.go에
governance_consts.go 없이 숫자를 직접 쓰면 go test 실패.

DCI-08 검증 파일 목록은 governance_benchmark_test.go 참조.
