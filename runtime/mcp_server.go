// NeuronFS MCP Server — Go Native
//
// PROVIDES: startMCPServer, startMCPServerWithStdout, startMCPHTTPServer,
//           buildFreshMCPServer, buildMCPInstructions,
//           registerActiveMCPServer, notifyMCPResourceUpdated
// DEPENDS ON: mcp_tools_native.go (registerMCPTools), emit_bootstrap.go (emitBootstrap),
//             emit_format_rules.go (buildPreamble), brain.go (scanBrain, runSubsumption)
//
// Node.js 래퍼(mcp/index.js)를 대체하여 Go에서 직접 MCP 프로토콜을 서빙한다.
// AI(Gemini/Claude) ↔ Go(stdio, JSON-RPC 2.0) ↔ Filesystem
//
// HTTP 모드: per-session factory → 매 세션마다 최신 brain 스캔 후 Instructions 주입
// ResourceUpdated: markBrainDirty() → 모든 활성 MCP 서버에 리소스 변경 알림

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// activeMCPServers tracks live MCP server instances for resource update notifications.
// When brain changes, we notify all active servers so IDE re-reads neuronfs://rules/current.
var (
	activeMCPServersMu sync.Mutex
	activeMCPServers   []*mcp.Server
)

// registerActiveMCPServer adds a server to the active list for notifications.
func registerActiveMCPServer(s *mcp.Server) {
	activeMCPServersMu.Lock()
	defer activeMCPServersMu.Unlock()
	activeMCPServers = append(activeMCPServers, s)
}

// notifyMCPResourceUpdated sends ResourceUpdated notification to all active MCP servers.
// Called when brain state changes (via markBrainDirty) so IDE re-reads the rules resource.
func notifyMCPResourceUpdated() {
	activeMCPServersMu.Lock()
	servers := make([]*mcp.Server, len(activeMCPServers))
	copy(servers, activeMCPServers)
	activeMCPServersMu.Unlock()

	for _, s := range servers {
		_ = s.ResourceUpdated(context.Background(), &mcp.ResourceUpdatedNotificationParams{
			URI: "neuronfs://rules/current",
		})
	}
}

// startMCPServer bootstraps the MCP stdio server using os.Stdin/os.Stdout.
// WARNING: Only use this when os.Stdout is clean (not redirected).
func startMCPServer(brainRoot string) {
	startMCPServerWithStdout(brainRoot, os.Stdout)
}

// startMCPServerWithStdout bootstraps the MCP stdio server with a specific stdout writer.
// This is used in --mcp mode where os.Stdout is redirected to stderr to prevent
// fmt.Print* from polluting the JSON-RPC channel.
func startMCPServerWithStdout(brainRoot string, stdout *os.File) {
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "neuronfs",
			Version: "1.0.0",
		},
		&mcp.ServerOptions{
			Instructions: buildMCPInstructions(brainRoot),
		},
	)

	registerMCPTools(server, brainRoot)

	// ━━━ Resource: 현재 규칙 (캐시 바이패스) ━━━
	server.AddResource(
		&mcp.Resource{
			URI:         "neuronfs://rules/current",
			Name:        "current_rules",
			Description: "현재 활성 뉴런 규칙. 매 호출 시 실시간 생성.",
			MIMEType:    "text/markdown",
		},
		func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			brain := scanBrain(brainRoot)
			result := runSubsumption(brain)
			content := emitBootstrap(result, brainRoot)
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{
					{URI: "neuronfs://rules/current", Text: content},
				},
			}, nil
		},
	)

	fmt.Fprintf(os.Stderr, "\033[36m[NEURON] Core Initialization Complete.\033[0m\n")
	fmt.Fprintf(os.Stderr, "\033[35m[SYNAPSE] Listening on stdio via Native MCP. Zero dependencies.\033[0m\n")
	fmt.Fprintf(os.Stderr, "\033[37m  - Waiting for context pulses...\033[0m\n")
	ctx := context.Background()
	// Use IOTransport instead of StdioTransport to avoid using os.Stdout
	// (which has been redirected to stderr in --mcp mode)
	transport := &mcp.IOTransport{
		Reader: os.Stdin,
		Writer: stdout,
	}
	if _, err := server.Connect(ctx, transport, nil); err != nil {
		log.Fatalf("[MCP] FATAL: %v\n", err)
	}

	// Block forever — stdio connection runs until process dies
	select {}
}

// registerMCPTools registers all NeuronFS commands as tools in the MCP server.
func registerMCPTools(server *mcp.Server, brainRoot string) {
	registerMCPReadTools(server, brainRoot)
	registerMCPCRUDTools(server, brainRoot)
	registerMCPSysTools(server, brainRoot)
	registerMCPEvolveTools(server, brainRoot)
	RegisterTemporalAndEpisodicTools(server, brainRoot)

	RegisterNativeTools(server, brainRoot)
}

// ─── Helpers ───

func mcpError(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: "❌ " + msg}},
		IsError: true,
	}
}

// mcpCallCounter tracks tool call count for periodic auto-commit.
var (
	mcpCallCounter int
	mcpCallMu      sync.Mutex
)

// autoGitCommitInterval: N회 MCP 호출마다 brain_v4 자동 git commit
const autoGitCommitInterval = 10

