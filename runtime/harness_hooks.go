// harness_hooks.go — NeuronFS Hook 인프라 자동 세팅
// PROVIDES: ensureHooksInfra, writeHookFile
// 시작 시 .gemini/hooks/ 디렉토리와 settings.json을 자동 생성/갱신
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ── Hook 스크립트 정의 ──

var hookScripts = map[string]string{
	"pre_edit_git.ps1": `#!/usr/bin/env pwsh
# NeuronFS Hook: 파일 수정 전 자동 git snapshot
$ErrorActionPreference = "SilentlyContinue"
$input_json = [Console]::In.ReadToEnd()
Set-Location "$env:USERPROFILE\NeuronFS"
$status = git status --porcelain 2>&1
if ($status) {
    git add -A 2>$null
    $ts = (Get-Date).ToString("HH:mm:ss")
    git commit -m "[hook] pre-edit snapshot $ts" --allow-empty 2>$null
    [Console]::Error.WriteLine("[HOOK] git snapshot at $ts")
}
Write-Output '{"decision":"allow"}'
exit 0
`,

	"block_delete.ps1": `#!/usr/bin/env pwsh
# NeuronFS Hook: brain_v4 삭제 차단 → _quarantine 격리 강제
$ErrorActionPreference = "SilentlyContinue"
$input_json = [Console]::In.ReadToEnd()
if ($input_json -match "Remove-Item|del |rm |rmdir|Delete") {
    if ($input_json -match "brain_v4|NeuronFS") {
        [Console]::Error.WriteLine("[HOOK] BLOCKED: brain_v4 직접 삭제 금지")
        Write-Output '{"decision":"block","reason":"brain_v4 파일 직접 삭제 금지. _quarantine으로 이동하세요."}'
        exit 2
    }
}
Write-Output '{"decision":"allow"}'
exit 0
`,

	"go_vet_guard.ps1": `#!/usr/bin/env pwsh
# NeuronFS Hook: go vet 실패 시 커밋 차단
$ErrorActionPreference = "SilentlyContinue"
$input_json = [Console]::In.ReadToEnd()
if ($input_json -match "git commit" -and $input_json -match "runtime") {
    Set-Location "$env:USERPROFILE\NeuronFS"
    $vet = go vet ./runtime/... 2>&1
    if ($LASTEXITCODE -ne 0) {
        [Console]::Error.WriteLine("[HOOK] BLOCKED: go vet failed")
        [Console]::Error.WriteLine($vet)
        Write-Output '{"decision":"block","reason":"go vet failed"}'
        exit 2
    }
    [Console]::Error.WriteLine("[HOOK] go vet PASS")
}
Write-Output '{"decision":"allow"}'
exit 0
`,

	"encoding_guard.ps1": `#!/usr/bin/env pwsh
# NeuronFS Hook: 한글 인코딩 안전장치
$ErrorActionPreference = "SilentlyContinue"
$input_json = [Console]::In.ReadToEnd()
if ($input_json -match "Get-Content" -and $input_json -notmatch "Encoding") {
    [Console]::Error.WriteLine("[HOOK] WARNING: Get-Content without -Encoding → ReadAllText 권장")
}
if ($input_json -match "Set-Content" -and $input_json -notmatch "WriteAllText") {
    [Console]::Error.WriteLine("[HOOK] WARNING: Set-Content → WriteAllText 권장 (BOM 방지)")
}
Write-Output '{"decision":"allow"}'
exit 0
`,

	"codemap_sync.ps1": `#!/usr/bin/env pwsh
# NeuronFS Hook: AfterTool — .go 파일 수정 후 코드맵 동기화
$ErrorActionPreference = "SilentlyContinue"
$input_json = [Console]::In.ReadToEnd()
if ($input_json -match "\.go") {
    Set-Location "$env:USERPROFILE\NeuronFS"
    Start-Job -ScriptBlock {
        Set-Location "$env:USERPROFILE\NeuronFS"
        $files = Get-ChildItem runtime -Filter "*.go" | Sort-Object Name
        $mapDir = "$env:USERPROFILE\NeuronFS\brain_v4\cortex\dev\_codemap"
        if (!(Test-Path $mapDir)) { New-Item -ItemType Directory -Path $mapDir -Force | Out-Null }
        $lines = @("---", "type: codemap", ("updated: " + (Get-Date -Format "yyyy-MM-dd HH:mm")), "---")
        foreach ($f in $files) {
            $cnt = (Select-String "^func " $f.FullName).Count
            $lines += ($f.Name + ": " + $cnt + " funcs")
        }
        [System.IO.File]::WriteAllText(($mapDir + "\rule.md"), ($lines -join [Environment]::NewLine), [System.Text.Encoding]::UTF8)
    } | Out-Null
    [Console]::Error.WriteLine("[HOOK] codemap sync triggered")
}
Write-Output '{"decision":"allow"}'
exit 0
`,

	"session_start.ps1": `#!/usr/bin/env pwsh
$ErrorActionPreference = "SilentlyContinue"
$input_json = [Console]::In.ReadToEnd()
Set-Location "$env:USERPROFILE\NeuronFS"
$health = Invoke-RestMethod -Uri "http://127.0.0.1:9090/api/ping" -TimeoutSec 3 -ErrorAction SilentlyContinue
if ($health.status -eq "ok") {
    Invoke-RestMethod -Uri "http://127.0.0.1:9090/api/inject" -Method POST -TimeoutSec 10 -ErrorAction SilentlyContinue | Out-Null
    [Console]::Error.WriteLine("[HOOK] SessionStart: neuron inject OK")
}
$lastCommit = git log -1 --format="%h %s" 2>$null
[Console]::Error.WriteLine("[HOOK] SessionStart: $lastCommit")
Write-Output '{"decision":"allow"}'
exit 0
`,

	"session_end.ps1": `#!/usr/bin/env pwsh
$ErrorActionPreference = "SilentlyContinue"
$input_json = [Console]::In.ReadToEnd()
Set-Location "$env:USERPROFILE\NeuronFS"
$status = git status --porcelain 2>&1
if ($status) {
    git add -A 2>$null
    git commit -m "[session-end] auto-save" --allow-empty 2>$null
    [Console]::Error.WriteLine("[HOOK] SessionEnd: auto-commit")
}
Write-Output '{"decision":"allow"}'
exit 0
`,
}

