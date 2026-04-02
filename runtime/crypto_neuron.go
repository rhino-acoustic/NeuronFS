package main

import (
	"crypto/rand"
	"errors"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
)

// ============================================================================
// Module: Neuron Encryption (Zero Trust Layer 1)
// Cipher: XChaCha20-Poly1305 (AEAD)
// DEK:    256-bit (32 bytes) ??Data Encryption Key
// Nonce:  192-bit (24 bytes) ??Extended nonce, safe for random generation
// ============================================================================

var (
	errDEKLength   = errors.New("neuron/crypto: DEK must be 32 bytes (256-bit)")
	errNonceLength = fmt.Errorf("neuron/crypto: nonce must be %d bytes", chacha20poly1305.NonceSizeX)
	errDecrypt     = errors.New("neuron/crypto: decryption failed (tampered or wrong key)")
	errEmptyData   = errors.New("neuron/crypto: plaintext is empty")
)

// EncryptNeuron encrypts plaintext neuron data using XChaCha20-Poly1305.
// Returns ciphertext (with appended auth tag) and a randomly generated nonce.
// The caller is responsible for storing the nonce alongside the ciphertext.
func EncryptNeuron(plaintext []byte, dek []byte) (ciphertext []byte, nonce []byte, err error) {
	if len(plaintext) == 0 {
		return nil, nil, errEmptyData
	}
	if len(dek) != chacha20poly1305.KeySize {
		return nil, nil, errDEKLength
	}

	aead, err := chacha20poly1305.NewX(dek)
	if err != nil {
		return nil, nil, fmt.Errorf("neuron/crypto: failed to create AEAD: %w", err)
	}

	// Generate random 24-byte nonce (XChaCha20 extended nonce)
	nonce = make([]byte, chacha20poly1305.NonceSizeX)
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, fmt.Errorf("neuron/crypto: nonce generation failed: %w", err)
	}

	// Seal: ciphertext = encrypted_data || 16-byte Poly1305 tag
	ciphertext = aead.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

// DecryptNeuron decrypts ciphertext using XChaCha20-Poly1305.
// Verifies the Poly1305 authentication tag before returning plaintext.
func DecryptNeuron(ciphertext []byte, dek []byte, nonce []byte) (plaintext []byte, err error) {
	if len(dek) != chacha20poly1305.KeySize {
		return nil, errDEKLength
	}
	if len(nonce) != chacha20poly1305.NonceSizeX {
		return nil, errNonceLength
	}

	aead, err := chacha20poly1305.NewX(dek)
	if err != nil {
		return nil, fmt.Errorf("neuron/crypto: failed to create AEAD: %w", err)
	}

	plaintext, err = aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errDecrypt
	}
	return plaintext, nil
}

// GenerateDEK creates a cryptographically secure 256-bit data encryption key.
func GenerateDEK() ([]byte, error) {
	dek := make([]byte, chacha20poly1305.KeySize)
	if _, err := rand.Read(dek); err != nil {
		return nil, fmt.Errorf("neuron/crypto: DEK generation failed: %w", err)
	}
	return dek, nil
}

