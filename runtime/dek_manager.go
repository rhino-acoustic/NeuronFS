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
// 기능: DEK 생성, 안전한 저장, 키 회전, 이력 관리
// 저장소: .neuronfs/keys/ (파일 시스템 기반, 600 권한)
// ============================================================================

var (
	errKeyStoreNotInit = errors.New("dek: key store not initialized")
	errKeyNotFound     = errors.New("dek: active key not found")
	errKeyCorrupted    = errors.New("dek: key file corrupted or tampered")
)

// KeyMetadata는 키의 메타데이터를 저장한다.
type KeyMetadata struct {
	ID        string `json:"id"`         // SHA256(key)[:16] — 키 식별자
	CreatedAt string `json:"created_at"` // RFC3339 생성 시각
	RotatedAt string `json:"rotated_at"` // RFC3339 회전 시각 (이전 키에 기록)
	Version   int    `json:"version"`    // 키 버전 (1부터 시작)
	Status    string `json:"status"`     // "active" | "retired" | "destroyed"
}

// KeyStore는 DEK 키 저장소를 관리한다.
type KeyStore struct {
	mu       sync.RWMutex
	storeDir string        // .neuronfs/keys/
	active   *KeyMetadata  // 현재 활성 키 메타데이터
}

// keyEnvelope는 키를 디스크에 저장할 때 사용하는 구조체이다.
// KEK(Key Encryption Key) 없이 파일 시스템 권한으로 보호한다.
// 향후 KEK 래핑 추가 가능.
type keyEnvelope struct {
	Meta    KeyMetadata `json:"meta"`
	KeyHex  string      `json:"key_hex"`  // DEK를 hex 인코딩
	KeyHash string      `json:"key_hash"` // SHA256(key) — 무결성 검증용
}

// InitKeyStore는 키 저장소를 초기화한다.
// .neuronfs/keys/ 디렉토리에 활성 키가 없으면 자동 생성한다.
func InitKeyStore(neuronfsDir string) (*KeyStore, error) {
	storeDir := filepath.Join(neuronfsDir, "keys")
	if err := os.MkdirAll(storeDir, 0700); err != nil {
		return nil, fmt.Errorf("dek: cannot create key store: %w", err)
	}

	ks := &KeyStore{storeDir: storeDir}

	// 활성 키 로드 시도
	if err := ks.loadActive(); err != nil {
		// 활성 키 없음 → 최초 키 생성
		if _, err := ks.generateAndSave(1); err != nil {
			return nil, fmt.Errorf("dek: initial key generation failed: %w", err)
		}
		if err := ks.loadActive(); err != nil {
			return nil, fmt.Errorf("dek: failed to load after generation: %w", err)
		}
	}

	return ks, nil
}

// GetActiveDEK는 현재 활성 DEK를 반환한다.
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

// RotateKey는 새 DEK를 생성하고 이전 키를 retired로 전환한다.
// 이전 키는 복호화 호환성을 위해 보존된다.
func (ks *KeyStore) RotateKey() (*KeyMetadata, error) {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	oldMeta := ks.active
	newVersion := 1
	if oldMeta != nil {
		newVersion = oldMeta.Version + 1
	}

	// 새 키 생성
	newMeta, err := ks.generateAndSave(newVersion)
	if err != nil {
		return nil, fmt.Errorf("dek: rotation failed: %w", err)
	}

	// 이전 키를 retired로 마킹
	if oldMeta != nil {
		if err := ks.retireKey(oldMeta.ID); err != nil {
			return nil, fmt.Errorf("dek: failed to retire old key: %w", err)
		}
	}

	ks.active = newMeta
	return newMeta, nil
}

// GetKeyByID는 특정 키 ID로 DEK를 조회한다 (retired 포함).
// 이전 버전 키로 암호화된 데이터 복호화 시 사용.
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

// ListKeys는 모든 키의 메타데이터를 반환한다 (키 값은 포함하지 않음).
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
			continue // 손상된 키는 스킵
		}
		keys = append(keys, *meta)
	}

	return keys, nil
}

// DestroyKey는 retired 키를 완전 삭제한다.
// 주의: 이 키로 암호화된 데이터는 복호화 불가.
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

// ─── 내부 함수 ───

func (ks *KeyStore) generateAndSave(version int) (*KeyMetadata, error) {
	dek, err := GenerateDEK()
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(dek)
	keyID := hex.EncodeToString(hash[:8]) // 16자 식별자

	meta := &KeyMetadata{
		ID:        keyID,
		CreatedAt: time.Now().Format(time.RFC3339),
		Version:   version,
		Status:    "active",
	}

	// 키 디렉토리 생성
	keyDir := filepath.Join(ks.storeDir, keyID)
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return nil, err
	}

	// 키 엔벨로프 저장
	env := keyEnvelope{
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

	// 활성 키 심볼릭 포인터 업데이트
	activePath := filepath.Join(ks.storeDir, "active.id")
	if err := os.WriteFile(activePath, []byte(keyID), 0600); err != nil {
		return nil, err
	}

	fmt.Printf("[DEK] 🔑 v%d generated (id=%s)\n", version, keyID)
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

	// 무결성 검증
	hash := sha256.Sum256(key)
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
