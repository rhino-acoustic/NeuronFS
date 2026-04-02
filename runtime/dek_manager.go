package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ============================================================================
// Module: DEK Manager (Zero Trust Key Lifecycle)
// ê¸°ëŠ¥: DEK ?ì„±, ?ˆì „???€?? ???Œì „, ?´ë ¥ ê´€ë¦?// ?€?¥ì†Œ: .neuronfs/keys/ (?Œì¼ ?œìŠ¤??ê¸°ë°˜, 600 ê¶Œí•œ)
// ============================================================================

var (
	errKeyStoreNotInit = errors.New("dek: key store not initialized")
	errKeyNotFound     = errors.New("dek: active key not found")
	errKeyCorrupted    = errors.New("dek: key file corrupted or tampered")
)

// KeyMetadata???¤ì˜ ë©”í??°ì´?°ë? ?€?¥í•œ??
type KeyMetadata struct {
	ID        string `json:"id"`         // SHA256(key)[:16] ?????ë³„??	CreatedAt string `json:"created_at"` // RFC3339 ?ì„± ?œê°
	RotatedAt string `json:"rotated_at"` // RFC3339 ?Œì „ ?œê° (?´ì „ ?¤ì— ê¸°ë¡)
	Version   int    `json:"version"`    // ??ë²„ì „ (1ë¶€???œì‘)
	Status    string `json:"status"`     // "active" | "retired" | "destroyed"
}

// KeyStore??DEK ???€?¥ì†Œë¥?ê´€ë¦¬í•œ??
type KeyStore struct {
	mu       sync.RWMutex
	storeDir string        // .neuronfs/keys/
	active   *KeyMetadata  // ?„ì¬ ?œì„± ??ë©”í??°ì´??}

// keyEnvelope???¤ë? ?”ìŠ¤?¬ì— ?€?¥í•  ???¬ìš©?˜ëŠ” êµ¬ì¡°ì²´ì´??
// KEK(Key Encryption Key) ?†ì´ ?Œì¼ ?œìŠ¤??ê¶Œí•œ?¼ë¡œ ë³´í˜¸?œë‹¤.
// ?¥í›„ KEK ?˜í•‘ ì¶”ê? ê°€??
type keyEnvelope struct {
	Meta    KeyMetadata `json:"meta"`
	KeyHex  string      `json:"key_hex"`  // DEKë¥?hex ?¸ì½”??	KeyHash string      `json:"key_hash"` // SHA256(key) ??ë¬´ê²°??ê²€ì¦ìš©
}

// InitKeyStore?????€?¥ì†Œë¥?ì´ˆê¸°?”í•œ??
// .neuronfs/keys/ ?”ë ‰? ë¦¬???œì„± ?¤ê? ?†ìœ¼ë©??ë™ ?ì„±?œë‹¤.
func InitKeyStore(neuronfsDir string) (*KeyStore, error) {
	storeDir := filepath.Join(neuronfsDir, "keys")
	if err := os.MkdirAll(storeDir, 0700); err != nil {
		return nil, fmt.Errorf("dek: cannot create key store: %w", err)
	}

	ks := &KeyStore{storeDir: storeDir}

	// ?œì„± ??ë¡œë“œ ?œë„
	if err := ks.loadActive(); err != nil {
		// ?œì„± ???†ìŒ ??ìµœì´ˆ ???ì„±
		if _, err := ks.generateAndSave(1); err != nil {
			return nil, fmt.Errorf("dek: initial key generation failed: %w", err)
		}
		if err := ks.loadActive(); err != nil {
			return nil, fmt.Errorf("dek: failed to load after generation: %w", err)
		}
	}

	return ks, nil
}

// GetActiveDEK???„ì¬ ?œì„± DEKë¥?ë°˜í™˜?œë‹¤.
func (ks *KeyStore) GetActiveDEK() ([]byte, *KeyMetadata, error) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	if ks.active == nil {
		return nil, nil, errKeyNotFound
	}

	key, err := ks.readKey(ks.active.ID)
	if err != nil {
		return nil, nil, err
	}

	return key, ks.active, nil
}

// RotateKey????DEKë¥??ì„±?˜ê³  ?´ì „ ?¤ë? retiredë¡??„í™˜?œë‹¤.
// ?´ì „ ?¤ëŠ” ë³µí˜¸???¸í™˜?±ì„ ?„í•´ ë³´ì¡´?œë‹¤.
func (ks *KeyStore) RotateKey() (*KeyMetadata, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	oldMeta := ks.active
	newVersion := 1
	if oldMeta != nil {
		newVersion = oldMeta.Version + 1
	}

	// ?????ì„±
	newMeta, err := ks.generateAndSave(newVersion)
	if err != nil {
		return nil, fmt.Errorf("dek: rotation failed: %w", err)
	}

	// ?´ì „ ?¤ë? retiredë¡?ë§ˆí‚¹
	if oldMeta != nil {
		if err := ks.retireKey(oldMeta.ID); err != nil {
			return nil, fmt.Errorf("dek: failed to retire old key: %w", err)
		}
	}

	ks.active = newMeta
	return newMeta, nil
}

