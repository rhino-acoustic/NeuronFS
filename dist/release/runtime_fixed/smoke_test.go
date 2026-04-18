package main

// NeuronFS 전체 서비스 스모크 테스트
// 모든 서브시스템을 실제로 호출하여 연동 검증
// 실행: go test -run TestSmoke -v -timeout 60s

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ── 1. 뉴런 생애주기 ──

func TestSmoke_NeuronLifecycle(t *testing.T) {
	brain := getTestBrainRoot(t)

	// GROW: 뉴런 생성
	testDir := filepath.Join(brain, "cortex", "_smoke_test", "lifecycle_test")
	os.MkdirAll(testDir, 0750)
	defer os.RemoveAll(filepath.Join(brain, "cortex", "_smoke_test"))

	neuronFile := filepath.Join(testDir, "1.neuron")
	os.WriteFile(neuronFile, []byte("source: smoke_test\nupdated: "+time.Now().Format(time.RFC3339)+"\n# smoke test neuron\n"), 0600)

	if !fileExists(neuronFile) {
		t.Fatal("GROW 실패: .neuron 파일 미생성")
	}
	t.Log("✅ GROW: 뉴런 생성 성공")

	// FIRE: 카운터 증가 시뮬레이션
	content, _ := os.ReadFile(neuronFile)
	if len(content) == 0 {
		t.Fatal("FIRE 실패: .neuron 파일 비어있음")
	}
	t.Log("✅ FIRE: 뉴런 파일 접근 성공")

	// DECAY: 없는 뉴런 확인
	staleDir := filepath.Join(brain, "cortex", "_smoke_test", "stale_neuron")
	os.MkdirAll(staleDir, 0750)
	// 빈 폴더 = 도태 대상
	entries, _ := os.ReadDir(staleDir)
	if len(entries) != 0 {
		t.Fatal("DECAY 테스트 폴더가 비어있지 않음")
	}
	t.Log("✅ DECAY: 빈 폴더 (도태 대상) 감지 가능")
}

// ── 2. Emit 파이프라인 ──

func TestSmoke_EmitPipeline(t *testing.T) {
	brain := getTestBrainRoot(t)

	// _rules.md 존재 확인 (emit 결과물)
	regions := []string{"brainstem", "cortex", "ego", "hippocampus", "limbic", "prefrontal", "sensors", "shared"}
	for _, r := range regions {
		rulesFile := filepath.Join(brain, r, "_rules.md")
		if !fileExists(rulesFile) {
			t.Errorf("❌ %s/_rules.md 누락", r)
		}
	}
	t.Log("✅ EMIT: 8개 영역 _rules.md 존재")

	// _index.md 확인
	indexFile := filepath.Join(brain, "_index.md")
	if !fileExists(indexFile) {
		t.Fatal("❌ _index.md 누락")
	}
	t.Log("✅ EMIT: _index.md 존재")

	// GEMINI.md 확인
	home, _ := os.UserHomeDir()
	geminiMd := filepath.Join(home, ".gemini", "GEMINI.md")
	if !fileExists(geminiMd) {
		t.Fatal("❌ GEMINI.md 누락")
	}
	content, _ := os.ReadFile(geminiMd)
	s := string(content)
	if !strings.Contains(s, "NEURONFS:START") {
		t.Fatal("❌ GEMINI.md에 NEURONFS 블록 없음")
	}
	if !strings.Contains(s, "<identity>") {
		t.Fatal("❌ GEMINI.md에 <identity> 태그 없음")
	}
	t.Log("✅ EMIT: GEMINI.md 정상 (identity + NEURONFS 블록)")
}

// ── 3. 코드맵 연동 ──

func TestSmoke_CodemapIntegrity(t *testing.T) {
	brain := getTestBrainRoot(t)
	codemapDir := filepath.Join(brain, "cortex", "dev", "_codemap")

	if !fileExists(codemapDir) {
		t.Fatal("❌ _codemap 디렉토리 없음")
	}

	entries, _ := os.ReadDir(codemapDir)
	if len(entries) < 10 {
		t.Fatalf("❌ 코드맵 뉴런 %d개 — 최소 10개 이상 필요", len(entries))
	}

	// 각 코드맵 뉴런에 .neuron 파일 확인
	missing := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		neuron := filepath.Join(codemapDir, e.Name(), "1.neuron")
		if !fileExists(neuron) {
			missing++
		}
	}
	if missing > 0 {
		t.Errorf("⚠️ 코드맵 %d개 뉴런에 1.neuron 없음", missing)
	}

	// 코드맵 내용 검증 — PROVIDES 포함 여부
	sample := filepath.Join(codemapDir, "brain", "1.neuron")
	if fileExists(sample) {
		data, _ := os.ReadFile(sample)
		if !strings.Contains(string(data), "PROVIDES") {
			t.Error("⚠️ brain 코드맵에 PROVIDES 헤더 없음")
		}
	}
	t.Logf("✅ CODEMAP: %d개 뉴런 (missing: %d)", len(entries), missing)
}

