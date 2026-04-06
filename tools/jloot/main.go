// jloot — MD/Folder → Binary Cartridge + In-Memory Brain Engine
// Pack, unpack, mount, and serve knowledge cartridges
//
// Usage:
//   jloot pack   <file.md|dir>          → output.jloot
//   jloot unpack <file.jloot> [-o dir]  → restore
//   jloot info   <file.jloot>           → metadata
//   jloot mount  <file.jloot> [-p port] → in-memory mount + HTTP API
//   jloot search <file.jloot> <keyword> → search in-memory

package main

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
)

// .jloot v2 format:
// [magic:7]["JLOOT\x00\x02"]
// [flags:1][0x00=single, 0x01=tree]
// [file_count:4][uint32 LE]
// [orig_total:4][uint32 LE — total original size]
// [hash:32][SHA-256 of concat all content]
// [timestamp:8][int64 LE]
// [gzip_data:*]
//
// Inside gzip for tree mode, each entry:
//   [path_len:2][path:*][content_len:4][content:*]

var magic = []byte("JLOOT\x00\x02")

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, `jloot — Binary Cartridge + In-Memory Brain Engine

Usage:
  jloot pack   <file.md|dir>          Pack into .jloot
  jloot unpack <file.jloot> [-o dir]  Unpack to dir/stdout
  jloot info   <file.jloot>           Show metadata
  jloot mount  <file.jloot> [-p port] Mount to RAM + HTTP API
  jloot search <file.jloot> <keyword> Search in-memory
`)
		os.Exit(1)
	}

	cmd, input := os.Args[1], os.Args[2]
	output := flagVal("-o")

	switch cmd {
	case "pack":
		pack(input, output)
	case "unpack":
		unpack(input, output)
	case "info":
		info(input)
	case "mount":
		port := flagVal("-p")
		if port == "" { port = "7700" }
		mountAndServe(input, port)
	case "search":
		if len(os.Args) < 4 { die("usage: jloot search <file.jloot> <keyword>") }
		searchCmd(input, os.Args[3])
	default:
		die("unknown: %s", cmd)
	}
}

// ─── PACK ──────────────────────────────────────────
func pack(src, dst string) {
	start := time.Now()
	stat, err := os.Stat(src)
	must(err, "stat")

	isDir := stat.IsDir()

	type entry struct {
		path    string
		content []byte
	}

	var entries []entry
	var totalOrig int

	if isDir {
		filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(src, p)
			rel = filepath.ToSlash(rel)
			data, err := os.ReadFile(p)
			if err != nil {
				return nil
			}
			entries = append(entries, entry{rel, data})
			totalOrig += len(data)
			return nil
		})
	} else {
		data, err := os.ReadFile(src)
		must(err, "read")
		entries = append(entries, entry{filepath.Base(src), data})
		totalOrig = len(data)
	}

	if dst == "" {
		base := filepath.Base(src)
		if isDir {
			dst = base + ".jloot"
		} else {
			dst = strings.TrimSuffix(base, filepath.Ext(base)) + ".jloot"
		}
	}

	// Build raw payload
	var payload bytes.Buffer
	for _, e := range entries {
		pathBytes := []byte(e.path)
		binary.Write(&payload, binary.LittleEndian, uint16(len(pathBytes)))
		payload.Write(pathBytes)
		binary.Write(&payload, binary.LittleEndian, uint32(len(e.content)))
		payload.Write(e.content)
	}

	// Hash
	hash := sha256.Sum256(payload.Bytes())

	// Gzip
	var compressed bytes.Buffer
	gz, _ := gzip.NewWriterLevel(&compressed, gzip.BestCompression)
	gz.Write(payload.Bytes())
	gz.Close()

	// Write .jloot
	f, err := os.Create(dst)
	must(err, "create")
	defer f.Close()

	flags := byte(0x00)
	if isDir {
		flags = 0x01
	}

	f.Write(magic)
	f.Write([]byte{flags})
	binary.Write(f, binary.LittleEndian, uint32(len(entries)))
	binary.Write(f, binary.LittleEndian, uint32(totalOrig))
	f.Write(hash[:])
	binary.Write(f, binary.LittleEndian, time.Now().Unix())
	f.Write(compressed.Bytes())

	elapsed := time.Since(start)
	fi, _ := os.Stat(dst)
	ratio := float64(fi.Size()) / float64(totalOrig) * 100

	fmt.Fprintf(os.Stderr, "✅ Packed: %s → %s\n", src, dst)
	fmt.Fprintf(os.Stderr, "   Files:    %d\n", len(entries))
	fmt.Fprintf(os.Stderr, "   Original: %s\n", humanSize(int64(totalOrig)))
	fmt.Fprintf(os.Stderr, "   Packed:   %s (%.1f%%)\n", humanSize(fi.Size()), ratio)
	fmt.Fprintf(os.Stderr, "   SHA-256:  %s\n", hex.EncodeToString(hash[:8]))
	fmt.Fprintf(os.Stderr, "   Time:     %s\n", elapsed.Round(time.Millisecond))
}

