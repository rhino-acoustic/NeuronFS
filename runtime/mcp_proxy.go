package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

// runMCPProxy starts a reverse proxy on listenPort that forwards to targetPort.
// It acts as a shield to hold client (IDE) connections gracefully while the actual MCP worker restarts.
func runMCPProxy(listenPort int, targetPort int) {
	targetURL, _ := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", targetPort))

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Transport = &retryTransport{
		MaxRetries: 15,
		RetryDelay: 1 * time.Second,
	}

	fmt.Fprintf(os.Stderr, "\033[36m[PROXY] Shielding layer active on :%d -> :%d\033[0m\n", listenPort, targetPort)
	if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", listenPort), proxy); err != nil {
		fmt.Fprintf(os.Stderr, "[PROXY] FATAL: %v\n", err)
	}
}

type retryTransport struct {
	MaxRetries int
	RetryDelay time.Duration
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	for i := 0; i < t.MaxRetries; i++ {
		resp, err = http.DefaultTransport.RoundTrip(req)
		if err == nil {
			return resp, nil
		}
		// If connection refused (backend worker is restarting), wait and retry
		time.Sleep(t.RetryDelay)
	}
	return resp, err
}
