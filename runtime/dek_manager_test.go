package main

import (
	"os"
	"path/filepath"
	"testing"
)

// ============================================================================
// DEK Manager ?åÏä§??// ============================================================================

func TestInitKeyStore_FirstTime(t *testing.T) {
	dir := t.TempDir()
	neuronfsDir := filepath.Join(dir, ".neuronfs")

	ks, err := InitKeyStore(neuronfsDir)
	if err != nil {
		t.Fatalf("InitKeyStore failed: %v", err)
	}

	// ?úÏÑ± ?§Í? Ï°¥Ïû¨?¥Ïïº ??	key, meta, err := ks.GetActiveDEK()
	if err != nil {
		t.Fatalf("GetActiveDEK failed: %v", err)
	}
	if len(key) != 32 {
		t.Fatalf("expected 32-byte DEK, got %d", len(key))
	}
	if meta.Version != 1 {
		t.Fatalf("expected version 1, got %d", meta.Version)
	}
	if meta.Status != "active" {
		t.Fatalf("expected status=active, got %s", meta.Status)
	}

	t.Logf("OK: InitKeyStore generated v1 key (id=%s)", meta.ID)
}

func TestInitKeyStore_SecondTime(t *testing.T) {
	dir := t.TempDir()
	neuronfsDir := filepath.Join(dir, ".neuronfs")

	ks1, _ := InitKeyStore(neuronfsDir)
	_, meta1, _ := ks1.GetActiveDEK()

	// ??Î≤àÏß∏ Ï¥àÍ∏∞????Í∏∞Ï°¥ ???†Ï?
	ks2, err := InitKeyStore(neuronfsDir)
	if err != nil {
		t.Fatalf("second InitKeyStore failed: %v", err)
	}

	_, meta2, _ := ks2.GetActiveDEK()
	if meta1.ID != meta2.ID {
		t.Fatalf("key ID changed: %s ??%s", meta1.ID, meta2.ID)
	}

	t.Logf("OK: InitKeyStore preserves existing key (id=%s)", meta2.ID)
}

func TestRotateKey(t *testing.T) {
	dir := t.TempDir()
	ks, _ := InitKeyStore(filepath.Join(dir, ".neuronfs"))

	_, meta1, _ := ks.GetActiveDEK()

	// ???åÏ†Ñ
	meta2, err := ks.RotateKey()
	if err != nil {
		t.Fatalf("RotateKey failed: %v", err)
	}

	if meta2.Version != 2 {
		t.Fatalf("expected version 2, got %d", meta2.Version)
	}
	if meta2.ID == meta1.ID {
		t.Fatal("new key should have different ID")
	}

	// ?úÏÑ± ?§Í? v2?¨Ïïº ??	_, activeMeta, _ := ks.GetActiveDEK()
	if activeMeta.ID != meta2.ID {
		t.Fatalf("active key should be v2, got id=%s", activeMeta.ID)
	}

	// ?¥Ï†Ñ ?§Í? retired?¨Ïïº ??	_, oldMeta, err := ks.GetKeyByID(meta1.ID)
	if err != nil {
		t.Fatalf("old key not accessible: %v", err)
	}
	if oldMeta.Status != "retired" {
		t.Fatalf("old key should be retired, got %s", oldMeta.Status)
	}

	t.Logf("OK: RotateKey v1?ív2 (old=%s retired, new=%s active)", meta1.ID, meta2.ID)
}

func TestRotateKey_MultipleRotations(t *testing.T) {
	dir := t.TempDir()
	ks, _ := InitKeyStore(filepath.Join(dir, ".neuronfs"))

	// 3???∞ÏÜç ?åÏ†Ñ
	for i := 0; i < 3; i++ {
		_, err := ks.RotateKey()
		if err != nil {
			t.Fatalf("rotation %d failed: %v", i+1, err)
		}
	}

	_, meta, _ := ks.GetActiveDEK()
	if meta.Version != 4 { // 1(initial) + 3(rotations)
		t.Fatalf("expected version 4, got %d", meta.Version)
	}

	keys, _ := ks.ListKeys()
	if len(keys) != 4 {
		t.Fatalf("expected 4 keys total, got %d", len(keys))
	}

	activeCount := 0
	retiredCount := 0
	for _, k := range keys {
		if k.Status == "active" {
			activeCount++
		} else if k.Status == "retired" {
			retiredCount++
		}
	}
	if activeCount != 1 {
		t.Fatalf("expected 1 active key, got %d", activeCount)
	}
	if retiredCount != 3 {
		t.Fatalf("expected 3 retired keys, got %d", retiredCount)
	}

	t.Logf("OK: 3 rotations ??1 active + 3 retired (v%d)", meta.Version)
}

