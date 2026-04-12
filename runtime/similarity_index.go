package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"unicode"
)

// ============================================================================
// Module: Neuron Similarity Index — Go-Native TF-IDF (Phase 44)
// Builds a lightweight vector index over .neuron files for semantic search
// without external embedding model dependencies.
// ============================================================================

// SimilarityIndex holds the TF-IDF vectors for all indexed neurons
type SimilarityIndex struct {
	mu       sync.RWMutex
	docs     []indexedDoc          // All indexed documents
	idf      map[string]float64    // Inverse Document Frequency per term
	docCount int
}

type indexedDoc struct {
	Path   string             // Relative path from brainRoot
	TfIdf  map[string]float64 // TF-IDF vector
	Norm   float64            // Pre-computed L2 norm for fast cosine
}

// SimilarResult represents a search result with similarity score
type SimilarResult struct {
	Path  string  `json:"path"`
	Score float64 `json:"score"`
}

var globalSimilarityIndex *SimilarityIndex

// BuildSimilarityIndex scans all .neuron files and builds a TF-IDF index
func BuildSimilarityIndex(brainRoot string) *SimilarityIndex {
	if brainRoot == "" {
		return nil
	}

	idx := &SimilarityIndex{
		idf: make(map[string]float64),
	}

	// Phase 1: Collect all documents and compute Term Frequency
	var allDocs []struct {
		path  string
		tf    map[string]int
		total int
	}

	_ = filepath.Walk(brainRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".neuron" {
			return nil
		}
		// Skip archives
		if strings.Contains(path, "_archive") || strings.Contains(path, "_quarantine") {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}

		tokens := tokenizeContent(string(data))
		if len(tokens) == 0 {
			return nil
		}

		tf := make(map[string]int)
		for _, t := range tokens {
			tf[t]++
		}

		relPath, _ := filepath.Rel(brainRoot, path)
		allDocs = append(allDocs, struct {
			path  string
			tf    map[string]int
			total int
		}{relPath, tf, len(tokens)})

		return nil
	})

	idx.docCount = len(allDocs)
	if idx.docCount == 0 {
		globalSimilarityIndex = idx
		return idx
	}

	// Phase 2: Compute IDF
	docFreq := make(map[string]int)
	for _, doc := range allDocs {
		seen := make(map[string]bool)
		for term := range doc.tf {
			if !seen[term] {
				docFreq[term]++
				seen[term] = true
			}
		}
	}

	for term, df := range docFreq {
		idx.idf[term] = math.Log(float64(idx.docCount+1) / float64(df+1))
	}

	// Phase 3: Compute TF-IDF vectors with L2 norms
	for _, doc := range allDocs {
		tfidf := make(map[string]float64)
		var normSum float64

		for term, count := range doc.tf {
			tf := float64(count) / float64(doc.total)
			val := tf * idx.idf[term]
			tfidf[term] = val
			normSum += val * val
		}

		norm := math.Sqrt(normSum)
		if norm == 0 {
			norm = 1
		}

		idx.docs = append(idx.docs, indexedDoc{
			Path:  doc.path,
			TfIdf: tfidf,
			Norm:  norm,
		})
	}

	globalSimilarityIndex = idx
	fmt.Printf("  🔍 SIMILARITY INDEX: %d neurons indexed (%d unique terms)\n", idx.docCount, len(idx.idf))

	return idx
}

// QuerySimilar finds the top-K most similar neurons to a query string
func QuerySimilar(query string, topK int) []SimilarResult {
	if globalSimilarityIndex == nil || globalSimilarityIndex.docCount == 0 {
		return nil
	}

	idx := globalSimilarityIndex
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	// Tokenize and compute query TF-IDF
	tokens := tokenizeContent(query)
	if len(tokens) == 0 {
		return nil
	}

	queryTf := make(map[string]int)
	for _, t := range tokens {
		queryTf[t]++
	}

	queryVec := make(map[string]float64)
	var queryNormSum float64
	for term, count := range queryTf {
		tf := float64(count) / float64(len(tokens))
		idfVal, ok := idx.idf[term]
		if !ok {
			continue // Unknown term
		}
		val := tf * idfVal
		queryVec[term] = val
		queryNormSum += val * val
	}

	queryNorm := math.Sqrt(queryNormSum)
	if queryNorm == 0 {
		return nil
	}

	// Compute cosine similarity against all documents
	var results []SimilarResult
	for _, doc := range idx.docs {
		dotProduct := 0.0
		for term, qVal := range queryVec {
			if dVal, ok := doc.TfIdf[term]; ok {
				dotProduct += qVal * dVal
			}
		}

		similarity := dotProduct / (queryNorm * doc.Norm)
		if similarity > 0.01 { // Minimum threshold
			results = append(results, SimilarResult{
				Path:  doc.Path,
				Score: math.Round(similarity*1000) / 1000,
			})
		}
	}

	// Sort by descending similarity
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > topK {
		results = results[:topK]
	}

	return results
}

// tokenizeContent splits text into lowercase word tokens, filtering noise
func tokenizeContent(text string) []string {
	var tokens []string
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	for _, w := range words {
		w = strings.ToLower(w)
		if len(w) >= 2 && len(w) <= 30 { // Skip single chars and very long tokens
			tokens = append(tokens, w)
		}
	}
	return tokens
}