// ─── UNPACK ────────────────────────────────────────
func unpack(src, dst string) {
	start := time.Now()
	data, err := os.ReadFile(src)
	must(err, "read")

	if !bytes.HasPrefix(data, magic) {
		die("not a .jloot v2 file")
	}

	pos := len(magic)
	flags := data[pos]; pos++
	fileCount := binary.LittleEndian.Uint32(data[pos:]); pos += 4
	origSize := binary.LittleEndian.Uint32(data[pos:]); pos += 4
	storedHash := data[pos:pos+32]; pos += 32
	pos += 8 // skip timestamp
	compressed := data[pos:]

	// Decompress
	gz, err := gzip.NewReader(bytes.NewReader(compressed))
	must(err, "gzip")
	payload, err := io.ReadAll(gz)
	must(err, "decompress")
	gz.Close()

	// Verify
	hash := sha256.Sum256(payload)
	if !bytes.Equal(hash[:], storedHash) {
		die("❌ integrity check FAILED")
	}

	// Parse entries
	r := bytes.NewReader(payload)
	restored := 0

	for i := uint32(0); i < fileCount; i++ {
		var pathLen uint16
		binary.Read(r, binary.LittleEndian, &pathLen)
		pathBuf := make([]byte, pathLen)
		r.Read(pathBuf)
		relPath := string(pathBuf)

		var contentLen uint32
		binary.Read(r, binary.LittleEndian, &contentLen)
		content := make([]byte, contentLen)
		r.Read(content)

		if dst != "" {
			outPath := filepath.Join(dst, filepath.FromSlash(relPath))
			os.MkdirAll(filepath.Dir(outPath), 0755)
			must(os.WriteFile(outPath, content, 0644), "write")
			restored++
		} else if flags == 0x00 {
			// Single file → stdout
			os.Stdout.Write(content)
		}
	}

	elapsed := time.Since(start)

	if dst != "" {
		fmt.Fprintf(os.Stderr, "✅ Unpacked: %s → %s/\n", src, dst)
		fmt.Fprintf(os.Stderr, "   Files: %d | Size: %s | Time: %s\n",
			restored, humanSize(int64(origSize)), elapsed.Round(time.Millisecond))
	}
}