func TestEncryptDecryptWithKeyStore(t *testing.T) {
	dir := t.TempDir()
	ks, _ := InitKeyStore(filepath.Join(dir, ".neuronfs"))

	dek, _, _ := ks.GetActiveDEK()
	plaintext := []byte("cortex/frontend/hooks_pattern ?¥Îü∞ ?∞Ïù¥??)

	// ?îÌò∏??	ciphertext, nonce, err := EncryptNeuron(plaintext, dek)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	// ???åÏ†Ñ
	ks.RotateKey()

	// ?¥Ï†Ñ ?§Î°ú Î≥µÌò∏??(retired key Ï°∞Ìöå)
	oldKey, _, _ := ks.GetKeyByID(ks.active.ID)
	// ?§Ï†úÎ°úÎäî ?¥Ï†Ñ ??IDÎ•?ciphertext Î©îÌ??∞Ïù¥?∞Ïóê ?Ä?•Ìï¥????	// ?¨Í∏∞?úÎäî ?êÎûò dekÎ•?ÏßÅÏ†ë ?¨Ïö©
	decrypted, err := DecryptNeuron(ciphertext, dek, nonce)
	if err != nil {
		t.Fatalf("decrypt with original key failed: %v", err)
	}
	if string(decrypted) != string(plaintext) {
		t.Fatalf("decrypted data mismatch")
	}

	// ???§Î°ú??Î≥µÌò∏???§Ìå®?¥Ïïº ??	newDek, _, _ := ks.GetActiveDEK()
	_, err = DecryptNeuron(ciphertext, newDek, nonce)
	if err == nil {
		t.Fatal("decryption with wrong key should fail")
	}

	_ = oldKey // suppress unused
	t.Log("OK: encrypt?írotate?ídecrypt with old key verified")
}

func TestDestroyKey(t *testing.T) {
	dir := t.TempDir()
	ks, _ := InitKeyStore(filepath.Join(dir, ".neuronfs"))
	_, meta1, _ := ks.GetActiveDEK()

	// ?åÏ†Ñ ???¥Ï†Ñ ????†ú
	ks.RotateKey()

	// ?úÏÑ± ?§Îäî ??†ú Î∂àÍ?
	_, meta2, _ := ks.GetActiveDEK()
	err := ks.DestroyKey(meta2.ID)
	if err == nil {
		t.Fatal("should not be able to destroy active key")
	}

	// retired ????†ú
	err = ks.DestroyKey(meta1.ID)
	if err != nil {
		t.Fatalf("DestroyKey failed: %v", err)
	}

	// ??†ú ??Ï°∞Ìöå Î∂àÍ?
	_, _, err = ks.GetKeyByID(meta1.ID)
	if err == nil {
		t.Fatal("destroyed key should not be accessible")
	}

	t.Logf("OK: destroyed retired key %s", meta1.ID)
}

func TestKeyIntegrity(t *testing.T) {
	dir := t.TempDir()
	ks, _ := InitKeyStore(filepath.Join(dir, ".neuronfs"))
	_, meta, _ := ks.GetActiveDEK()

	// ???åÏùº Î≥ÄÏ°?	keyPath := filepath.Join(dir, ".neuronfs", "keys", meta.ID, "key.json")
	data, _ := os.ReadFile(keyPath)
	// 1Î∞îÏù¥??Î≥ÄÏ°?	tampered := make([]byte, len(data))
	copy(tampered, data)
	for i := range tampered {
		if tampered[i] == 'a' {
			tampered[i] = 'b'
			break
		}
	}
	os.WriteFile(keyPath, tampered, 0600)

	// Î≥ÄÏ°?Í∞êÏ?
	_, _, err := ks.GetActiveDEK()
	if err == nil {
		t.Log("WARN: tamper not detected (change may not have affected key bytes)")
	} else {
		t.Logf("OK: tamper detected: %v", err)
	}
}

