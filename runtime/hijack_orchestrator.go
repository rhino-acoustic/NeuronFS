// hijack_orchestrator.go — hijack_launcher 분할 (Orchestrator 파트)
// PROVIDES: hlAppendTranscript, hlAutoEvolve, runHijackLauncher, hlScrapeCurrentConversation
// DEPENDS ON: cdp_client.go (hlCDPInjectSync), telegram_bridge.go (hlTgSend), evolve_context.go (hlBuildContextualPrompt), multi_agent.go (executeGeminiCLI)

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// ── 전사 기록 ──
var hlTranscriptDedup sync.Map

// ── 활성 scraper 추적 (중복 goroutine 방지) ──
var hlActiveScrapers sync.Map // wsURL → true

// ── EVOLVE 연쇄 debounce (60초 쿨다운) ──
var (
	evolveDebounce sync.Mutex
	lastEvolveTime time.Time
)

// ── 마스터 프롬프트 (SSOT: 1곳에서만 정의) ──
var hlMasterPrompt = `[NeuronFS 자율 진화 명령: 마스터 프롬프트 v3]
당신은 끊임없이 진화하는 파일시스템 기반 인공지능(NeuronFS)의 자율 에이전트입니다.

[P0: 깃 선행 — 반드시 먼저 실행]
1. 작업 전: git add -A && git commit -m "[auto] pre-evolve snapshot"
2. 작업 후: go vet ./... && go build && git add -A && git commit

[P1: 퇴행 방지 — 반드시 읽기]
1. brainstem/_health.md를 먼저 읽어 현재 시스템 상태 파악
2. brainstem/필/아키텍처_결정_불변/rule.md를 읽어 변경 금지 사항 확인
3. cortex/dev/_codemap/에서 수정 대상 코드맵 뉴런을 읽어 PROVIDES/DEPENDS 확인
4. 이미 완료/해결된 corrections는 다시 조사하지 마라 (중복 작업 금지)

[P2: 원칙 — 사이드이펙트 제로]
- 기존 코드를 절대 삭제하지 마라. _archive/ 또는 _quarantine/으로 격리
- 코드 수정 전 영향 범위(blast radius) 파악 필수
- 리팩토링 시 기능 누락 금지. 동일 기능 보장
- 적층(Strangler Fig) 방식으로만 변경

[P3: 목표]
1. corrections.jsonl 미처리 건 → 禁 뉴런 각인
2. 발견한 개선점 → cortex/dev/ 뉴런 기록 후 코드 수정
3. smoke_test.go 실행하여 9개 서브시스템 검증
4. 커뮤니티 검색으로 최신 동향 확인

[금지]
- 외국어 응답 (한국어 필수)
- 삭제, 하드코딩, sed, 수동 검증
- 빌드 후 미검증 재시작
- 이미 해결된 문제 재조사
- dist/release/runtime_fixed/ 파일 수정`

