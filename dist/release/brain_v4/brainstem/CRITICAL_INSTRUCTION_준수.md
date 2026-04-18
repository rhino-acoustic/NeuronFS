# CRITICAL_INSTRUCTION_준수
(시스템 플랫폼 제약 사항)

## 금지된 터미널 명령
1. **절대 금지**: `bash/powershell` 커맨드 내에서 `cat`을 사용하여 파일을 생성하거나 덧붙이지 마라. (대신 `write_to_file` 혹은 `replace_file_content` 툴을 써라)
2. **절대 금지**: 터미널 내 접근으로 `grep`을 실행하지 마라. (대신 `grep_search` 도구를 사용하라)
3. **절대 금지**: 열람, 리스팅, 대체를 위해 `ls`, `cat`, `sed` 를 터미널에서 수행하지 마라. (`view_file`, `list_dir`, `replace_file_content` 사용 권장)

기본적으로 "가장 특화된 전문 도구(Tool)"가 존재한다면, 쉘명령어(run_command)보다 그걸 우선 호출해야 토큰 효율성과 파싱 에러를 유발하지 않는다.
이는 하네스 루프상에서 내가 스스로 짠 스크립트가 샌드박스의 방어 로직에 막혀 끊임없이 통과 실패(Failed)하는 반복적인 비효율성을 근원적으로 제거하기 위한 각인(Imprinting)이다.
