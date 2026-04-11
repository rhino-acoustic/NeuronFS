package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterTemporalAndEpisodicTools adds Temporal Query and Episodic Context automation tools.
func RegisterTemporalAndEpisodicTools(s *mcp.Server, brainRoot string) {
	// ─── TOOL: mcp_neuronfs_search ───
	s.AddTool(
		&mcp.Tool{
			Name:        "mcp_neuronfs_search",
			Description: "시간축 정렬(Temporal Query) 기반 뉴런 검색 도구. '최근', '어제' 등의 시간 질문 시 이 도구를 호출한다.",
			InputSchema: json.RawMessage(`{"type": "object", "properties": {"sort_by": {"type": "string", "description":"created_at or activation", "default": "created_at"}, "time_filter": {"type": "string", "description":"e.g. 24h, 7d"}}, "required": ["sort_by"]}`),
		},
		func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				SortBy     string `json:"sort_by"`
				TimeFilter string `json:"time_filter"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return mcpError("invalid arguments: " + err.Error()), nil
			}

			log.Printf("[MCP] search sort=%s filter=%s", args.SortBy, args.TimeFilter)
			brain := scanBrain(brainRoot)
			type Result struct {
				Path    string
				ModTime time.Time
				Count   int
			}
			var results []Result

			now := time.Now()
			var durationFilter time.Duration
			if args.TimeFilter != "" {
				parsed, err := time.ParseDuration(strings.ReplaceAll(args.TimeFilter, "d", "h") + "h")
				if err == nil {
					durationFilter = parsed
				} else if args.TimeFilter == "24h" || args.TimeFilter == "1d" {
					durationFilter = 24 * time.Hour
				} else if args.TimeFilter == "7d" {
					durationFilter = 7 * 24 * time.Hour
				}
			}

			for _, r := range brain.Regions {
				for _, n := range r.Neurons {
					fullPath := filepath.Join(brainRoot, strings.ReplaceAll(n.Path, "/", string(filepath.Separator)))
					info, err := os.Stat(fullPath)
					if err != nil {
						continue
					}

					modTime := info.ModTime()
					if durationFilter > 0 && now.Sub(modTime) > durationFilter {
						continue
					}

					results = append(results, Result{
						Path:    n.Path,
						ModTime: modTime,
						Count:   n.Counter,
					})
				}
			}

			if args.SortBy == "created_at" || args.SortBy == "updated_at" {
				sort.Slice(results, func(i, j int) bool {
					return results[i].ModTime.After(results[j].ModTime)
				})
			} else {
				sort.Slice(results, func(i, j int) bool {
					return results[i].Count > results[j].Count
				})
			}

			limit := 10
			if len(results) < limit {
				limit = len(results)
			}

			var output strings.Builder
			output.WriteString(fmt.Sprintf("🔍 검색 완료 (%d 매칭). 상위 %d개 (정렬: %s):\n", len(results), limit, args.SortBy))
			for i := 0; i < limit; i++ {
				r := results[i]
				output.WriteString(fmt.Sprintf("- [%s] 발화: %d | 갱신: %s\n", r.Path, r.Count, r.ModTime.Format("2006-01-02 15:04:05")))
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: output.String()}},
			}, nil
		},
	)

	// ─── TOOL: mcp_neuronfs_log_episode ───
	s.AddTool(
		&mcp.Tool{
			Name:        "mcp_neuronfs_log_episode",
			Description: "대화 중 발생한 에피소드(컨텍스트, 주요 변경 등)를 영구 기록한다. 호출 시 hippocampus/episodes 경로에 뉴런이 자동 생성된다.",
			InputSchema: json.RawMessage(`{"type": "object", "properties": {"title": {"type": "string", "description":"짧은 사유. 예: A를_B로_변경"}, "content": {"type": "string"}}, "required": ["title", "content"]}`),
		},
		func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				Title   string `json:"title"`
				Content string `json:"content"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return mcpError("invalid arguments: " + err.Error()), nil
			}

			// Generate path: hippocampus/episodes/2026-04-09_Title
			dateStr := time.Now().Format("2006-01-02")
			safeTitle := strings.ReplaceAll(args.Title, " ", "_")
			neuronStr := fmt.Sprintf("%s_%s", dateStr, safeTitle)
			neuronPath := filepath.Join("hippocampus", "episodes", neuronStr)

			log.Printf("[MCP] log_episode title=%s", args.Title)
			if err := growNeuron(brainRoot, neuronPath); err != nil {
				return mcpError("failed to create episode neuron: " + err.Error()), nil
			}

			fullPath := filepath.Join(brainRoot, strings.ReplaceAll(neuronPath, "/", string(filepath.Separator)))
			os.WriteFile(filepath.Join(fullPath, "payload.json"), []byte(args.Content), 0600)

			// TTL GC check: Remove episodes older than 7 days
			go func() {
				episodesDir := filepath.Join(brainRoot, "hippocampus", "episodes")
				infos, _ := os.ReadDir(episodesDir)
				threshold := time.Now().Add(-168 * time.Hour)
				for _, info := range infos {
					if info.IsDir() {
						fullP := filepath.Join(episodesDir, info.Name())
						stat, err := os.Stat(fullP)
						if err == nil && stat.ModTime().Before(threshold) {
							os.RemoveAll(fullP) // TTL trigger
						}
					}
				}
			}()

			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "✅ 에피소드 생성 성공: " + neuronPath + " (7일 후 비동기 GC 파기 예정)"}},
			}, nil
		},
	)
}
