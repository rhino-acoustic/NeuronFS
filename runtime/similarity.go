package main

// ━━━ similarity.go ━━━
// Module: Hybrid Similarity Engine
//
// PROVIDES:
//   tokenize, stem, jaccardSimilarity, hybridSimilarity,
//   cosineTokens, levenshteinDistance, extractPrefix,
//   newtonSqrt, maxInt, minInt
//
// CONSUMED BY:
//   main.go         → growNeuron() calls hybridSimilarity()
//   main.go         → deduplicateNeurons() calls hybridSimilarity()
//   mcp_server.go   → health_check calls hybridSimilarity()
//
// DEPENDS ON:
//   strings (stdlib only — leaf module, no cross-file dependencies)

import "strings"

// tokenize splits a snake_case neuron name into stemmed lowercase tokens
// "no_console_logging" → {"no", "console", "log"}
func tokenize(name string) []string {
	// 밑줄과 공백 모두 분리자로 처리
	normalized := strings.ReplaceAll(strings.ToLower(name), "_", " ")

	// 한자 접두어 분리 (必/禁/推/絶 등)
	hanjaMap := map[rune]string{
		'必': "필수", '禁': "금지", '推': "추천", '絶': "절대",
	}
	var hanjaTokens []string
	cleanRunes := []rune(normalized)
	for i := 0; i < len(cleanRunes); i++ {
		if ko, ok := hanjaMap[cleanRunes[i]]; ok {
			hanjaTokens = append(hanjaTokens, ko)
			cleanRunes = append(cleanRunes[:i], cleanRunes[i+1:]...)
			i--
		}
	}
	normalized = string(cleanRunes)

	parts := strings.Fields(normalized)
	var tokens []string
	tokens = append(tokens, hanjaTokens...)

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		// 한글 감지: rune 범위 0xAC00-0xD7A3
		runes := []rune(p)
		isKorean := false
		for _, r := range runes {
			if r >= 0xAC00 && r <= 0xD7A3 {
				isKorean = true
				break
			}
		}

		if isKorean {
			// 한글 2-gram: "한국어사고" → ["한국", "국어", "어사", "사고"]
			koreanRunes := []rune{}
			for _, r := range runes {
				if r >= 0xAC00 && r <= 0xD7A3 {
					koreanRunes = append(koreanRunes, r)
				}
			}
			if len(koreanRunes) >= 2 {
				for i := 0; i < len(koreanRunes)-1; i++ {
					tokens = append(tokens, string(koreanRunes[i:i+2]))
				}
			}
			// 전체 한글 문자열도 토큰으로 (짧은 단어 매칭 보강)
			if len(koreanRunes) > 0 {
				tokens = append(tokens, string(koreanRunes))
			}
		} else {
			// 영문 stemming
			p = stem(p)
			tokens = append(tokens, p)
		}
	}
	return tokens
}

// stem applies minimal suffix stripping for merge matching
// Not a full Porter stemmer — just handles common AI naming patterns
func stem(word string) string {
	// Order matters: check longer suffixes first
	suffixes := []string{"ation", "ting", "ning", "ding", "ring", "sing", "ling", "ping", "ging", "ing", "ied", "ies", "ness", "ment", "able", "ible", "ful", "less", "ous", "ive", "ed"}
	for _, s := range suffixes {
		if len(word) > len(s)+2 && strings.HasSuffix(word, s) {
			return word[:len(word)-len(s)]
		}
	}
	// Trailing 's' (plural) — only if word is 4+ chars
	if len(word) >= 4 && strings.HasSuffix(word, "s") && !strings.HasSuffix(word, "ss") {
		return word[:len(word)-1]
	}
	return word
}

// jaccardSimilarity computes |A∩B| / |A∪B| between two token sets (legacy)
func jaccardSimilarity(a, b []string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	setA := make(map[string]bool)
	for _, t := range a {
		setA[t] = true
	}
	setB := make(map[string]bool)
	for _, t := range b {
		setB[t] = true
	}
	intersection := 0
	for t := range setA {
		if setB[t] {
			intersection++
		}
	}
	union := len(setA)
	for t := range setB {
		if !setA[t] {
			union++
		}
	}
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

// ─── Hybrid Similarity Engine (Jaccard 대체) ───
// Cosine Bigram (60%) + Levenshtein Ratio (40%)
// - Cosine: 빈도 고려, 부분 매칭, 한글 2-gram에 최적
// - Levenshtein: 편집 거리, 짧은 문자열 비교에 최적
// - 접두어(禁/推/必/核) 제거 후 비교 (접두어가 다르면 별도 처리)

func hybridSimilarity(tokensA, tokensB []string) float64 {
	if len(tokensA) == 0 || len(tokensB) == 0 {
		return 0
	}
	// Cosine similarity over token frequency vectors
	cosSim := cosineTokens(tokensA, tokensB)
	// Levenshtein over joined strings
	strA := strings.Join(tokensA, "")
	strB := strings.Join(tokensB, "")
	levRatio := 1.0 - float64(levenshteinDistance([]rune(strA), []rune(strB)))/float64(maxInt(len([]rune(strA)), len([]rune(strB))))
	if levRatio < 0 {
		levRatio = 0
	}
	return cosSim*0.6 + levRatio*0.4
}

// cosineTokens: token frequency vector cosine similarity
func cosineTokens(a, b []string) float64 {
	freqA := make(map[string]float64)
	freqB := make(map[string]float64)
	for _, t := range a {
		freqA[t]++
	}
	for _, t := range b {
		freqB[t]++
	}
	// dot product, magnitudes
	dot := 0.0
	magA := 0.0
	magB := 0.0
	for t, va := range freqA {
		magA += va * va
		if vb, ok := freqB[t]; ok {
			dot += va * vb
		}
	}
	for _, vb := range freqB {
		magB += vb * vb
	}
	import_math := magA * magB
	if import_math == 0 {
		return 0
	}
	// manual sqrt: Newton's method (avoid math import conflict)
	sqrtVal := newtonSqrt(import_math)
	if sqrtVal == 0 {
		return 0
	}
	return dot / sqrtVal
}

// levenshteinDistance: edit distance between two rune slices
func levenshteinDistance(a, b []rune) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	// Single-row DP
	prev := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		curr := make([]int, lb+1)
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			ins := curr[j-1] + 1
			del := prev[j] + 1
			sub := prev[j-1] + cost
			curr[j] = minInt(ins, minInt(del, sub))
		}
		prev = curr
	}
	return prev[lb]
}

func newtonSqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x
	for i := 0; i < 20; i++ {
		z = (z + x/z) / 2
	}
	return z
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// extractPrefix returns the Han prefix (禁/推/必/核/絶/最/有) from a neuron name, or ""
func extractPrefix(name string) string {
	runes := []rune(name)
	if len(runes) == 0 {
		return ""
	}
	first := runes[0]
	switch first {
	case '禁', '推', '必', '核', '絶', '最', '有':
		return string(first)
	}
	return ""
}
