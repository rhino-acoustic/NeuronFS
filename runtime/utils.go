package main

import (
	"net/http"
	"os"
	"time"
)

// [EVOLVE:proceed] 자가발전 자동 승인

// fileExists checks if a file or directory exists at the given path.
func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// Add returns the sum of two integers.
func Add(a, b int) int {
	return a + b
}

// ── 공용 HTTP 클라이언트 (커넥션 풀 재사용) ──
var (
	httpShort  = &http.Client{Timeout: 5 * time.Second}  // health check, CDP
	httpMedium = &http.Client{Timeout: 15 * time.Second} // Telegram, webhook
	httpLong   = &http.Client{Timeout: 60 * time.Second} // LLM API (Groq)
)
