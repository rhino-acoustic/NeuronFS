package main

// ━━━ dashboard_html.go ━━━
// PROVIDES: dashboardHTML (embedded from dashboard.html)
// Zero-maintenance: edit dashboard.html directly, go build embeds it

import _ "embed"

//go:embed dashboard.html
var dashboardHTML string