// ─── INFO ──────────────────────────────────────────
func info(src string) {
	data, err := os.ReadFile(src)
	must(err, "read")

	if !bytes.HasPrefix(data, magic) {
		die("not a .jloot v2 file")
	}

	pos := len(magic)
	flags := data[pos]; pos++
	fileCount := binary.LittleEndian.Uint32(data[pos:]); pos += 4
	origSize := binary.LittleEndian.Uint32(data[pos:]); pos += 4
	hash := data[pos:pos+32]; pos += 32
	ts := int64(binary.LittleEndian.Uint64(data[pos:])); pos += 8

	fi, _ := os.Stat(src)
	mode := "single file"
	if flags == 0x01 { mode = "folder tree" }

	fmt.Printf("📦 %s\n", filepath.Base(src))
	fmt.Printf("   Mode:      %s\n", mode)
	fmt.Printf("   Files:     %d\n", fileCount)
	fmt.Printf("   Original:  %s\n", humanSize(int64(origSize)))
	fmt.Printf("   Packed:    %s (%.1f%%)\n", humanSize(fi.Size()),
		float64(fi.Size())/float64(origSize)*100)
	fmt.Printf("   SHA-256:   %s\n", hex.EncodeToString(hash[:8]))
	fmt.Printf("   Created:   %s\n", time.Unix(ts, 0).Format("2006-01-02 15:04:05"))
}

// ─── BRAIN (In-Memory Mount) ──────────────────────
type Brain struct {
	files    map[string][]byte
	invIndex map[string][]string // word → [paths] (역인덱스)
	tf       map[string]map[string]float64 // path → word → TF score
	idf      map[string]float64  // word → IDF score
	mu       sync.RWMutex
	jlootSrc string
}

func loadBrain(jlootPath string) (*Brain, time.Duration) {
	start := time.Now()
	data, err := os.ReadFile(jlootPath)
	must(err, "read")

	if !bytes.HasPrefix(data, magic) {
		die("not a .jloot v2 file")
	}

	pos := len(magic) + 1 // skip flags
	fileCount := binary.LittleEndian.Uint32(data[pos:]); pos += 4
	pos += 4  // orig size
	pos += 32 // hash
	pos += 8  // timestamp
	compressed := data[pos:]

	gz, err := gzip.NewReader(bytes.NewReader(compressed))
	must(err, "gzip")
	payload, _ := io.ReadAll(gz)
	gz.Close()

	brain := &Brain{
		files:    make(map[string][]byte, fileCount),
		invIndex: make(map[string][]string),
		tf:       make(map[string]map[string]float64),
		idf:      make(map[string]float64),
		jlootSrc: jlootPath,
	}

	r := bytes.NewReader(payload)
	for i := uint32(0); i < fileCount; i++ {
		var pathLen uint16
		binary.Read(r, binary.LittleEndian, &pathLen)
		pathBuf := make([]byte, pathLen)
		r.Read(pathBuf)
		path := string(pathBuf)

		var contentLen uint32
		binary.Read(r, binary.LittleEndian, &contentLen)
		content := make([]byte, contentLen)
		r.Read(content)

		brain.files[path] = content
	}

	// 역인덱스 구축
	buildIndex(brain)

	return brain, time.Since(start)
}

func (b *Brain) Get(path string) ([]byte, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	v, ok := b.files[path]
	return v, ok
}

func (b *Brain) List(dir string) []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	var result []string
	prefix := dir + "/"
	for k := range b.files {
		if strings.HasPrefix(k, prefix) {
			rest := k[len(prefix):]
			if !strings.Contains(rest, "/") {
				result = append(result, k)
			}
		}
	}
	return result
}

func (b *Brain) Search(keyword string) []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	// O(1) 역인덱스 조회 먼저
	if paths, ok := b.invIndex[keyword]; ok {
		return paths
	}
	// 폴백: 부분 매칭
	var results []string
	for path, content := range b.files {
		if strings.Contains(path, keyword) || strings.Contains(string(content), keyword) {
			results = append(results, path)
		}
	}
	return results
}

// TF-IDF 기반 관련도 검색 — 맹점 #3 해결
type SearchResult struct {
	Path  string  `json:"path"`
	Score float64 `json:"score"`
	Snippet string `json:"snippet"`
}

