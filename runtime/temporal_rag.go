package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ============================================================================
// Module: Temporal RAG Processor (Phase 39)
// Provides past context to the NPU based on delta snapshot history mapping.
// ============================================================================

// DeltaMeta represents metadata of a tracked delta timeline
type DeltaMeta struct {
	Path      string
	Timestamp int64
	Content   string
}

// BuildTemporalRAGContext scans the temporal_log for keywords present in the prompt
// and returns a formatted RAG context block.
func BuildTemporalRAGContext(brainRoot, prompt string) string {
	if brainRoot == "" {
		return ""
	}

	temporalDir := filepath.Join(brainRoot, "hippocampus", "temporal_log")

	// Extremely naive keyword extraction (in a real system we'd use TF-IDF or Vector Embeddings)
	words := strings.Fields(prompt)
	keywords := []string{}
	for _, w := range words {
		if len(w) > 4 { // Only care about substantial words
			keywords = append(keywords, strings.ToLower(w))
		}
	}

	if len(keywords) == 0 {
		return ""
	}

	var matchedDeltas []DeltaMeta

	_ = filepath.WalkDir(temporalDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".delta") {
			return nil
		}

		// Read content
		contentBytes, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		
		contentStr := string(contentBytes)
		contentLower := strings.ToLower(contentStr)

		// Naive score based on keyword hits
		score := 0
		for _, kw := range keywords {
			if strings.Contains(contentLower, kw) {
				score++
			}
		}

		if score > 0 {
			info, _ := d.Info()
			matchedDeltas = append(matchedDeltas, DeltaMeta{
				Path:      path,
				Timestamp: info.ModTime().UnixMilli(),
				Content:   contentStr,
			})
		}
		return nil
	})

	if len(matchedDeltas) == 0 {
		return ""
	}

	// Sort by newest first
	sort.Slice(matchedDeltas, func(i, j int) bool {
		return matchedDeltas[i].Timestamp > matchedDeltas[j].Timestamp
	})

	// Take Top 3 latest matched historical snapshots
	limit := 3
	if len(matchedDeltas) < limit {
		limit = len(matchedDeltas)
	}

	var sb strings.Builder
	sb.WriteString("\n<TemporalContext>\n")
	sb.WriteString("The following are historical states of local files from the past (T-n):\n")

	for i := 0; i < limit; i++ {
		dm := matchedDeltas[i]
		tStr := time.UnixMilli(dm.Timestamp).Format(time.RFC3339)
		sb.WriteString(fmt.Sprintf("\n[Snapshot Date: %s]\n", tStr))
		
		// Truncate to save tokens (Max 500 chars per delta)
		displayContent := dm.Content
		if len(displayContent) > 500 {
			displayContent = displayContent[:500] + "\n...(truncated)"
		}
		sb.WriteString(displayContent)
		sb.WriteString("\n")
	}
	sb.WriteString("</TemporalContext>\n\n")

	return sb.String()
}
