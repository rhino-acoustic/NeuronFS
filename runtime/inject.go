package main

// ━━━ inject.go ━━━
// PROVIDES: markBrainDirty, consumeDirty, computeMountHash, autoReinject, processInbox, runInjectionLoop
// DEPENDS ON: brain.go, emit.go, lifecycle.go, mcp_server.go (notifyMCPResourceUpdated)

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// DIRTY FLAG + BATCH INJECTION
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
var (
	brainDirty    bool
	brainDirtyMu  sync.Mutex
	lastMountHash string
	triggerChan   = make(chan struct{}, 1)
)

// markBrainDirty signals that the brain state has changed and needs an event broadcast.
func markBrainDirty() {
	brainDirtyMu.Lock()
	brainDirty = true
	brainDirtyMu.Unlock()

	// Non-blocking trigger
	select {
	case triggerChan <- struct{}{}:
	default:
	}

	// MCP: 리소스 변경 알림 → IDE가 neuronfs://rules/current 재읽기
	go notifyMCPResourceUpdated()
}

// consumeDirty checks and clears the brain's dirty state flag.
func consumeDirty() bool {
	brainDirtyMu.Lock()
	defer brainDirtyMu.Unlock()
	if brainDirty {
		brainDirty = false
		return true
	}
	return false
}

// computeMountHash returns a hash of the current mount set (neuron IDs + counters that are mounted)
func computeMountHash(brainRoot string) string {
	brain := scanBrain(brainRoot)
	result := runSubsumption(brain)
	var parts []string
	for _, region := range result.ActiveRegions {
		for _, n := range region.Neurons {
			if n.IsDormant {
				continue
			}
			// 폴더 존재 = 활성. 모든 비dormant 뉴런을 마운트 해시에 포함.
			parts = append(parts, fmt.Sprintf("%s:%d", n.Path, n.Counter))
		}
	}
	sort.Strings(parts)
	return fmt.Sprintf("%x", len(parts)) + "|" + strings.Join(parts, ",")
}

