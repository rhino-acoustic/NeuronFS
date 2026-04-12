package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

// GenerateDashboardSVG creates a stunning procedural 2D SVG version of the V2 Dashboard.
// It maps the Brain's architecture into a mathematical constellation of nodes
// and overlays the Autopilot Nerve Center TTY.
func GenerateDashboardSVG(brain Brain, totalNeurons int, totalActivation int) string {
	width := 1200
	height := 600
	rand.Seed(time.Now().UnixNano())

	// ── 1. BACKGROUND & LAYOUT ──
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d">`, width, height, width, height))
	// Fonts & Animations
	sb.WriteString(`
	<defs>
		<style>
			@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;800&amp;family=JetBrains+Mono:wght@400;700&amp;display=swap');
			.bg { fill: #030305; }
			.border { stroke: rgba(255,255,255,0.08); stroke-width: 1; }
			.accent { fill: #10b981; }
			.text-main { fill: #f8fafc; font-family: 'Inter', sans-serif; }
			.text-mono { fill: #a1a1aa; font-family: 'JetBrains Mono', monospace; font-size: 12px; }
			.pulse { animation: pulseAnim 2s infinite; }
			@keyframes pulseAnim {
				0% { r: 5; opacity: 1; }
				50% { r: 15; opacity: 0; }
				100% { r: 5; opacity: 0; }
			}
			.blink { animation: blinkAnim 1s step-end infinite; }
			@keyframes blinkAnim { 50% { opacity: 0; } }
		</style>
		<radialGradient id="hologram" cx="50%" cy="50%" r="50%">
			<stop offset="0%" stop-color="#0a0a14" />
			<stop offset="100%" stop-color="#030305" />
		</radialGradient>
	</defs>
	`)

	// Canvas backgrounds
	sb.WriteString(fmt.Sprintf(`<rect width="%d" height="%d" fill="url(#hologram)" />`, width, height))
	// Right panel background
	panelWidth := 450
	panelX := width - panelWidth
	sb.WriteString(fmt.Sprintf(`<rect x="%d" y="0" width="%d" height="%d" fill="rgba(15, 20, 30, 0.6)" class="border" />`, panelX, panelWidth, height))
	// Separator line
	sb.WriteString(fmt.Sprintf(`<line x1="%d" y1="0" x2="%d" y2="%d" class="border" />`, panelX, panelX, height))

	// ── 2. HOLOGRAPHIC BRAIN (Procedural Scatter) ──
	// Simulate the 3D geometry of the V2 dashboard on a 2D plane
	nodesToDraw := 1500 // SVG performance limit
	if totalNeurons < 1500 && totalNeurons > 0 {
		nodesToDraw = totalNeurons
	}
	
	centerX := float64(panelX) / 2.0
	centerY := float64(height) / 2.0

	for i := 0; i < nodesToDraw; i++ {
		t := rand.Float64() * math.Pi * 2
		p := math.Acos(2*rand.Float64() - 1)
		
		rx, ry, rz := 220.0, 160.0, 190.0
		f := 1 + 0.2*math.Pow(math.Max(0, math.Sin(t)*math.Cos(p)), 2)
		tm := 1 + 0.12*math.Pow(math.Sin(t), 2)*math.Pow(math.Sin(p), 2)
		bt := 1.0
		if t > math.Pi*0.75 {
			bt = 1 - 0.3*(t-math.Pi*0.75)/(math.Pi*0.25)
		}
		gr := 1 - 0.08*math.Pow(math.Max(0, math.Cos(p)), 4)*math.Pow(math.Sin(t), 2)
		r := f * tm * bt * gr * (0.5 + rand.Float64()*0.5)

		x := rx * r * math.Sin(t) * math.Cos(p)
		y := ry * r * math.Cos(t)
		z := rz * r * math.Sin(t) * math.Sin(p)

		// Simple perspective projection
		scale := 800.0 / (800.0 + z)
		px := centerX + x*scale
		py := centerY - y*scale

		// Determine color based on Y height (Cortex vs Brainstem)
		color := "#10b981" // Emerald
		if y > 50 {
			color = "#3b82f6" // Blue
		} else if y < -50 {
			color = "#dc2626" // Red
		}
		opacity := math.Min(1.0, math.Max(0.1, 1.0-(dist(x,y)/250.0)))
		radius := math.Max(0.5, 2.5*scale*opacity)

		sb.WriteString(fmt.Sprintf(`<circle cx="%.2f" cy="%.2f" r="%.2f" fill="%s" opacity="%.2f" />`, px, py, radius, color, opacity))
	}

	// ── 3. RIGHT PANEL CONSOLE ──
	// Header
	sb.WriteString(fmt.Sprintf(`<text x="%d" y="40" class="text-main" font-size="22" font-weight="800">Neuron</text>`, panelX+30))
	sb.WriteString(fmt.Sprintf(`<text x="%d" y="40" font-family="'Inter', sans-serif" font-size="22" font-weight="800" fill="#10b981">FS V2</text>`, panelX+112))
	sb.WriteString(fmt.Sprintf(`<text x="%d" y="60" class="text-mono" fill="#94a3b8">Autopilot Nerve Center</text>`, panelX+30))
	sb.WriteString(fmt.Sprintf(`<line x1="%d" y1="80" x2="%d" y2="80" class="border" />`, panelX, panelX+panelWidth))

	// Autopilot Status
	sb.WriteString(fmt.Sprintf(`<circle cx="%d" cy="115" r="5" fill="#10b981" />`, panelX+40))
	sb.WriteString(fmt.Sprintf(`<circle cx="%d" cy="115" r="5" fill="none" stroke="#10b981" stroke-width="2" class="pulse" />`, panelX+40))
	sb.WriteString(fmt.Sprintf(`<text x="%d" y="120" class="text-main" font-size="14" font-weight="600">System Listening &amp; Ready</text>`, panelX+60))
	sb.WriteString(fmt.Sprintf(`<line x1="%d" y1="150" x2="%d" y2="150" class="border" />`, panelX, panelX+panelWidth))

	// TTY Feed (Mocked latest operations)
	now := time.Now().Format("15:04:05")
	sb.WriteString(fmt.Sprintf(`<text x="%d" y="190" class="text-mono">[SYSTEM] Booting Strangler Fig CLI Router...</text>`, panelX+30))
	sb.WriteString(fmt.Sprintf(`<text x="%d" y="215" class="text-mono">[SYSTEM] Neural filesystem mounted successfully.</text>`, panelX+30))
	sb.WriteString(fmt.Sprintf(`<text x="%d" y="240" class="text-mono">[%s] [WARN] [EVOLVE:proceed] tag detected.</text>`, panelX+30, now))
	sb.WriteString(fmt.Sprintf(`<text x="%d" y="265" class="text-mono" fill="#10b981">[%s] [INFO] CDP Injection Cooldown: 60s</text>`, panelX+30, now))
	
	// Live Cursor
	sb.WriteString(fmt.Sprintf(`<text x="%d" y="290" class="text-mono blink" font-weight="700">_</text>`, panelX+30))

	// Metrics Footer
	sb.WriteString(fmt.Sprintf(`<line x1="%d" y1="480" x2="%d" y2="480" class="border" />`, panelX, panelX+panelWidth))
	
	// Metric 1: Total Neurons
	sb.WriteString(fmt.Sprintf(`<text x="%d" y="520" font-family="'JetBrains Mono', monospace" font-size="10" fill="#94a3b8" letter-spacing="1">ACTIVE NEURONS</text>`, panelX+30))
	sb.WriteString(fmt.Sprintf(`<text x="%d" y="550" class="text-main" font-size="24" font-weight="800">%d</text>`, panelX+30, totalNeurons))

	// Metric 2: Total Activation
	sb.WriteString(fmt.Sprintf(`<text x="%d" y="520" font-family="'JetBrains Mono', monospace" font-size="10" fill="#94a3b8" letter-spacing="1">TOTAL ACTIVATION</text>`, panelX+250))
	sb.WriteString(fmt.Sprintf(`<text x="%d" y="550" class="text-main" font-size="24" font-weight="800">%d</text>`, panelX+250, totalActivation))

	sb.WriteString(`</svg>`)
	return sb.String()
}

func dist(x, y float64) float64 {
	return math.Sqrt(x*x + y*y)
}