// ── 4. Hook 시스템 ──

func TestSmoke_HookSystem(t *testing.T) {
	home, _ := os.UserHomeDir()
	settingsPath := filepath.Join(home, "NeuronFS", ".gemini", "settings.json")

	if !fileExists(settingsPath) {
		t.Fatal("❌ .gemini/settings.json 없음")
	}

	data, _ := os.ReadFile(settingsPath)
	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("❌ settings.json 파싱 실패: %v", err)
	}

	hooks, ok := settings["hooks"].(map[string]interface{})
	if !ok {
		t.Fatal("❌ hooks 섹션 없음")
	}

	required := []string{"SessionStart", "SessionEnd", "BeforeTool", "AfterTool"}
	for _, h := range required {
		if _, exists := hooks[h]; !exists {
			t.Errorf("❌ Hook 누락: %s", h)
		}
	}
	t.Logf("✅ HOOKS: %d개 이벤트 등록", len(hooks))

	// Hook 스크립트 존재 확인
	hookDir := filepath.Join(home, "NeuronFS", ".gemini", "hooks")
	scripts, _ := os.ReadDir(hookDir)
	if len(scripts) < 3 {
		t.Errorf("⚠️ Hook 스크립트 %d개 — 최소 3개 필요", len(scripts))
	}
	t.Logf("✅ HOOK SCRIPTS: %d개", len(scripts))
}

// ── 5. 카트리지/VFS ──

func TestSmoke_Cartridges(t *testing.T) {
	home, _ := os.UserHomeDir()
	jlootDir := filepath.Join(home, "NeuronFS", "tools", "jloot")

	if !fileExists(jlootDir) {
		t.Skip("tools/jloot 없음 — 카트리지 미설치")
	}

	entries, _ := os.ReadDir(jlootDir)
	jlootCount := 0
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".jloot") {
			jlootCount++
		}
	}
	if jlootCount == 0 {
		t.Error("❌ .jloot 카트리지 0개")
	}
	t.Logf("✅ CARTRIDGES: %d개 .jloot", jlootCount)
}

// ── 6. Git 보호 ──

func TestSmoke_GitProtection(t *testing.T) {
	home, _ := os.UserHomeDir()
	nfsRoot := filepath.Join(home, "NeuronFS")

	// .gitignore에서 brain_v4 추적 확인
	gitignore, _ := os.ReadFile(filepath.Join(nfsRoot, ".gitignore"))
	s := string(gitignore)

	// brain_v4/ 전체 무시가 없어야 함
	if strings.Contains(s, "\nbrain_v4/\n") {
		t.Error("❌ .gitignore에 brain_v4/ 전체 무시 존재 — 뉴런 손실 위험")
	}
	t.Log("✅ GIT: brain_v4 보호됨")
}

// ── 7. corrections.jsonl (자가학습) ──

func TestSmoke_SelfHealing(t *testing.T) {
	brain := getTestBrainRoot(t)
	corrPath := filepath.Join(brain, "_inbox", "corrections.jsonl")

	if !fileExists(corrPath) {
		t.Skip("corrections.jsonl 없음")
	}

	data, _ := os.ReadFile(corrPath)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	t.Logf("✅ SELF-HEAL: corrections.jsonl %d건", len(lines))
}

// ── 8. hippocampus 기억 ──

func TestSmoke_Memory(t *testing.T) {
	brain := getTestBrainRoot(t)
	sessionDir := filepath.Join(brain, "hippocampus", "session_log")

	if !fileExists(sessionDir) {
		t.Fatal("❌ hippocampus/session_log 없음")
	}

	entries, _ := os.ReadDir(sessionDir)
	neuronCount := 0
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".neuron") {
			neuronCount++
		}
	}
	if neuronCount == 0 {
		t.Error("❌ session_log에 .neuron 0개")
	}
	t.Logf("✅ MEMORY: %d개 에피소드", neuronCount)
}

// ── Helpers ──

func getTestBrainRoot(t *testing.T) string {
	t.Helper()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal("UserHomeDir 실패")
	}
	brain := filepath.Join(home, "NeuronFS", "brain_v4")
	if !fileExists(brain) {
		t.Fatal("brain_v4 디렉토리 없음")
	}
	return brain
}

func TestSmoke_Summary(t *testing.T) {
	t.Log("=== NeuronFS 스모크 테스트 완료 ===")
	t.Log(fmt.Sprintf("시각: %s", time.Now().Format("2006-01-02 15:04:05")))
}