func (b *Brain) SearchRanked(query string) []SearchResult {
	b.mu.RLock()
	defer b.mu.RUnlock()
	words := tokenize(query)
	scores := make(map[string]float64)

	for _, w := range words {
		paths, ok := b.invIndex[w]
		if !ok { continue }
		idf := b.idf[w]
		for _, p := range paths {
			tfScore := 0.0
			if tf, ok := b.tf[p]; ok {
				tfScore = tf[w]
			}
			scores[p] += tfScore * idf
		}
	}

	var results []SearchResult
	for p, s := range scores {
		snippet := ""
		if c, ok := b.files[p]; ok {
			snippet = string(c)
			if len(snippet) > 80 { snippet = snippet[:80] + "..." }
		}
		results = append(results, SearchResult{p, s, snippet})
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Score > results[j].Score })
	if len(results) > 50 { results = results[:50] }
	return results
}

// 질문 → 최적 경로 자동 라우팅 — 맹점 #1 해결
func (b *Brain) Route(query string) []SearchResult {
	return b.SearchRanked(query)
}

// ─── 역인덱스 빌더 ────────────────────────────────
func buildIndex(b *Brain) {
	docCount := float64(len(b.files))
	df := make(map[string]int) // document frequency

	for path, content := range b.files {
		text := string(content)
		words := tokenize(text)
		// 경로에서도 토큰 추출
		pathWords := tokenize(strings.ReplaceAll(path, "/", " "))
		words = append(words, pathWords...)

		wc := make(map[string]int)
		for _, w := range words {
			wc[w]++
		}

		total := float64(len(words))
		if total == 0 { total = 1 }

		b.tf[path] = make(map[string]float64)
		seen := make(map[string]bool)
		for w, c := range wc {
			b.tf[path][w] = float64(c) / total
			b.invIndex[w] = append(b.invIndex[w], path)
			if !seen[w] { df[w]++; seen[w] = true }
		}
	}

	for w, d := range df {
		b.idf[w] = math.Log(docCount / float64(d))
	}
}

func tokenize(s string) []string {
	var tokens []string
	var buf strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			buf.WriteRune(r)
		} else {
			if buf.Len() > 0 {
				tokens = append(tokens, buf.String())
				buf.Reset()
			}
		}
	}
	if buf.Len() > 0 { tokens = append(tokens, buf.String()) }
	return tokens
}

func (b *Brain) HasBan(dir string) (string, bool) {
	for _, f := range b.List(dir) {
		parts := strings.Split(f, "/")
		name := parts[len(parts)-1]
		if strings.HasPrefix(name, "禁_") {
			return name, true
		}
	}
	return "", false
}

