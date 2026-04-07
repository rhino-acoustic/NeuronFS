package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"

	"golang.org/x/crypto/chacha20poly1305"
)

var (
	errCartridgeFormat = errors.New("neuron/crypto: invalid cartridge format (too small)")
)

// EncryptCartridge encrypts raw ZIP bytes and saves it as .jloot.
// Structure: [24-byte Nonce][Ciphertext]
func EncryptCartridge(plaintext []byte, dek []byte, outPath string) error {
	if len(dek) != chacha20poly1305.KeySize {
		return errDEKLength
	}

	aead, err := chacha20poly1305.NewX(dek)
	if err != nil {
		return fmt.Errorf("neuron/crypto: failed to create AEAD: %w", err)
	}

	nonce := make([]byte, chacha20poly1305.NonceSizeX)
	if _, err := rand.Read(nonce); err != nil {
		return fmt.Errorf("neuron/crypto: nonce generation failed: %w", err)
	}

	ciphertext := aead.Seal(nil, nonce, plaintext, nil)

	// Combine Nonce + Ciphertext
	finalPayload := make([]byte, 0, len(nonce)+len(ciphertext))
	finalPayload = append(finalPayload, nonce...)
	finalPayload = append(finalPayload, ciphertext...)

	return os.WriteFile(outPath, finalPayload, 0600)
}

// DecryptCartridgeToRAM loads a .jloot enc-zip, decrypts it using the 32-byte key, 
// and returns the raw ZIP bytes entirely in memory.
func DecryptCartridgeToRAM(jlootPath string, key []byte) ([]byte, error) {
	if len(key) != chacha20poly1305.KeySize {
		return nil, errDEKLength
	}

	data, err := os.ReadFile(jlootPath)
	if err != nil {
		return nil, fmt.Errorf("neuron/crypto: failed to read cartridge: %w", err)
	}

	if len(data) < chacha20poly1305.NonceSizeX {
		return nil, errCartridgeFormat
	}

	nonce := data[:chacha20poly1305.NonceSizeX]
	ciphertext := data[chacha20poly1305.NonceSizeX:]

	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("neuron/crypto: failed to create AEAD: %w", err)
	}

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errDecrypt
	}

	return plaintext, nil
}