// GetKeyByID???¹ì • ??IDë¡?DEKë¥?ì¡°íšŒ?œë‹¤ (retired ?¬í•¨).
// ?´ì „ ë²„ì „ ?¤ë¡œ ?”í˜¸?”ëœ ?°ì´??ë³µí˜¸?????¬ìš©.
func (ks *KeyStore) GetKeyByID(keyID string) ([]byte, *KeyMetadata, error) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	key, err := ks.readKey(keyID)
	if err != nil {
		return nil, nil, err
	}

	meta, err := ks.readMeta(keyID)
	if err != nil {
		return nil, nil, err
	}

	return key, meta, nil
}

// ListKeys??ëª¨ë“  ?¤ì˜ ë©”í??°ì´?°ë? ë°˜í™˜?œë‹¤ (??ê°’ì? ?¬í•¨?˜ì? ?ŠìŒ).
func (ks *KeyStore) ListKeys() ([]KeyMetadata, error) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	entries, err := os.ReadDir(ks.storeDir)
	if err != nil {
		return nil, fmt.Errorf("dek: cannot list keys: %w", err)
	}

	var keys []KeyMetadata
	for _, e := range entries {
		if !e.IsDir() || e.Name() == "." || e.Name() == ".." {
			continue
		}
		meta, err := ks.readMeta(e.Name())
		if err != nil {
			continue // ?ìƒ???¤ëŠ” ?¤í‚µ
		}
		keys = append(keys, *meta)
	}

	return keys, nil
}

// DestroyKey??retired ?¤ë? ?„ì „ ?? œ?œë‹¤.
// ì£¼ì˜: ???¤ë¡œ ?”í˜¸?”ëœ ?°ì´?°ëŠ” ë³µí˜¸??ë¶ˆê?.
func (ks *KeyStore) DestroyKey(keyID string) error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	if ks.active != nil && ks.active.ID == keyID {
		return errors.New("dek: cannot destroy active key")
	}

	meta, err := ks.readMeta(keyID)
	if err != nil {
		return err
	}
	if meta.Status != "retired" {
		return errors.New("dek: can only destroy retired keys")
	}

	keyDir := filepath.Join(ks.storeDir, keyID)
	return os.RemoveAll(keyDir)
}

// ?€?€?€ ?´ë? ?¨ìˆ˜ ?€?€?€

func (ks *KeyStore) generateAndSave(version int) (*KeyMetadata, error) {
	dek, err := GenerateDEK()
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(dek)
	keyID := hex.EncodeToString(hash[:8]) // 16???ë³„??
	meta := &KeyMetadata{
		ID:        keyID,
		CreatedAt: time.Now().Format(time.RFC3339),
		Version:   version,
		Status:    "active",
	}

	// ???”ë ‰? ë¦¬ ?ì„±
	keyDir := filepath.Join(ks.storeDir, keyID)
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return nil, err
	}

	// ???”ë²¨ë¡œí”„ ?€??	env := keyEnvelope{
		Meta:    *meta,
		KeyHex:  hex.EncodeToString(dek),
		KeyHash: hex.EncodeToString(hash[:]),
	}

	envData, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return nil, err
	}

	envPath := filepath.Join(keyDir, "key.json")
	if err := os.WriteFile(envPath, envData, 0600); err != nil {
		return nil, err
	}

	// ?œì„± ???¬ë³¼ë¦??¬ì¸???…ë°?´íŠ¸
	activePath := filepath.Join(ks.storeDir, "active.id")
	if err := os.WriteFile(activePath, []byte(keyID), 0600); err != nil {
		return nil, err
	}

	fmt.Printf("[DEK] ?”‘ v%d generated (id=%s)\n", version, keyID)
	return meta, nil
}

func (ks *KeyStore) loadActive() error {
	activePath := filepath.Join(ks.storeDir, "active.id")
	data, err := os.ReadFile(activePath)
	if err != nil {
		return errKeyNotFound
	}

	keyID := string(data)
	meta, err := ks.readMeta(keyID)
	if err != nil {
		return err
	}

	ks.active = meta
	return nil
}

func (ks *KeyStore) readKey(keyID string) ([]byte, error) {
	envPath := filepath.Join(ks.storeDir, keyID, "key.json")
	data, err := os.ReadFile(envPath)
	if err != nil {
		return nil, errKeyNotFound
	}

	var env keyEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, errKeyCorrupted
	}

	key, err := hex.DecodeString(env.KeyHex)
	if err != nil {
		return nil, errKeyCorrupted
	}

	// ë¬´ê²°??ê²€ì¦?	hash := sha256.Sum256(key)
	if hex.EncodeToString(hash[:]) != env.KeyHash {
		return nil, errKeyCorrupted
	}

	return key, nil
}

func (ks *KeyStore) readMeta(keyID string) (*KeyMetadata, error) {
	envPath := filepath.Join(ks.storeDir, keyID, "key.json")
	data, err := os.ReadFile(envPath)
	if err != nil {
		return nil, errKeyNotFound
	}

	var env keyEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, errKeyCorrupted
	}

	return &env.Meta, nil
}

func (ks *KeyStore) retireKey(keyID string) error {
	envPath := filepath.Join(ks.storeDir, keyID, "key.json")
	data, err := os.ReadFile(envPath)
	if err != nil {
		return err
	}

	var env keyEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return err
	}

	env.Meta.Status = "retired"
	env.Meta.RotatedAt = time.Now().Format(time.RFC3339)

	updated, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(envPath, updated, 0600)
}

