@echo off
REM ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
REM NeuronFS Gemini CLI 자율 진화 호출기
REM 사용: gemini_evolve.bat [프롬프트 파일 경로]
REM 기본: brain_v4\_inbox\master_prompt.md
REM ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

set PROMPT_FILE=%~1
if "%PROMPT_FILE%"=="" set PROMPT_FILE=%~dp0brain_v4\_inbox\master_prompt.md

REM 타임스탬프
for /f "tokens=1-4 delims=/ " %%a in ('date /t') do set DATE=%%a%%b%%c
for /f "tokens=1-2 delims=: " %%a in ('time /t') do set TIME=%%a%%b
set TS=%DATE%_%TIME%

echo [%TS%] Gemini CLI 자율 진화 시작
echo   프롬프트: %PROMPT_FILE%

REM MCP 서버 실행 확인
curl -s http://127.0.0.1:9247/api/status >nul 2>&1
if errorlevel 1 (
    echo [ERROR] NeuronFS MCP 서버 미실행. start.bat 먼저 실행하세요.
    exit /b 1
)

REM Gemini CLI 호출 (sandbox=none으로 파일접근 허용)
if exist "%PROMPT_FILE%" (
    echo [%TS%] 프롬프트 파일 읽기...
    gemini -p "$(type %PROMPT_FILE%)" --sandbox=none
) else (
    echo [%TS%] 기본 마스터 프롬프트 실행...
    gemini -p "[NeuronFS 자율 진화] status 확인 후 corrections.jsonl 분석, 개선점 뉴런 각인, health_check 실행" --sandbox=none
)

echo [%TS%] 완료