// autoReinject checks mount set hash and only writes if changed
func autoReinject(brainRoot string) {
	newHash := computeMountHash(brainRoot)
	if newHash == lastMountHash {
		return // mount set unchanged, skip injection
	}
	lastMountHash = newHash
	writeAllTiers(brainRoot)
	fmt.Printf("[INJECT] ♻️  Mount set changed → GEMINI.md updated\n")
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// INBOX PROCESSOR: AI tool call → _inbox → neurons
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// inboxEntry represents a correction or insight from AI or auto-accept
type inboxEntry struct {
	Ts         string `json:"ts"`
	Type       string `json:"type"` // "correction" | "insight"
	Text       string `json:"text"`
	Source     string `json:"source"`      // "ai" | "auto-accept"
	Path       string `json:"path"`        // optional: pre-computed neuron path
	CounterAdd int    `json:"counter_add"` // optional: how much to add
	Author     string `json:"author"`      // optional: explicit author mapping
}

// processInbox reads _inbox/corrections.jsonl, creates/fires neurons, then clears
func processInbox(brainRoot string) {
	inboxPath := filepath.Join(brainRoot, "_inbox", "corrections.jsonl")
	data, err := os.ReadFile(inboxPath)
	if err != nil {
		return // no inbox file = nothing to process
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return
	}

	processed := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var entry inboxEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			fmt.Printf("[INBOX] ⚠️ parse error: %s\n", line)
			continue
		}

		// Determine neuron path
		neuronPath := entry.Path
		if neuronPath == "" {
			// Auto-generate path from text if not provided
			// Simple heuristic: cortex/_inbox_pending/<sanitized_text>
			sanitized := strings.Map(func(r rune) rune {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
					r == '_' || r == '-' ||
					(r >= 0xAC00 && r <= 0xD7AF) || // 한글 음절
					(r >= 0x3131 && r <= 0x318E) || // 한글 자모
					(r >= 0x4E00 && r <= 0x9FFF) { // 한자 CJK
					return r
				}
				return '_'
			}, strings.ReplaceAll(entry.Text, " ", "_"))
			if len(sanitized) > 60 {
				sanitized = sanitized[:60]
			}
			neuronPath = "hippocampus/_inbox_pending/" + sanitized
		}

		// Security: Basic Prompt Injection & Path Traversal Defense
		if strings.Contains(neuronPath, "..") || strings.Contains(neuronPath, `\`) ||
			strings.Contains(neuronPath, "$") || strings.Contains(neuronPath, "&") ||
			strings.Contains(neuronPath, "|") || strings.Contains(neuronPath, ">") {
			fmt.Printf("[SECURITY] 🛡️ Injection blocked: %s\n", neuronPath)
			continue
		}

		// 기계적 칭찬(Dopamine Inflation) 필터링
		isPraise := false
		if entry.Type == "correction" && entry.Text == "PD칭찬" {
			isPraise = true
		}
		praiseRegex := regexp.MustCompile(`(?i)(칭찬|잘\s*쓰셨습니다|좋아|훌륭|완벽|최고)`)
		if praiseRegex.MatchString(entry.Text) || strings.Contains(strings.ToLower(neuronPath), "dopamine") {
			isPraise = true
		}

		if isPraise {
			authorId := entry.Author
			if authorId == "" {
				authorId = entry.Source
			}
			authorId = strings.ToLower(authorId)

			if authorId != "pm" && authorId != "basement_admin" && !strings.Contains(authorId, "pd") {
				fmt.Printf("[INBOX] 🛡️ 도파민 인플레이션 차단 (침해자: %s): %s\n", authorId, entry.Text)
				continue
			}
			// PM 칭찬은 바로 도파민 발화
			fullPath := filepath.Join(brainRoot, strings.ReplaceAll(neuronPath, "/", string(filepath.Separator)))
			_ = os.MkdirAll(fullPath, 0750)
			_ = signalNeuron(brainRoot, neuronPath, "dopamine")
			fmt.Printf("[INBOX] 🟢 PM 칭찬 확인 — 도파민 배포: %s\n", neuronPath)
			processed++
			continue
		}

		// Determine action
		counterAdd := entry.CounterAdd
		if counterAdd <= 0 {
			counterAdd = 1
		}

		// Check if neuron exists → fire, else → grow
		fullPath := filepath.Join(brainRoot, strings.ReplaceAll(neuronPath, "/", string(filepath.Separator)))
		if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
			// Exists → fire N times
			for i := 0; i < counterAdd; i++ {
				fireNeuron(brainRoot, neuronPath)
			}
			fmt.Printf("[INBOX] 🔥 fire %s (×%d)\n", neuronPath, counterAdd)
		} else {
			// New → grow
			if err := growNeuron(brainRoot, neuronPath); err != nil {
				fmt.Printf("[INBOX] ❌ grow failed: %s — %v\n", neuronPath, err)
				continue
			}
			// Fire additional times if counter_add > 1
			for i := 1; i < counterAdd; i++ {
				fireNeuron(brainRoot, neuronPath)
			}
			fmt.Printf("[INBOX] 🌱 grow %s (counter=%d)\n", neuronPath, counterAdd)
		}
		processed++
	}

	if processed > 0 {
		// Append to persistent history before clearing (for --neuronize)
		historyPath := filepath.Join(brainRoot, "_inbox", "corrections_history.jsonl")
		f, err := os.OpenFile(historyPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err == nil {
			f.Write(data)
			f.Close()
		}
		// Clear inbox
		os.WriteFile(inboxPath, []byte{}, 0600)
		markBrainDirty()
		fmt.Printf("[INBOX] ✅ %d entries processed, inbox cleared (history preserved)\n", processed)
	}
}

// runInjectionLoop uses fsnotify and channels for event-driven, real-time updates
func runInjectionLoop(brainRoot string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("[ERROR] fsnotify: %v\n", err)
		return
	}
	defer watcher.Close()

	inboxDir := filepath.Join(brainRoot, "_inbox")
	os.MkdirAll(inboxDir, 0750)
	if err := watcher.Add(inboxDir); err != nil {
		fmt.Printf("[ERROR] fsnotify watch _inbox: %v\n", err)
	}

	debounceDuration := 100 * time.Millisecond
	var timer *time.Timer

	executeUpdate := func() {
		// [Hot Reload] Background Delegate (Delegate Logic to Worker)
		nfsExe, _ := os.Executable()
		// _inbox changes detect -> Spawn stateless worker to process them.
		cmd := exec.Command(nfsExe, brainRoot, "--tool", "inject_tick", "{}")
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] worker 'inject_tick' crashed: %v\nOutput: %s\n", err, string(out))
		} else {
			if len(strings.TrimSpace(string(out))) > 0 {
				fmt.Print(string(out))
			}
		}
	}

	queueUpdate := func() {
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(debounceDuration, executeUpdate)
	}

	// Initial check
	queueUpdate()

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				fmt.Fprintf(os.Stderr, "\033[33m[PULSE] %s evolved. (27ms)\033[0m\n", filepath.Base(event.Name))
				queueUpdate()
			} else if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
				fmt.Fprintf(os.Stderr, "\033[90m[PRUNE] 데드 시냅스 제거 완료: %s\033[0m\n", filepath.Base(event.Name))
				queueUpdate()
			}
		case <-triggerChan:
			queueUpdate()
		case <-time.After(5 * time.Minute):
			// 주기적 폴백 확인 (5분)
			queueUpdate()
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Fprintf(os.Stderr, "[ERROR] watcher: %v\n", err)
		}
	}
}

// gitSnapshot takes a single git snapshot of the brain state
// Called only during idle via --snapshot flag (not on every fire/grow)
// Lifecycle: active → changes accumulate → idle → groq analysis → snapshot