// ─── MOUNT + HTTP API ─────────────────────────────
func mountAndServe(jlootPath, port string) {
	brain, mountTime := loadBrain(jlootPath)

	fmt.Printf("🧠 Brain mounted in %s\n", mountTime.Round(time.Millisecond))
	fmt.Printf("   Files: %d\n", len(brain.files))
	fmt.Printf("   Index: %d unique words\n", len(brain.invIndex))
	fmt.Printf("   Source: %s\n", jlootPath)
	fmt.Printf("   API: http://localhost:%s\n\n", port)
	fmt.Println("Endpoints:")
	fmt.Println("   GET /get?path=<path>       → file content (51ns)")
	fmt.Println("   GET /search?q=<keyword>    → keyword search")
	fmt.Println("   GET /ranked?q=<query>      → TF-IDF ranked search")
	fmt.Println("   GET /route?q=<question>    → auto-route question → paths")
	fmt.Println("   GET /ban?dir=<dir>         → check 禁")
	fmt.Println("   GET /list?dir=<dir>        → list directory")
	fmt.Println("   GET /stats                 → brain stats")
	fmt.Println("   POST /materialize          → background folder creation")

	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Query().Get("path")
		t := time.Now()
		content, ok := brain.Get(path)
		d := time.Since(t)
		if !ok {
			w.WriteHeader(404)
			fmt.Fprintf(w, `{"error":"not found","path":%q}`, path)
			return
		}
		w.Header().Set("X-Lookup-Time", d.String())
		w.Write(content)
	})

	http.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		dir := r.URL.Query().Get("dir")
		t := time.Now()
		files := brain.List(dir)
		d := time.Since(t)
		w.Header().Set("X-Lookup-Time", d.String())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(files)
	})

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		t := time.Now()
		results := brain.Search(q)
		d := time.Since(t)
		w.Header().Set("X-Lookup-Time", d.String())
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"query":   q,
			"count":   len(results),
			"time":    d.String(),
			"results": results,
		}
		json.NewEncoder(w).Encode(resp)
	})

	http.HandleFunc("/ban", func(w http.ResponseWriter, r *http.Request) {
		dir := r.URL.Query().Get("dir")
		t := time.Now()
		banFile, blocked := brain.HasBan(dir)
		d := time.Since(t)
		w.Header().Set("X-Lookup-Time", d.String())
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"dir":     dir,
			"blocked": blocked,
			"ban":     banFile,
			"time":    d.String(),
		}
		json.NewEncoder(w).Encode(resp)
	})

	http.HandleFunc("/ranked", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		t := time.Now()
		results := brain.SearchRanked(q)
		d := time.Since(t)
		w.Header().Set("X-Lookup-Time", d.String())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"query":   q,
			"count":   len(results),
			"time":    d.String(),
			"results": results,
		})
	})

	http.HandleFunc("/route", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		t := time.Now()
		results := brain.Route(q)
		d := time.Since(t)
		w.Header().Set("X-Lookup-Time", d.String())
		w.Header().Set("Content-Type", "application/json")

		blocked := false
		banFile := ""
		if len(results) > 0 {
			dir := results[0].Path
			if idx := strings.LastIndex(dir, "/"); idx > 0 {
				dir = dir[:idx]
			}
			banFile, blocked = brain.HasBan(dir)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"query":   q,
			"blocked": blocked,
			"ban":     banFile,
			"routes":  results,
			"time":    d.String(),
		})
	})

	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"files":       len(brain.files),
			"index_words": len(brain.invIndex),
			"source":      brain.jlootSrc,
			"mount_time":  mountTime.String(),
		})
	})

	http.HandleFunc("/materialize", func(w http.ResponseWriter, r *http.Request) {
		outDir := r.URL.Query().Get("dir")
		if outDir == "" { outDir = "brain_phys" }
		w.Header().Set("Content-Type", "application/json")
		go materializeBrain(brain, outDir) // background
		json.NewEncoder(w).Encode(map[string]string{
			"status": "materializing in background",
			"dir":    outDir,
		})
	})
	// ─── /ask: Jloot 하네스 + Ollama LLM ───────────────
	http.HandleFunc("/ask", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		model := r.URL.Query().Get("model")
		if model == "" { model = "llama3.2:3b" }
		bare := r.URL.Query().Get("bare") == "true" // bare=true면 하네스 없이

		w.Header().Set("Content-Type", "application/json")
		start := time.Now()

		if bare {
			// ━━━ 하네스 없이: LLM에 직접 질문 ━━━
			llmResp, llmTime, err := callOllama(model, q, "")
			if err != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": err.Error(),
				})
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"mode":     "bare (no harness)",
				"query":    q,
				"model":    model,
				"answer":   llmResp,
				"llm_time": llmTime.String(),
				"total":    time.Since(start).String(),
			})
			return
		}

		// ━━━ 1단계: Jloot 라우팅 + 거버넌스 (1ms) ━━━
		routes := brain.Route(q)
		routeTime := time.Since(start)

		// 禁 체크
		blocked := false
		banFile := ""
		var context strings.Builder
		for _, route := range routes {
			dir := route.Path
			if idx := strings.LastIndex(dir, "/"); idx > 0 {
				dir = dir[:idx]
			}
			if ban, found := brain.HasBan(dir); found {
				blocked = true
				banFile = ban
				break
			}
			// 통과한 경로의 콘텐츠를 컨텍스트로 수집
			if content, ok := brain.Get(route.Path); ok {
				context.WriteString(fmt.Sprintf("[%s]\n%s\n\n", route.Path, string(content)))
			}
			if context.Len() > 2000 { break } // 컨텍스트 제한
		}

		if blocked {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"mode":       "harness (BLOCKED)",
				"query":      q,
				"blocked":    true,
				"ban":        banFile,
				"answer":     fmt.Sprintf("⛔ 이 질문은 거버넌스 규칙 [%s]에 의해 차단되었습니다.", banFile),
				"routes":     len(routes),
				"route_time": routeTime.String(),
				"llm_time":   "0s (호출 안 함)",
				"llm_cost":   "0원",
				"total":      time.Since(start).String(),
			})
			return
		}

		// ━━━ 2단계: 하네스 + JSON Constrained Decoding ━━━
		harness := `[NeuronFS 하네스]
반드시 아래 JSON 형식으로만 답하라. 다른 형식 절대 금지.
{"answer":"참고자료 구절만으로 자연스럽게 답한 내용","sources":["경로1.md","경로2.md"],"unanswerable":["답할수없는주제"]}
규칙: 참고자료에 없으면 answer를 비우고 unanswerable에 넣어라. 한국어.`
		prompt := fmt.Sprintf("%s\n\n[검색결과]\n%s\n\n질문: %s", harness, context.String(), q)
		llmResp, llmTime, err := callOllama(model, prompt, "", true)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": err.Error(), "routes": len(routes), "route_time": routeTime.String(),
			})
			return
		}

		// ━━━ 3단계: 구조적 검증 (JSON 파싱 + 출처 대조) ━━━
		validPaths := make(map[string]bool)
		for _, route := range routes {
			validPaths[route.Path] = true
		}

		var structured struct {
			Answer       string   `json:"answer"`
			Sources      []string `json:"sources"`
			Unanswerable []string `json:"unanswerable"`
		}
		jsonParsed := true
		if err := json.Unmarshal([]byte(llmResp), &structured); err != nil {
			// JSON 파싱 실패 → 폴백: 정규식으로 출처 추출
			jsonParsed = false
			structured.Answer = llmResp
			citationRe := regexp.MustCompile(`\[([^\]]+\.md)\]`)
			for _, m := range citationRe.FindAllStringSubmatch(llmResp, -1) {
				if len(m) > 1 { structured.Sources = append(structured.Sources, m[1]) }
			}
		}

		verified, hallucinated := 0, 0
		var hallucinatedPaths []string
		for _, src := range structured.Sources {
			if validPaths[src] { verified++ } else { hallucinated++; hallucinatedPaths = append(hallucinatedPaths, src) }
		}

		// 환각 출처 마킹
		filteredAnswer := structured.Answer
		for _, hp := range hallucinatedPaths {
			filteredAnswer = strings.ReplaceAll(filteredAnswer, hp, "⚠️"+hp)
		}

		totalCitations := len(structured.Sources)
		trustScore := 100.0
		if totalCitations > 0 { trustScore = float64(verified) / float64(totalCitations) * 100 }

		json.NewEncoder(w).Encode(map[string]interface{}{
			"mode":       "harness (Jloot + JSON Constrained + Validation)",
			"query":      q,
			"model":      model,
			"blocked":    false,
			"routes":     len(routes),
			"route_time": routeTime.String(),
			"llm_time":   llmTime.String(),
			"answer":     filteredAnswer,
			"validation": map[string]interface{}{
				"json_parsed":            jsonParsed,
				"total_citations":        totalCitations,
				"verified_citations":     verified,
				"hallucinated_citations": hallucinated,
				"hallucinated_paths":     hallucinatedPaths,
				"unanswerable":           structured.Unanswerable,
				"trust_score":            fmt.Sprintf("%.0f%%", trustScore),
			},
			"total": time.Since(start).String(),
		})
	})

	fmt.Println("   GET /ask?q=<question>&model=<model>&bare=true|false")

	must(http.ListenAndServe(":"+port, nil), "http")
}

