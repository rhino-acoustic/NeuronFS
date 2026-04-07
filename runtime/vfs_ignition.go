package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/argon2"
	"golang.org/x/term"
)

// ============================================================================
// Module: Jloot Ignition Router
// Purpose: Interactive prompt & KDF layer for Brainwallet
// ============================================================================

// DeriveMasterKey converts a mnemonic string into a 32-byte Key using Argon2id.
// It uses a fixed deterministic salt because Brainwallet IS the key.
func DeriveMasterKey(mnemonic string) []byte {
	salt := []byte("Jloot.NeuronFS.Brainwallet.Salt.V1")
	
	// Argon2id parameters (moderately secure for CLI fast-boot)
	time := uint32(1)
	memory := uint32(64 * 1024)
	threads := uint8(4)
	keyLen := uint32(32)

	key := argon2.IDKey([]byte(mnemonic), salt, time, memory, threads, keyLen)
	return key
}

// StartIgnition handles the boot prompt if a lower cartridge exists.
// Returns the 32-byte cryptographic key. Returns nil if bypassed/empty.
func StartIgnition(jlootPath string) []byte {
	if jlootPath == "" {
		return nil
	}

	if _, err := os.Stat(jlootPath); os.IsNotExist(err) {
		return nil // No cartridge, no prompt needed.
	}

	fmt.Printf("\n[🔒 Jloot OS] Sealed memory cartridge detected: %s\n", jlootPath)
	
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		fmt.Println("[!] Warning: Non-interactive terminal cannot supply Brainwallet passphrase securely.")
		fmt.Println("[!] Cartridge will not be loaded. Booting purely physical OS.")
		return nil
	}

	fmt.Print("Enter Brainwallet (Mnemonic) to ignite: ")

	// x/term handles hiding the typed password
	passwordBytes, err := term.ReadPassword(fd)
	if err != nil {
		fmt.Println("\n[VFS] Ignition aborted.")
		os.Exit(1)
	}

	fmt.Println("\n[⏳] Deriving synaptic master key (Argon2)...")
	
	mnemonic := string(passwordBytes)
	key := DeriveMasterKey(mnemonic)

	// Clean memory aggressively
	for i := range passwordBytes {
		passwordBytes[i] = 0
	}
	mnemonic = "" // Best effort

	return key
}
