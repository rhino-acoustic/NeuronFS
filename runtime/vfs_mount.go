package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
)

// ============================================================================
// Module: Jloot VFS Mounter (vfs_mount.go)
// Purpose: Ignition sequence and mounting logic for the Jloot Cartridge.
// Architecture: Brainwallet (Master Key) -> XChaCha20 Decrypt -> ZIP Mount -> VFS
// ============================================================================

// MountCartridge sets up the global VFS namespace.
// jlootPath: The path to the physical .jloot cartridge (Lower Layer)
// rootDir: The path to the physical directory (Upper Layer, e.g. brain_v4)
func MountCartridge(jlootPath string, rootDir string) error {
	initVFS(rootDir)
	GlobalVFS.Upper = os.DirFS(rootDir)

	// Ignition Loop Intercept
	masterKey := StartIgnition(jlootPath)

	if masterKey == nil {
		fmt.Println("[VFS] Booting without encrypted lower cartridge. Upper-only mode.")
		return nil
	}

	fmt.Println("[VFS] Neural decryption sequence initiated...")

	// 100% RAM-based decryption (Zero Disks Drops)
	decryptedBytes, err := DecryptCartridgeToRAM(jlootPath, masterKey)
	if err != nil {
		fmt.Printf("[VFS] ❌ FATAL: Brainwallet verification failed or cartridge is corrupted.\nError: %v\n", err)
		fmt.Println("[VFS] Falling back to Upper-only mode for safety.")
		return nil
	}

	// Mount decrypted payload as a Virtual ZIP File System
	zrc, err := zip.NewReader(bytes.NewReader(decryptedBytes), int64(len(decryptedBytes)))
	if err != nil {
		fmt.Printf("[VFS] ❌ FATAL: Decrypted payload is not a valid Jloot Archive.\nError: %v\n", err)
		return nil
	}

	GlobalVFS.Lower = zrc
	fmt.Printf("[VFS] 🌌 Cartridge successfully unfolded in memory: %s (O(1) Route active)\n", jlootPath)

	// Wipe master key from RAM after successful mount
	for i := range masterKey {
		masterKey[i] = 0
	}

	return nil
}
