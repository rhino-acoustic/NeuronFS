package main

import (
	"archive/zip"
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

	if jlootPath == "" {
		fmt.Println("[VFS] Booting without lower cartridge. Upper-only mode.")
		return nil
	}

	// In the real ignition sequence, we would ask for the Brainwallet passphrase,
	// run Argon2 KDF (via dek_manager.go), decrypt the XChaCha20 stream (crypto_neuron.go),
	// and pass the decrypted bytes to a virtual zip.Reader.
	// For compilation and Phase 2 proof-of-concept, we attempt a standard ZIP read.
	
	zrc, err := zip.OpenReader(jlootPath)
	if err != nil {
		fmt.Printf("[VFS] Warning: Failed to mount cartridge '%s': %v\n", jlootPath, err)
		fmt.Println("[VFS] Falling back to Upper-only mode.")
		return nil // Non-fatal, just means no lower knowledge base
	}

	GlobalVFS.Lower = zrc
	fmt.Printf("[VFS] Mounted Cartridge: %s as LowerDir (O(1) Route active)\n", jlootPath)

	return nil
}
