// PROVIDES: runMCPProxy, healSession
// DEPENDS ON: (stdlib net/http only)
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	mcpSessionMap   = make(map[string]string)
	mcpSessionMapMu sync.RWMutex
	mcpSessionFile  string // set by runMCPProxy
)

// loadSessionMap reads persisted session mappings from disk
func loadSessionMap() {
	if mcpSessionFile == "" {
		return
	}
	data, err := os.ReadFile(mcpSessionFile)
	if err != nil {
		return // first run or missing file — normal
	}
	mcpSessionMapMu.Lock()
	defer mcpSessionMapMu.Unlock()
	_ = json.Unmarshal(data, &mcpSessionMap)
}

// saveSessionMap writes current session mappings to disk atomically
func saveSessionMap() {
	if mcpSessionFile == "" {
		return
	}
	mcpSessionMapMu.RLock()
	data, _ := json.MarshalIndent(mcpSessionMap, "", "  ")
	mcpSessionMapMu.RUnlock()
	_ = os.WriteFile(mcpSessionFile, data, 0600)
}


// runMCPProxy starts a reverse proxy on listenPort that forwards to targetPort.
func runMCPProxy(listenPort int, targetPort int) {
	// Persist session map to brain root
	if mcpBrainRoot != "" {
		mcpSessionFile = filepath.Join(mcpBrainRoot, ".mcp_sessions.json")
		loadSessionMap()
	}

	targetBase := fmt.Sprintf("http://127.0.0.1:%d", targetPort)
	targetURL, _ := url.Parse(targetBase)

	director := func(req *http.Request) {
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host

		clientSess := req.Header.Get("Mcp-Session-Id")
		if clientSess != "" {
			req.Header.Set("X-Original-Client-Session-Id", clientSess)
			mcpSessionMapMu.RLock()
			backendSess := mcpSessionMap[clientSess]
			mcpSessionMapMu.RUnlock()
			if backendSess != "" {
				req.Header.Set("Mcp-Session-Id", backendSess)
			}
		}
	}

	proxy := &httputil.ReverseProxy{
		Director: director,
		Transport: &retryTransport{
			MaxRetries: 3,
			RetryDelay: 500 * time.Millisecond,
			TargetBase: targetBase,
		},
		ModifyResponse: func(resp *http.Response) error {
			if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
				sess := resp.Header.Get("Mcp-Session-Id")
				clientSess := resp.Request.Header.Get("X-Original-Client-Session-Id")
				if sess != "" && clientSess == "" {
					mcpSessionMapMu.Lock()
					mcpSessionMap[sess] = sess
					mcpSessionMapMu.Unlock()
					saveSessionMap()
				}
			}
			return nil
		},
	}

	fmt.Fprintf(os.Stderr, "\033[36m[PROXY] Shielding layer active on :%d -> :%d\033[0m\n", listenPort, targetPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", listenPort), proxy); err != nil {
		fmt.Fprintf(os.Stderr, "[PROXY] FATAL: %v\n", err)
	}
}

type retryTransport struct {
	MaxRetries int
	RetryDelay time.Duration
	TargetBase string
}

func healSession(targetBase string) string {
	initPayload := `{"jsonrpc":"2.0","id":999,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"proxy-healer","version":"1"}}}`
	resp, err := http.Post(targetBase+"/mcp", "application/json", strings.NewReader(initPayload))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	
	if resp.StatusCode == http.StatusCreated {
		sess := resp.Header.Get("Mcp-Session-Id")
		
		// Send initialized notification
		notif := `{"jsonrpc":"2.0","method":"notifications/initialized"}`
		req, _ := http.NewRequest("POST", targetBase+"/mcp/message", strings.NewReader(notif))
		req.Header.Set("Mcp-Session-Id", sess)
		req.Header.Set("Content-Type", "application/json")
		nr, err := http.DefaultClient.Do(req)
		if err == nil && nr != nil {
			nr.Body.Close()
		}
		return sess
	}
	return ""
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	var reqBody []byte
	if req.Body != nil {
		reqBody, _ = io.ReadAll(req.Body)
		req.Body.Close()
	}

	for i := 0; i < t.MaxRetries; i++ {
		if reqBody != nil {
			req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
		}

		resp, err = http.DefaultTransport.RoundTrip(req)
		
		if err != nil {
			time.Sleep(t.RetryDelay)
			continue
		}

		clientSess := req.Header.Get("X-Original-Client-Session-Id")
		if clientSess == "" {
			clientSess = req.URL.Query().Get("sessionId")
		}
		isLoss := resp.StatusCode == http.StatusNotFound && strings.Contains(req.URL.Path, "/mcp")
		
		if isLoss {
			if clientSess != "" {
				// We can try to heal, but if it's an SSE GET request, healing might be complex.
				if req.Method != http.MethodGet {
					resp.Body.Close() // Discard original 404 response
					
					// Heal the session with a full initialize cycle
					newBackendSess := healSession(t.TargetBase)
					if newBackendSess != "" {
						mcpSessionMapMu.Lock()
						mcpSessionMap[clientSess] = newBackendSess
						mcpSessionMapMu.Unlock()
						saveSessionMap()
						
						fmt.Fprintf(os.Stderr, "\033[35m[PROXY] Auto-healed session mappings: %s -> %s\033[0m\n", clientSess[:8], newBackendSess[:8])
						
						// Retry the original request with the new session ID
						req.Header.Set("Mcp-Session-Id", newBackendSess)
						q := req.URL.Query()
						if q.Get("sessionId") != "" {
							q.Set("sessionId", newBackendSess)
							req.URL.RawQuery = q.Encode()
						}
						
						if reqBody != nil {
							req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
						}
						resp, err = http.DefaultTransport.RoundTrip(req)
						return resp, err
					}
				} else {
					fmt.Fprintf(os.Stderr, "\033[31m[PROXY] 🚨 MCP SSE Session Drop Detected! (IDE Reconnect Required) Session: %s\033[0m\n", clientSess)
				}
			} else {
				fmt.Fprintf(os.Stderr, "\033[31m[PROXY] 🚨 Unrecoverable MCP 404 (Session Lost) at %s\033[0m\n", req.URL.Path)
			}
		}

		return resp, nil
	}
	return resp, err
}