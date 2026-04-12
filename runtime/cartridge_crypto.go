package main

import (
	"archive/tar"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ============================================================================
// Module: Cartridge Export/Import — AES-256-GCM Encrypted Brain Packaging (V12-D)
// Exports the entire brain_v4/ folder into a single .cartridge file
// that is encrypted, portable, and verifiable.
// ============================================================================

// ExportCartridgeCmd implements the --export-cartridge CLI command
type ExportCartridgeCmd struct{}

func (c *ExportCartridgeCmd) Name() string        { return "--export-cartridge" }
func (c *ExportCartridgeCmd) Description() string  { return "Export brain as encrypted .cartridge file" }

func (c *ExportCartridgeCmd) Execute(brainRoot string, args []string) error {
	// Parse passphrase from args
	passphrase := ""
	outputPath := ""
	for i, arg := range args {
		if arg == "--pass" && i+1 < len(args) {
			passphrase = args[i+1]
		}
		if arg == "--out" && i+1 < len(args) {
			outputPath = args[i+1]
		}
	}

	marketplace := false
	for _, arg := range args {
		if arg == "--marketplace" {
			marketplace = true
		}
	}

	if passphrase == "" {
		return fmt.Errorf("--pass <passphrase> is required for encryption")
	}

	if outputPath == "" {
		outputPath = filepath.Join(filepath.Dir(brainRoot), fmt.Sprintf("brain_%d.cartridge", time.Now().Unix()))
	}

	fmt.Printf("[Cartridge] Packaging %s...\n", brainRoot)

	// Step 1: Create tar.gz of brain directory (with PII masking if marketplace)
	tarData, err := createTarGz(brainRoot)
	if err != nil {
		return fmt.Errorf("tar.gz creation failed: %w", err)
	}

	// Step 1.5: PII Masking for marketplace distribution
	if marketplace {
		tarData, err = maskTarGzPII(tarData)
		if err != nil {
			fmt.Printf("[Cartridge] PII masking warning: %v (continuing)\n", err)
		}
	}

	fmt.Printf("[Cartridge] Compressed: %d bytes\n", len(tarData))

	// Step 2: Encrypt with AES-256-GCM
	encrypted, err := encryptAES256GCM(tarData, passphrase)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	// Step 3: Write cartridge file with magic header
	header := []byte("NFSC") // NeuronFS Cartridge magic bytes
	version := []byte{0x01}  // Version 1

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("file creation failed: %w", err)
	}
	defer f.Close()

	f.Write(header)
	f.Write(version)
	f.Write(encrypted)

	fmt.Printf("[Cartridge] Exported: %s (%d bytes)\n", outputPath, len(header)+len(version)+len(encrypted))
	RecordAudit(brainRoot, "cartridge", "export", outputPath, fmt.Sprintf("%d bytes encrypted", len(encrypted)), true)
	return nil
}

// ImportCartridgeCmd implements the --import-cartridge CLI command
type ImportCartridgeCmd struct{}

func (c *ImportCartridgeCmd) Name() string        { return "--import-cartridge" }
func (c *ImportCartridgeCmd) Description() string  { return "Import and decrypt a .cartridge file into brain" }

func (c *ImportCartridgeCmd) Execute(brainRoot string, args []string) error {
	passphrase := ""
	cartridgePath := ""
	for i, arg := range args {
		if arg == "--pass" && i+1 < len(args) {
			passphrase = args[i+1]
		}
		if arg == "--file" && i+1 < len(args) {
			cartridgePath = args[i+1]
		}
	}

	if passphrase == "" || cartridgePath == "" {
		return fmt.Errorf("--pass <passphrase> and --file <path.cartridge> are required")
	}

	// Step 1: Read cartridge file
	data, err := os.ReadFile(cartridgePath)
	if err != nil {
		return fmt.Errorf("failed to read cartridge: %w", err)
	}

	// Verify magic header
	if len(data) < 5 || string(data[:4]) != "NFSC" {
		return fmt.Errorf("invalid cartridge file (bad magic header)")
	}
	version := data[4]
	if version != 0x01 {
		return fmt.Errorf("unsupported cartridge version: %d", version)
	}
	encrypted := data[5:]

	fmt.Printf("[Cartridge] Decrypting %s...\n", cartridgePath)

	// Step 2: Decrypt
	tarData, err := decryptAES256GCM(encrypted, passphrase)
	if err != nil {
		return fmt.Errorf("decryption failed (wrong passphrase?): %w", err)
	}

	// Step 3: Extract tar.gz to brain root
	if err := extractTarGz(tarData, brainRoot); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	fmt.Printf("[Cartridge] Imported successfully to %s\n", brainRoot)
	RecordAudit(brainRoot, "cartridge", "import", cartridgePath, "decrypted and extracted", true)
	return nil
}

// ── Crypto Helpers ──

func deriveKey(passphrase string) []byte {
	hash := sha256.Sum256([]byte(passphrase))
	return hash[:]
}

func encryptAES256GCM(plaintext []byte, passphrase string) ([]byte, error) {
	key := deriveKey(passphrase)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func decryptAES256GCM(ciphertext []byte, passphrase string) ([]byte, error) {
	key := deriveKey(passphrase)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// ── Tar/Gz Helpers ──

func createTarGz(sourceDir string) ([]byte, error) {
	var buf strings.Builder
	_ = buf // placeholder — we need bytes.Buffer
	
	// Use a proper bytes buffer
	byteBuf := &bytesBuffer{}
	gzWriter := gzip.NewWriter(byteBuf)
	tarWriter := tar.NewWriter(gzWriter)

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}
		// Skip _archive and _quarantine
		if strings.Contains(path, "_archive") || strings.Contains(path, "_quarantine") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		relPath, _ := filepath.Rel(sourceDir, path)
		header, headerErr := tar.FileInfoHeader(info, "")
		if headerErr != nil {
			return nil
		}
		header.Name = filepath.ToSlash(relPath)
		if writeErr := tarWriter.WriteHeader(header); writeErr != nil {
			return writeErr
		}
		if !info.IsDir() {
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return nil
			}
			if _, writeErr := tarWriter.Write(data); writeErr != nil {
				return writeErr
			}
		}
		return nil
	})

	tarWriter.Close()
	gzWriter.Close()
	return byteBuf.Bytes(), err
}

// bytesBuffer is a simple byte buffer to avoid import collision
type bytesBuffer struct {
	data []byte
}

func (b *bytesBuffer) Write(p []byte) (int, error) {
	b.data = append(b.data, p...)
	return len(p), nil
}

func (b *bytesBuffer) Bytes() []byte {
	return b.data
}

func extractTarGz(data []byte, destDir string) error {
	gzReader, err := gzip.NewReader(strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		targetPath := filepath.Join(destDir, filepath.FromSlash(header.Name))

		switch header.Typeflag {
		case tar.TypeDir:
			_ = os.MkdirAll(targetPath, 0755)
		case tar.TypeReg:
			_ = os.MkdirAll(filepath.Dir(targetPath), 0755)
			outFile, createErr := os.Create(targetPath)
			if createErr != nil {
				continue
			}
			io.Copy(outFile, tarReader)
			outFile.Close()
		}
	}
	return nil
}
