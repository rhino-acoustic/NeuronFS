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
			// Track activity (skip dashboard polling to avoid resetting idle timer)
			if r.URL.Path != "/api/brain" && r.URL.Path != "/api/state" && r.URL.Path != "/favicon.ico" {
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

	// Start background loops
	go runInjectionLoop(brainRoot)
	go runIdleLoop(brainRoot)

	fmt.Printf("  🔄 IDLE ENGINE: auto evolve/snapshot/NAS every %dm idle\n", idleThresholdMinutes)

	http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}