// mcpWithRules wraps a tool response with P0 rules reminder.
// Every MCP tool call response gets P0 rules appended, so rules are
// continuously re-injected into the LLM's context window.
// This combats the "Lost in the Middle" attention decay problem.
// Also triggers auto git commit every N calls.
func mcpWithRules(brainRoot string, text string) *mcp.CallToolResult {
	reminder := buildP0Reminder(brainRoot)

	// 자동 로컬 깃 커밋 (N회마다)
	mcpCallMu.Lock()
	mcpCallCounter++
	count := mcpCallCounter
	mcpCallMu.Unlock()
	if count%autoGitCommitInterval == 0 {
		go func() {
			ts := time.Now().Format("2006-01-02T15:04:05")
			SafeExecDir(ExecTimeoutGit, brainRoot, "git", "add", "-A")
			SafeExecDir(ExecTimeoutGit, brainRoot, "git", "commit", "-m",
				fmt.Sprintf("[auto] MCP call #%d — %s", count, ts))
			log.Printf("[MCP] auto git commit #%d at %s", count, ts)
		}()
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text + reminder}},
	}
}

// buildP0Reminder generates a compact P0 rules string from brainstem 禁 neurons.
// Cached for 60 seconds to avoid scanning on every tool call.
var (
	cachedP0     string
	cachedP0Time time.Time
)

func buildP0Reminder(brainRoot string) string {
	// Cache for 60 seconds
	if time.Since(cachedP0Time) < 60*time.Second && cachedP0 != "" {
		return cachedP0
	}

	brain := scanBrain(brainRoot)
	var bans []string
	for _, region := range brain.Regions {
		if region.Name != "brainstem" {
			continue
		}
		for _, n := range region.Neurons {
			if n.IsDormant || n.Counter < 5 {
				continue
			}
			if strings.ContainsAny(n.Path, "禁") {
				sentence := pathToSentence(n.Path)
				trimmed := strings.TrimSpace(strings.ReplaceAll(sentence, "절대 금지:", ""))
				if len([]rune(trimmed)) >= 2 {
					bans = append(bans, trimmed)
				}
			}
		}
	}

	if len(bans) > 3 {
		bans = bans[:3]
	}

	// Read preamble for language rule
	preamblePath := filepath.Join(brainRoot, "_preamble.txt")
	langRule := ""
	if data, err := os.ReadFile(preamblePath); err == nil {
		lines := strings.Split(string(data), "\n")
		if len(lines) > 0 {
			langRule = strings.TrimSpace(lines[0])
		}
	}

	var sb strings.Builder
	sb.WriteString("\n\n---\n⚡ P0: ")
	if langRule != "" {
		sb.WriteString(langRule)
		sb.WriteString(" | ")
	}
	sb.WriteString("금지: " + strings.Join(bans, ", "))

	cachedP0 = sb.String()
	cachedP0Time = time.Now()
	return cachedP0
}

// boolPtr returns a pointer to a boolean value.
func boolPtr(b bool) *bool {
	return &b
}

// buildMCPInstructions generates the server instructions from current brain state.
// These instructions are injected into the LLM's context on every session initialize.
func buildMCPInstructions(brainRoot string) string {
	brain := scanBrain(brainRoot)
	result := runSubsumption(brain)
	return emitBootstrap(result, brainRoot)
}

// logWriter returns stderr for MCP mode (stdout is reserved for JSON-RPC)
func logWriter() *os.File {
	return os.Stderr
}

// buildFreshMCPServer creates a new MCP server with latest brain state.
// Called per-session in HTTP mode so each conversation gets fresh Instructions.
func buildFreshMCPServer(brainRoot string) *mcp.Server {
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "neuronfs",
			Version: "1.0.0",
		},
		&mcp.ServerOptions{
			Instructions: buildMCPInstructions(brainRoot),
		},
	)

	registerMCPTools(server, brainRoot)

	// Resource: 현재 규칙 (캐시 바이패스)
	server.AddResource(
		&mcp.Resource{
			URI:         "neuronfs://rules/current",
			Name:        "current_rules",
			Description: "현재 활성 뉴런 규칙. 매 호출 시 실시간 생성.",
			MIMEType:    "text/markdown",
		},
		func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			brain := scanBrain(brainRoot)
			result := runSubsumption(brain)
			content := emitBootstrap(result, brainRoot)
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{
					{URI: "neuronfs://rules/current", Text: content},
				},
			}, nil
		},
	)

	registerActiveMCPServer(server)
	return server
}

// startMCPHTTPServer bootstraps the MCP server over Streamable HTTP transport.
// Unlike stdio, this survives IDE restarts — the server runs independently.
// Clients connect via HTTP POST/GET to http://localhost:MCPStreamPort/mcp
func startMCPHTTPServer(brainRoot string, port int) {
	// Factory: 매 세션(대화)마다 새 서버를 생성하여 최신 Instructions 주입
	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return buildFreshMCPServer(brainRoot)
	}, &mcp.StreamableHTTPOptions{})

	mux := http.NewServeMux()
	mux.Handle("/mcp", handler)

	// Health check endpoint for supervisor
	mux.HandleFunc("/mcp/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","transport":"streamable-http","port":%d}`, port)
	})

	fmt.Fprintf(os.Stderr, "\033[36m[NEURON] MCP Streamable HTTP on :%d\033[0m\n", port)
	fmt.Fprintf(os.Stderr, "\033[35m[SYNAPSE] IDE-independent. Survives restarts.\033[0m\n")

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		// Port conflict: retry with backoff instead of killing entire process
		fmt.Fprintf(os.Stderr, "[MCP-HTTP] port :%d in use, retrying...\n", port)
		for i := 0; i < 10; i++ {
			time.Sleep(time.Duration(3*(i+1)) * time.Second)
			fmt.Fprintf(os.Stderr, "[MCP-HTTP] retry %d/10 on :%d\n", i+1, port)
			if err2 := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err2 == nil {
				return
			}
		}
		fmt.Fprintf(os.Stderr, "[MCP-HTTP] gave up after 10 retries on :%d\n", port)
	}
}
