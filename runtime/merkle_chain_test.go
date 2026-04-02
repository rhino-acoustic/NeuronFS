package main

import (
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"
)

func makeTestRegion(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, name)
		os.MkdirAll(filepath.Dir(path), 0o755)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("writeFile %s: %v", name, err)
		}
	}
	return dir
}

func testHMACKey() []byte {
	key := make([]byte, 32)
	rand.Read(key)
	return key
}

// ---------------------------------------------------------------------------
// Core Tests
// ---------------------------------------------------------------------------

func TestBuildChain_Basic(t *testing.T) {
	dir := makeTestRegion(t, map[string]string{
		"canon" + string(os.PathSeparator) + "1.neuron": "never_use_fallback counter=103",
		"canon" + string(os.PathSeparator) + "2.neuron": "no_simulation counter=100",
		"reflexes" + string(os.PathSeparator) + "1.neuron": "self_debug counter=100",
	})

	key := testHMACKey()
	chain, err := BuildChain(dir, key)
	if err != nil {
		t.Fatalf("BuildChain: %v", err)
	}

	if len(chain.Nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(chain.Nodes))
	}
	if chain.RootHash == "" {
		t.Fatal("root hash is empty")
	}

	t.Logf("OK: chain built (%d nodes, root=%s...)", len(chain.Nodes), chain.RootHash[:16])
}

func TestVerifyChain_Intact(t *testing.T) {
	dir := makeTestRegion(t, map[string]string{
		"rule1.neuron": "dopamine_reward counter=50",
		"rule2.neuron": "detect_urgency counter=30",
	})

	key := testHMACKey()
	chain, err := BuildChain(dir, key)
	if err != nil {
		t.Fatalf("BuildChain: %v", err)
	}

	valid, brokenAt, err := VerifyChain(chain, dir)
	if err != nil {
		t.Fatalf("VerifyChain: %v", err)
	}
	if !valid {
		t.Fatalf("chain should be valid, broken at: %s", brokenAt)
	}
	t.Log("OK: intact chain verified successfully")
}

func TestVerifyChain_TamperedFile(t *testing.T) {
	dir := makeTestRegion(t, map[string]string{
		"alpha.neuron": "original content A",
		"beta.neuron":  "original content B",
		"gamma.neuron": "original content C",
	})

	key := testHMACKey()
	chain, err := BuildChain(dir, key)
	if err != nil {
		t.Fatalf("BuildChain: %v", err)
	}

	// Tamper with beta.neuron
	tamperedPath := filepath.Join(dir, "beta.neuron")
	if err := os.WriteFile(tamperedPath, []byte("TAMPERED CONTENT"), 0o644); err != nil {
		t.Fatalf("tamper write: %v", err)
	}

	valid, brokenAt, err := VerifyChain(chain, dir)
	if valid {
		t.Fatal("tampered chain should NOT be valid")
	}
	if err != errChainBroken {
		t.Fatalf("expected errChainBroken, got: %v", err)
	}
	if brokenAt != "beta.neuron" {
		t.Fatalf("expected broken at beta.neuron, got: %s", brokenAt)
	}
	t.Logf("OK: tampered file detected at: %s", brokenAt)
}

func TestVerifyChain_FileAdded(t *testing.T) {
	dir := makeTestRegion(t, map[string]string{
		"a.neuron": "file A",
		"b.neuron": "file B",
	})

	key := testHMACKey()
	chain, err := BuildChain(dir, key)
	if err != nil {
		t.Fatalf("BuildChain: %v", err)
	}

	// Add a new file
	os.WriteFile(filepath.Join(dir, "c.neuron"), []byte("new file"), 0o644)

	valid, brokenAt, err := VerifyChain(chain, dir)
	if valid {
		t.Fatal("chain with added file should NOT be valid")
	}
	t.Logf("OK: file addition detected (broken=%s, err=%v)", brokenAt, err)
}

func TestVerifyChain_FileDeleted(t *testing.T) {
	dir := makeTestRegion(t, map[string]string{
		"x.neuron": "file X",
		"y.neuron": "file Y",
		"z.neuron": "file Z",
	})

	key := testHMACKey()
	chain, err := BuildChain(dir, key)
	if err != nil {
		t.Fatalf("BuildChain: %v", err)
	}

	// Delete y.neuron
	os.Remove(filepath.Join(dir, "y.neuron"))

	valid, _, err := VerifyChain(chain, dir)
	if valid {
		t.Fatal("chain with deleted file should NOT be valid")
	}
	t.Logf("OK: file deletion detected (err=%v)", err)
}

func TestBuildChain_EmptyRegion(t *testing.T) {
	dir := t.TempDir()
	key := testHMACKey()

	_, err := BuildChain(dir, key)
	if err != errNoFiles {
		t.Fatalf("expected errNoFiles, got: %v", err)
	}
	t.Log("OK: empty region correctly rejected")
}

func TestBuildChain_EmptyPath(t *testing.T) {
	key := testHMACKey()
	_, err := BuildChain("", key)
	if err != errEmptyRegion {
		t.Fatalf("expected errEmptyRegion, got: %v", err)
	}
	t.Log("OK: empty path correctly rejected")
}

func TestBuildChain_Deterministic(t *testing.T) {
	files := map[string]string{
		"a.neuron": "content A",
		"b.neuron": "content B",
		"c.neuron": "content C",
	}

	key := testHMACKey()

	dir1 := makeTestRegion(t, files)
	chain1, _ := BuildChain(dir1, key)

	dir2 := makeTestRegion(t, files)
	chain2, _ := BuildChain(dir2, key)

	if chain1.RootHash != chain2.RootHash {
		t.Fatal("identical regions should produce identical root hashes")
	}
	t.Logf("OK: deterministic chain verified (root=%s...)", chain1.RootHash[:16])
}

func TestIncrementalUpdate(t *testing.T) {
	dir := makeTestRegion(t, map[string]string{
		"a.neuron": "content A",
		"b.neuron": "content B",
		"c.neuron": "content C",
	})

	key := testHMACKey()
	chain, _ := BuildChain(dir, key)
	originalRoot := chain.RootHash

	// Modify b.neuron
	os.WriteFile(filepath.Join(dir, "b.neuron"), []byte("MODIFIED B"), 0o644)

	// Incremental update from index 1 (b.neuron)
	err := IncrementalUpdate(chain, dir, 1)
	if err != nil {
		t.Fatalf("IncrementalUpdate: %v", err)
	}

	if chain.RootHash == originalRoot {
		t.Fatal("root hash should change after modification")
	}

	// Verify the updated chain
	valid, brokenAt, err := VerifyChain(chain, dir)
	if err != nil {
		t.Fatalf("VerifyChain after update: %v", err)
	}
	if !valid {
		t.Fatalf("updated chain should be valid, broken at: %s", brokenAt)
	}
	t.Logf("OK: incremental update verified (old=%s..., new=%s...)", originalRoot[:16], chain.RootHash[:16])
}

func TestDifferentKeys_DifferentHashes(t *testing.T) {
	dir := makeTestRegion(t, map[string]string{
		"n.neuron": "neuron data",
	})

	key1 := testHMACKey()
	key2 := testHMACKey()

	chain1, _ := BuildChain(dir, key1)
	chain2, _ := BuildChain(dir, key2)

	if chain1.RootHash == chain2.RootHash {
		t.Fatal("different HMAC keys should produce different root hashes")
	}
	t.Log("OK: different keys produce different hashes (key isolation verified)")
}