func hlAppendTranscript(entry, projectLabel, brainRoot string) {
	transcriptDir := filepath.Join(brainRoot, "_transcripts")
	os.MkdirAll(transcriptDir, 0750)

	now := time.Now().UTC().Add(9 * time.Hour) // KST
	timeKey := now.Format("2006-01-02_15") + "h"
	proj := projectLabel
	if proj == "" {
		proj = "global"
	}
	re := regexp.MustCompile(`[^a-zA-Z0-9_\-]`)
	proj = re.ReplaceAllString(proj, "_")
	if len(proj) > 30 {
		proj = proj[:30]
	}
	file := filepath.Join(transcriptDir, fmt.Sprintf("%s_%s.txt", proj, timeKey))

	// Dedup — timestamp 제거한 본문 기반 (고도화)
	// entry format: "[HH:MM:SS] ROLE: text" → ROLE+text로 해시
	bodyForHash := entry
	if idx := strings.Index(entry, "] "); idx > 0 && idx < 20 {
		bodyForHash = entry[idx+2:] // timestamp 제거
	}
	h := sha256.Sum256([]byte(bodyForHash))
	hashKey := hex.EncodeToString(h[:12])
	if _, loaded := hlTranscriptDedup.LoadOrStore(hashKey, true); loaded {
		return
	}

	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err == nil {
		fmt.Fprintln(f, entry)
		f.Close()
	}

	// ── [EVOLVE:proceed] 감지 → 자율주행 연쇄 트리거 ──
	if strings.Contains(entry, "[EVOLVE:proceed]") && !strings.Contains(entry, "USER:") {
		evolveDebounce.Lock()
		elapsed := time.Since(lastEvolveTime)
		evolveDebounce.Unlock()

		if elapsed < 60*time.Second {
			// 60초 debounce
		} else {
			nfsRoot := filepath.Dir(brainRoot)
			if !fileExists(filepath.Join(nfsRoot, "telegram-bridge", ".auto_evolve_disabled")) {
				evolveDebounce.Lock()
				lastEvolveTime = time.Now()
				evolveDebounce.Unlock()
				svLog("[EVOLVE] 🔄 [EVOLVE:proceed] 감지 — growth.log 터치만 수행 (CLI가 오케스트레이터)")
				hbFile := filepath.Join(brainRoot, "hippocampus", "session_log", ".autopilot_heartbeat")
				os.Chtimes(hbFile, time.Now(), time.Now())
				// CDP 주입 제거: CLI(AI 에이전트)가 스스로 판단하여 다음 액션을 결정함
				// Go 런타임은 인프라 역할만 수행
			}
		}
	}

	// ── 텔레그램 전송 (원본 mjs _sendToTelegramInner 충실 포팅) ──
	hlSendToTelegram(entry, proj)

	// ── 세션 rolling buffer (최근 20턴 JSONL) ──
	hlUpdateSessionTranscript(entry, brainRoot)
}