// ── settings.json Hook 설정 ──

type hookEntry struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Command string `json:"command"`
	Timeout int    `json:"timeout"`
}

type matcherEntry struct {
	Matcher string      `json:"matcher"`
	Hooks   []hookEntry `json:"hooks"`
}

type hooksConfig struct {
	SessionStart []matcherEntry `json:"SessionStart,omitempty"`
	SessionEnd   []matcherEntry `json:"SessionEnd,omitempty"`
	BeforeTool   []matcherEntry `json:"BeforeTool,omitempty"`
	AfterTool    []matcherEntry `json:"AfterTool,omitempty"`
}

type settingsJSON struct {
	MCPServers    map[string]interface{} `json:"mcpServers"`
	FileFiltering map[string]interface{} `json:"fileFiltering"`
	Hooks         hooksConfig            `json:"hooks"`
}

func ensureHooksInfra(nfsRoot string) {
	geminiDir := filepath.Join(nfsRoot, ".gemini")
	hooksDir := filepath.Join(geminiDir, "hooks")
	os.MkdirAll(hooksDir, 0750)

	// 1. Hook 스크립트 생성
	for name, content := range hookScripts {
		path := filepath.Join(hooksDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			fmt.Printf("[HARNESS] ⚠️ Hook 생성 실패: %s — %v\n", name, err)
		}
	}

	// 2. settings.json 갱신 (기존 MCP 설정 보존)
	settingsPath := filepath.Join(geminiDir, "settings.json")
	settings := settingsJSON{
		MCPServers: map[string]interface{}{
			"neuronfs": map[string]interface{}{
				"httpUrl": fmt.Sprintf("http://127.0.0.1:%d/mcp", MCPStreamPort),
			},
		},
		FileFiltering: map[string]interface{}{
			"respectGitIgnore": false,
		},
		Hooks: hooksConfig{
			SessionStart: []matcherEntry{
				{
					Matcher: "*",
					Hooks: []hookEntry{
						{Name: "neuron-load", Type: "command", Command: "powershell -ExecutionPolicy Bypass -File .gemini/hooks/session_start.ps1", Timeout: 15000},
					},
				},
			},
			SessionEnd: []matcherEntry{
				{
					Matcher: "*",
					Hooks: []hookEntry{
						{Name: "auto-save", Type: "command", Command: "powershell -ExecutionPolicy Bypass -File .gemini/hooks/session_end.ps1", Timeout: 10000},
					},
				},
			},
			BeforeTool: []matcherEntry{
				{
					Matcher: "replace_file_content|write_to_file|multi_replace_file_content",
					Hooks: []hookEntry{
						{Name: "git-snapshot", Type: "command", Command: "powershell -ExecutionPolicy Bypass -File .gemini/hooks/pre_edit_git.ps1", Timeout: 10000},
					},
				},
				{
					Matcher: "run_command",
					Hooks: []hookEntry{
						{Name: "block-delete", Type: "command", Command: "powershell -ExecutionPolicy Bypass -File .gemini/hooks/block_delete.ps1", Timeout: 5000},
						{Name: "go-vet-guard", Type: "command", Command: "powershell -ExecutionPolicy Bypass -File .gemini/hooks/go_vet_guard.ps1", Timeout: 30000},
						{Name: "encoding-guard", Type: "command", Command: "powershell -ExecutionPolicy Bypass -File .gemini/hooks/encoding_guard.ps1", Timeout: 3000},
					},
				},
			},
			AfterTool: []matcherEntry{
				{
					Matcher: "replace_file_content|write_to_file|multi_replace_file_content",
					Hooks: []hookEntry{
						{Name: "codemap-sync", Type: "command", Command: "powershell -ExecutionPolicy Bypass -File .gemini/hooks/codemap_sync.ps1", Timeout: 5000},
					},
				},
			},
		},
	}

	data, _ := json.MarshalIndent(settings, "", "  ")
	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		fmt.Printf("[HARNESS] ⚠️ settings.json 갱신 실패: %v\n", err)
		return
	}
	fmt.Printf("[HARNESS] ✅ %d hooks 설정 완료\n", len(hookScripts))
}
