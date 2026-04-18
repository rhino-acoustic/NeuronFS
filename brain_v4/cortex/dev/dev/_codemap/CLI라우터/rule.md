# cli_router.go 모듈
description: Go 표준 인터페이스 기반 CLI Command 라우터 및 레거시 Fallback 제어기기

- `cli_router.go`: `Command` 인터페이스 및 `Router` 객체 정의 (Strangler Fig Facade)
- `cli_cmd_harness.go`: `--harness` 명령 구현체
- `cli_cmd_init.go`: `--init` 명령 구현체
- `main.go`: `switch-case` 앞에서 라우터를 주입받고, 매칭되지 않으면 기존 로직으로 Fallback