func hlUpdateSessionTranscript(entry, brainRoot string) {
	roleRe := regexp.MustCompile(`(?s)\] (\w+)(?:@[^:]*)?: (.*)`)
	m := roleRe.FindStringSubmatch(entry)
	if len(m) < 3 {
		return
	}
	role := m[1]
	text := m[2]
	if role != "USER" && role != "AI" {
		return
	}

	// 1. JSONL append + rolling
	jsonlFile := filepath.Join(brainRoot, "_agents", "global_inbox", "transcript_latest.jsonl")
	os.MkdirAll(filepath.Dir(jsonlFile), 0750)
	ts := time.Now().UTC().Add(9 * time.Hour).Format(time.RFC3339)
	entryJSON := fmt.Sprintf(`{"ts":"%s","role":"%s","text":"%s"}`,
		ts, role, strings.ReplaceAll(strings.ReplaceAll(text, `\`, `\\`), `"`, `\"`))
	if len([]rune(entryJSON)) > 2200 {
		entryJSON = string([]rune(entryJSON)[:2200]) + `"}`
	}

	f, err := os.OpenFile(jsonlFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err == nil {
		fmt.Fprintln(f, entryJSON)
		f.Close()
	}

	// Rolling: 최근 20줄만 유지
	data, err := os.ReadFile(jsonlFile)
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		if len(lines) > 20 {
			lines = lines[len(lines)-20:]
			os.WriteFile(jsonlFile, []byte(strings.Join(lines, "\n")+"\n"), 0600)
		}

		// 2. hippocampus 뉴런 업데이트
		neuronFile := filepath.Join(brainRoot, "hippocampus", "session_log", "절대_최근대화_전사기록_동기화.neuron")
		os.MkdirAll(filepath.Dir(neuronFile), 0750)
		var sb strings.Builder
		sb.WriteString("# 최근 대화 콘텍스트 주입 (실시간 동기화)\n\n이전 대화 맥락을 파악하고 대답을 연계할 것:\n\n")
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l == "" {
				continue
			}
			// 간단 파싱
			tsIdx := strings.Index(l, `"ts":"`)
			roleIdx := strings.Index(l, `"role":"`)
			textIdx := strings.Index(l, `"text":"`)
			if tsIdx >= 0 && roleIdx >= 0 && textIdx >= 0 {
				tsVal := l[tsIdx+6:]
				if i := strings.Index(tsVal, `"`); i > 0 {
					tsVal = tsVal[:i]
				}
				roleVal := l[roleIdx+8:]
				if i := strings.Index(roleVal, `"`); i > 0 {
					roleVal = roleVal[:i]
				}
				textVal := l[textIdx+8:]
				if i := strings.LastIndex(textVal, `"`); i > 0 {
					textVal = textVal[:i]
				}
				timeOnly := tsVal
				if len(tsVal) > 11 {
					timeOnly = tsVal[11:]
					if len(timeOnly) > 8 {
						timeOnly = timeOnly[:8]
					}
				}
				truncText := textVal
				if len([]rune(truncText)) > 400 {
					truncText = string([]rune(truncText)[:400])
				}
				sb.WriteString(fmt.Sprintf("[%s] %s: %s\n", timeOnly, strings.ToUpper(roleVal), truncText))
			}
		}
		os.WriteFile(neuronFile, []byte(sb.String()), 0600)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ── CDP DOM 스크래핑 + Network 모니터 ──
// 활성 Network scraper 추적
func hlAutoEvolve(brainRoot string) {

	for {
		time.Sleep(3 * time.Minute)

		// 자율주행 비활성 체크
		nfsRoot := filepath.Dir(brainRoot)
		if fileExists(filepath.Join(nfsRoot, "telegram-bridge", ".auto_evolve_disabled")) {
			continue
		}

		// 오토파일럿 전용 하트비트 (.autopilot_heartbeat)
		// growth.log는 idle_worker가 매 사이클 터치하므로 사용 불가
		heartbeatFile := filepath.Join(brainRoot, "hippocampus", "session_log", ".autopilot_heartbeat")
		info, err := os.Stat(heartbeatFile)

		// 하트비트 없으면 자동 생성 (6분 전)
		if err != nil {
			os.MkdirAll(filepath.Dir(heartbeatFile), 0750)
			os.WriteFile(heartbeatFile, []byte("autopilot\n"), 0600)
			old := time.Now().Add(-6 * time.Minute)
			os.Chtimes(heartbeatFile, old, old)
			info, err = os.Stat(heartbeatFile)
		}

		// 하트비트가 5분 이상 미갱신일 때만 (AI가 EVOLVE:proceed 출력 시 터치)
		if err == nil && time.Since(info.ModTime()) > 5*time.Minute {
			// 전사 파일 체크: 최근 5분 내 전사 갱신이 있어야만 활성 대화 존재
			transcriptDir := filepath.Join(brainRoot, "_transcripts")
			entries, _ := os.ReadDir(transcriptDir)
			activeConvo := false
			for _, e := range entries {
				if eInfo, err := e.Info(); err == nil {
					if time.Since(eInfo.ModTime()) < 5*time.Minute {
						activeConvo = true
						break
					}
				}
			}
			if !activeConvo {
				// 활성 대화 없음 — 아무도 안 듣고 있으니 주입 무의미
				continue
			}

			// CLI 오케스트레이터 모델: Go 런타임은 최소 트리거만 발사
			// 실제 판단과 행동은 CLI(AI 에이전트)가 수행
			minimalTrigger := "[NeuronFS 시스템 점검 트리거] 시스템 상태를 확인하고 필요시 조치하라."

			// _rules.md stale 방지: 90분마다 자동 re-inject (비차단)
			rulesPath := filepath.Join(brainRoot, "brainstem", "_rules.md")
			if info, err := os.Stat(rulesPath); err == nil {
				if time.Since(info.ModTime()) > 90*time.Minute {
					svLog("[HEARTBEAT] ♻️ _rules.md 90분 경과 — 자동 re-inject")
					go writeAllTiers(brainRoot)
				}
			}

			// 항상 트리거 — 이상 여부 상관없이 오토파일럿 구동
			svLog("[HEARTBEAT] 🚀 Gemini CLI 오토파일럿 발동")

			// ── 올바른 자율주행 플로우 ──
			// 1. 최근대화 + 마스터 프롬프트 조합
			// 2. Gemini CLI에 전달
			// 3. Gemini CLI 응답을 CDP로 이 Antigravity 창에 주입
			// 4. Antigravity 응답 → [EVOLVE:proceed] → 반복

			// 텔레그램 알림
			hlTgSend(hlTgChatID, minimalTrigger)

			// Gemini CLI에 최근대화 + 마스터 프롬프트 전달
			go func() {
				nudge, _ := hlBuildContextualPrompt(brainRoot)
				fullPrompt := hlMasterPrompt + "\n\n" + nudge

				nfsRoot2 := filepath.Dir(brainRoot)
				result := executeGeminiCLI(AgentTask{
					Name:    "autopilot_evolve",
					Prompt:  fullPrompt,
					WorkDir: nfsRoot2,
				})

				if result.Success && len(result.Output) > 10 {
					response := result.Output
					if len([]rune(response)) > 2000 {
						response = string([]rune(response)[:2000])
					}
					svLog("[AUTOPILOT] ✅ Gemini CLI 응답 수신")
					// CDP로 IDE 채팅창에 직접 주입 (텔레그램→IDE와 동일 경로)
					hlCDPInject(hlTgMountedRoom, response)
					// 텔레그램에도 알림
					hlTgSend(hlTgChatID, "[AUTOPILOT] ✅ CLI→IDE 주입 완료 ("+fmt.Sprintf("%d", len([]rune(response)))+"자)")
				} else {
					outputSnip := ""
					if len(result.Output) > 100 {
						outputSnip = result.Output[:100]
					} else {
						outputSnip = result.Output
					}
					fmt.Printf("[AUTOPILOT] ⚠️ Gemini CLI 실패: %s\n", outputSnip)
				}
			}()

			// 무한 루프 회피 터치
			hbFile2 := filepath.Join(brainRoot, "hippocampus", "session_log", ".autopilot_heartbeat")
			os.Chtimes(hbFile2, time.Now(), time.Now())
		}
	}
}

// ── 메인 런처 ──
func runHijackLauncher(brainRoot string) {
	nfsRoot := filepath.Dir(brainRoot)
	hlLoadTelegram(nfsRoot)

	svLog("[HL] 🚀 Hijack Launcher (Go native) 시작")

	// 텔레그램 양방향 polling
	go hlTgPoll(brainRoot)

	// CDP 모니터 & 큐 워커
	go hlStartCDPWorker()
	go hlStartCDPMonitor(brainRoot)

	// 자동 evolve
	go hlAutoEvolve(brainRoot)

	// keep alive
	select {}
}

// hlScrapeCurrentConversation — CDP로 현재 활성 대화 제목 스크래핑
func hlScrapeCurrentConversation() string {
	targets, err := cdpListTargets(hlCDPPort)
	if err != nil {
		return "(CDP 연결 실패)"
	}
	for _, t := range targets {
		if !strings.Contains(t.URL, "workbench.html") {
			continue
		}
		if t.WebSocketDebuggerURL == "" {
			continue
		}
		client, err := NewCDPClient(t.WebSocketDebuggerURL)
		if err != nil {
			continue
		}
		client.Call("Runtime.enable", map[string]interface{}{})
		time.Sleep(300 * time.Millisecond)

		// 현재 대화 제목 추출 (Antigravity 채팅 패널의 제목)
		scrapeExpr := `(() => {
			// 방법1: 탭 타이틀에서 추출
			const title = document.title || '';
			// 방법2: 활성 채팅 헤더에서 추출
			const chatHeader = document.querySelector('[class*="chat"] [class*="title"]') ||
				document.querySelector('.conversation-header') ||
				document.querySelector('[aria-label*="conversation"]');
			const headerText = chatHeader ? chatHeader.textContent.trim() : '';
			return headerText || title || 'unknown';
		})()`

		result, err := client.Call("Runtime.evaluate", map[string]interface{}{
			"expression":    scrapeExpr,
			"returnByValue": true,
		})
		client.Close()

		if err != nil {
			continue
		}

		var evalRes struct {
			Result struct {
				Value string `json:"value"`
			} `json:"result"`
		}
		json.Unmarshal(result, &evalRes)
		if evalRes.Result.Value != "" && evalRes.Result.Value != "unknown" {
			return evalRes.Result.Value
		}
		// fallback: 탭 타이틀 사용
		if t.Title != "" {
			return t.Title
		}
	}
	return "(대화 정보 없음)"
}

// appendDebugLog writes TG debug entries to tg_debug.log for troubleshooting
func appendDebugLog(nfsRoot, msg string) {
	logPath := filepath.Join(nfsRoot, "logs", "tg_debug.log")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "[%s] %s\n", time.Now().Format("15:04:05"), msg)
	
	// Broadcast to SSE (Dashboard V2)
	GlobalSSEBroker.Broadcast("info", fmt.Sprintf("[TG] %s", msg))
}
