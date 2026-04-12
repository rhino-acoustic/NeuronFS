// api_server.go — REST API 서버 엔트리 + 라우트 등록
//
// PROVIDES: startAPI
// DEPENDS:  api_handlers.go (핸들러 함수들)
//           api_static.go (대시보드/정적 파일 서빙)
//
// CALL GRAPH:
//   main.go --api → startAPI(brainRoot, port)
//     ├→ withCORS (CORS 미들웨어)
//     ├→ api_handlers.go 핸들러 등록 (CRUD, emotion, integrity, etc.)
//     ├→ api_static.go 정적 파일 라우트 등록
//     ├→ runInjectionLoop (background)
//     ├→ runIdleLoop (background)
//     └→ http.ListenAndServe

package main

import (
	"fmt"
	"net/http"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// REST API: Programmatic growth for n8n/dashboard/webhooks
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
func startAPI(brainRoot string, port int) {
	mux := http.NewServeMux()

	// Initialize activity tracker
	touchActivity()

	// CORS middleware with activity tracking
	withCORS := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == "OPTIONS" {
				w.WriteHeader(200)
				return
			}
			// Track activity (skip dashboard/monitoring polling to avoid resetting idle timer)
			// 대시보드 자동 폴링: /api/brain(3s), /api/health(heartbeat), /api/state, GET 요청
			skipTouch := r.URL.Path == "/api/brain" || r.URL.Path == "/api/state" ||
				r.URL.Path == "/api/health" || r.URL.Path == "/api/dashboard_state" ||
				r.URL.Path == "/favicon.ico" || r.URL.Path == "/"
			if !skipTouch && r.Method != "OPTIONS" {
				touchActivity()
			}
			h(w, r)
		}
	}

	// ── CRUD 핸들러 등록 (api_handlers.go) ──
	registerCRUDRoutes(mux, brainRoot, withCORS)

	// ── 설정/감정/샌드박스 핸들러 등록 (api_handlers.go) ──
	registerConfigRoutes(mux, brainRoot, withCORS)

	// ── 무결성/롤백/진화 핸들러 등록 (api_handlers.go) ──
	registerSystemRoutes(mux, brainRoot, withCORS)

	// ── 정적 파일/대시보드 등록 (api_static.go) ──
	registerStaticRoutes(mux, brainRoot, withCORS)

	// ── GraphQL / B2B Webhook 통합 (Phase 32) ──
	mux.HandleFunc("/graphql", withCORS(HandleGraphQL(brainRoot)))

	// Start background loops
	go runInjectionLoop(brainRoot)
	go runIdleLoop(brainRoot)
	go startFSWatcherPool(brainRoot) // <--- Phase 31: V7 OS Sensory Monitor
	go startWebhookDaemon(brainRoot) // <--- Phase 32: B2B Webhook Forwarder
	go startP2PSyncDaemon(brainRoot) // <--- Phase 36: V8 P2P Knowledge Crossover
	go startDreamCycleDaemon(brainRoot) // <--- Phase 41: V10 Dream Cycle
	go func() { BuildSimilarityIndex(brainRoot) }() // <--- Phase 44: V11 TF-IDF Index

	fmt.Printf("  🔄 IDLE ENGINE: auto evolve/snapshot/NAS every %dm idle\n", idleThresholdMinutes)

	http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}
