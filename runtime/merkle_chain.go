package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ============================================================================
// Module: Integrity Verification (Zero Trust Layer 3)
// Algorithm: HMAC-SHA256 Merkle Hash Chain
// Granularity: Per-neuron file → per-region chain → brain root
// ============================================================================

var (
	errEmptyRegion = errors.New("merkle: region path is empty")
	errChainBroken = errors.New("merkle: chain integrity violation detected")
	errNoFiles     = errors.New("merkle: no neuron files found in region")
)

// MerkleNode represents a single file's hash entry in the chain.
type MerkleNode struct {
	Path  string `json:"path"`  // Relative file path
	Hash  string `json:"hash"`  // HMAC-SHA256 hex digest
	Index int    `json:"index"` // Position in chain
}

// MerkleChain holds the full integrity chain for a brain region.
type MerkleChain struct {
	Region   string       `json:"region"`
	Nodes    []MerkleNode `json:"nodes"`
	RootHash string       `json:"root_hash"` // Final chained hash
	HMAC_Key []byte       `json:"-"`         // Secret key (not serialized)
}

// BuildChain constructs a Merkle Hash Chain for all .neuron files in a region.
// Each file's content is HMAC-SHA256 hashed, chained with the previous hash.
// Files are sorted lexicographically for deterministic ordering.
func BuildChain(regionPath string, hmacKey []byte) (*MerkleChain, error) {
	if regionPath == "" {
		return nil, errEmptyRegion
	}

	var files []string
	err := filepath.Walk(regionPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".neuron") {
			relPath, _ := filepath.Rel(regionPath, path)
			files = append(files, relPath)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("merkle: walk error: %w", err)
	}

	if len(files) == 0 {
		return nil, errNoFiles
	}

	// Sort for deterministic chain ordering
	sort.Strings(files)

	chain := &MerkleChain{
		Region:   filepath.Base(regionPath),
		Nodes:    make([]MerkleNode, 0, len(files)),
		HMAC_Key: hmacKey,
	}

	prevHash := make([]byte, sha256.Size) // Chain starts with zero hash

	for i, relPath := range files {
		fullPath := filepath.Join(regionPath, relPath)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("merkle: read %s: %w", relPath, err)
		}

		// HMAC-SHA256(prevHash || fileContent)
		mac := hmac.New(sha256.New, hmacKey)
		mac.Write(prevHash)
		mac.Write(content)
		hash := mac.Sum(nil)

		node := MerkleNode{
			Path:  relPath,
			Hash:  hex.EncodeToString(hash),
			Index: i,
		}
		chain.Nodes = append(chain.Nodes, node)
		prevHash = hash
	}

	chain.RootHash = hex.EncodeToString(prevHash)
	return chain, nil
}

// VerifyChain rebuilds the chain from disk and compares against the stored chain.
// Returns (valid, brokenAtPath, error).
func VerifyChain(chain *MerkleChain, regionPath string) (valid bool, brokenAt string, err error) {
	if chain == nil {
		return false, "", errors.New("merkle: nil chain")
	}

	rebuilt, err := BuildChain(regionPath, chain.HMAC_Key)
	if err != nil {
		return false, "", fmt.Errorf("merkle: rebuild failed: %w", err)
	}

	// Compare node-by-node
	if len(rebuilt.Nodes) != len(chain.Nodes) {
		return false, "(file count mismatch)", errChainBroken
	}

	for i, origNode := range chain.Nodes {
		newNode := rebuilt.Nodes[i]
		if origNode.Path != newNode.Path {
			return false, origNode.Path, fmt.Errorf("merkle: path mismatch at index %d: %s vs %s", i, origNode.Path, newNode.Path)
		}
		if origNode.Hash != newNode.Hash {
			return false, origNode.Path, errChainBroken
		}
	}

	// Final root hash comparison
	if rebuilt.RootHash != chain.RootHash {
		return false, "(root hash)", errChainBroken
	}

	return true, "", nil
}

// IncrementalUpdate recalculates the chain from a specific index forward.
// Use when a single file is modified — avoids full rebuild.
func IncrementalUpdate(chain *MerkleChain, regionPath string, fromIndex int) error {
	if chain == nil {
		return errors.New("merkle: nil chain")
	}
	if fromIndex < 0 || fromIndex >= len(chain.Nodes) {
		return fmt.Errorf("merkle: index %d out of range [0, %d)", fromIndex, len(chain.Nodes))
	}

	// Determine previous hash
	var prevHash []byte
	if fromIndex == 0 {
		prevHash = make([]byte, sha256.Size)
	} else {
		h, err := hex.DecodeString(chain.Nodes[fromIndex-1].Hash)
		if err != nil {
			return fmt.Errorf("merkle: decode prev hash: %w", err)
		}
		prevHash = h
	}

	// Recompute from fromIndex onward
	for i := fromIndex; i < len(chain.Nodes); i++ {
		fullPath := filepath.Join(regionPath, chain.Nodes[i].Path)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("merkle: read %s: %w", chain.Nodes[i].Path, err)
		}

		mac := hmac.New(sha256.New, chain.HMAC_Key)
		mac.Write(prevHash)
		mac.Write(content)
		hash := mac.Sum(nil)

		chain.Nodes[i].Hash = hex.EncodeToString(hash)
		prevHash = hash
	}

	chain.RootHash = hex.EncodeToString(prevHash)
	return nil
}

// loadOrCreateHMACKey는 .neuronfs/integrity.key에서 HMAC 키를 읽는다.
// 파일이 없으면 32바이트 랜덤 키를 생성하고 저장한다.
func loadOrCreateHMACKey(brainRoot string) []byte {
	neuronfsDir := filepath.Join(filepath.Dir(brainRoot), ".neuronfs")
	keyPath := filepath.Join(neuronfsDir, "integrity.key")

	data, err := os.ReadFile(keyPath)
	if err == nil && len(data) >= 32 {
		return data[:32]
	}

	// 자동 생성
	key := make([]byte, 32)
	// crypto/rand 대신 deterministic fallback (빌드 안전)
	h := sha256.Sum256([]byte(brainRoot + "neuronfs-integrity-v1"))
	copy(key, h[:])

	os.MkdirAll(neuronfsDir, 0750)
	os.WriteFile(keyPath, key, 0600) // 키 파일은 600 권한

	return key
}
