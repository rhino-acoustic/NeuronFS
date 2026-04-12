package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// PackageCmd creates a marketplace-ready .zip archive of the brain folder
type PackageCmd struct{}

func (c *PackageCmd) Name() string {
	return "--package"
}

func (c *PackageCmd) Execute(brainRoot string, args []string) error {
	outputZip := "neuronfs_brain.zip"

	// Parse custom output if provided: --package output.zip
	for i, arg := range args {
		if (arg == "--package" || arg == "--sell") && i+1 < len(args) {
			if !strings.HasPrefix(args[i+1], "--") {
				outputZip = args[i+1]
			}
		}
	}

	fmt.Printf("[Marketplace] Packaging brain from: %s\n", brainRoot)
	fmt.Printf("[Marketplace] Target archive: %s\n", outputZip)

	zipFile, err := os.Create(outputZip)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %v", err)
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	var fileCount int
	var skippedCount int

	err = filepath.Walk(brainRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(brainRoot, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		// Marketplace rules: exclude non-essential operational tracking folders to protect privacy
		// and keep the brain "pure" for execution only.
		if strings.HasPrefix(relPath, ".git") ||
			strings.HasPrefix(relPath, "_inbox") ||
			strings.HasPrefix(relPath, "_transcripts") ||
			strings.HasPrefix(relPath, "_agents") ||
			strings.HasPrefix(relPath, ".archive") ||
			strings.HasPrefix(relPath, ".neuronfs_backup") ||
			strings.HasPrefix(relPath, "scratch") {
			skippedCount++
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(relPath)
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		if err == nil {
			fileCount++
		}
		return err
	})

	if err != nil {
		return fmt.Errorf("packaging error: %v", err)
	}

	fmt.Printf("\033[32m[SUCCESS] Brain packaged successfully.\033[0m\n")
	fmt.Printf("          Included: %d files\n", fileCount)
	fmt.Printf("          Excluded: %d operational/private files\n", skippedCount)
	fmt.Printf("          Ready for Marketplace distribution.\n")

	return nil
}