// ─── Ollama API 호출 ───────────────────────────────
func callOllama(model, prompt, system string, jsonMode ...bool) (string, time.Duration, error) {
	start := time.Now()

	body := map[string]interface{}{
		"model":  model,
		"prompt": prompt,
		"stream": false,
	}
	if system != "" {
		body["system"] = system
	}
	if len(jsonMode) > 0 && jsonMode[0] {
		body["format"] = "json"
	}

	jsonBody, _ := json.Marshal(body)
	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return "", 0, fmt.Errorf("ollama 연결 실패: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Response string `json:"response"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	return result.Response, time.Since(start), nil
}

// 백그라운드 난수 폴더 물질화
func materializeBrain(brain *Brain, outDir string) {
	start := time.Now()
	brain.mu.RLock()
	defer brain.mu.RUnlock()

	// 매핑 테이블
	mapping := make(map[string]string) // realPath → randHex

	for path, content := range brain.files {
		parts := strings.Split(path, "/")
		randParts := make([]string, len(parts))
		for i := range parts {
			key := strings.Join(parts[:i+1], "/")
			if rp, ok := mapping[key]; ok {
				randParts[i] = rp
			} else {
				rh := randHex8()
				mapping[key] = rh
				randParts[i] = rh
			}
		}
		dir := filepath.Join(outDir, filepath.Join(randParts[:len(randParts)-1]...))
		os.MkdirAll(dir, 0755)
		os.WriteFile(filepath.Join(dir, randParts[len(randParts)-1]), content, 0644)
	}

	// 매핑 테이블 저장 (암호화 대상)
	mapData, _ := json.MarshalIndent(mapping, "", "  ")
	os.WriteFile(filepath.Join(outDir, "_mapping.json"), mapData, 0644)

	fmt.Printf("✅ Materialized %d files → %s (%s)\n",
		len(brain.files), outDir, time.Since(start).Round(time.Millisecond))
}

func randHex8() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// ─── SEARCH CMD ───────────────────────────────────
func searchCmd(jlootPath, keyword string) {
	brain, mountTime := loadBrain(jlootPath)
	fmt.Printf("🧠 Mounted in %s (%d files)\n", mountTime.Round(time.Millisecond), len(brain.files))

	t := time.Now()
	results := brain.Search(keyword)
	d := time.Since(t)

	fmt.Printf("🔍 \"%s\" → %d results in %s\n\n", keyword, len(results), d)
	for i, r := range results {
		if i >= 20 { fmt.Printf("... and %d more\n", len(results)-20); break }
		content, _ := brain.Get(r)
		text := string(content)
		if len(text) > 60 { text = text[:60] + "..." }
		fmt.Printf("  %s → %s\n", r, text)
	}
}

// ─── HELPERS ───────────────────────────────────────
func humanSize(b int64) string {
	switch {
	case b >= 1<<30: return fmt.Sprintf("%.1f GB", float64(b)/(1<<30))
	case b >= 1<<20: return fmt.Sprintf("%.1f MB", float64(b)/(1<<20))
	case b >= 1<<10: return fmt.Sprintf("%.1f KB", float64(b)/(1<<10))
	default: return fmt.Sprintf("%d B", b)
	}
}

func flagVal(name string) string {
	for i, a := range os.Args {
		if a == name && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
	}
	return ""
}

func must(err error, ctx string) {
	if err != nil { die("%s: %v", ctx, err) }
}

func die(f string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "💀 "+f+"\n", a...)
	os.Exit(1)
}
