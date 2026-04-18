package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ============================================================================
// Module: Jloot VFS Ops (vfs_ops.go)
// Purpose: Global singleton operations for the filesystem wrapper.
// Architecture: Acts as an interceptor for all os.ReadDir/ReadFile.
// ============================================================================

// GlobalVFS is the singleton instance of the RouterFS used everywhere in NeuronFS.
// It is perfectly safe because reads are inherently concurrent-safe.
var GlobalVFS *RouterFS

// GlobalVFSRoot stores the absolute brain root path for absolute→relative conversion.
var GlobalVFSRoot string

// initVFS initializes the primary router with fallback empty lower if needed.
func initVFS(upperPath string) {
	if GlobalVFS == nil {
		GlobalVFS = &RouterFS{
			Upper: nil,
		}
	}
	GlobalVFSRoot = upperPath
}

// vfsRelativize converts absolute paths to VFS-relative paths.
// os.DirFS only accepts relative paths from its root.
func vfsRelativize(path string) string {
	if GlobalVFSRoot != "" {
		rel, err := filepath.Rel(GlobalVFSRoot, path)
		if err == nil && !strings.HasPrefix(rel, "..") {
			path = rel
		}
	}
	path = filepath.ToSlash(path)
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		path = "."
	}
	return path
}

// vfsReadDir replaces os.ReadDir
func vfsReadDir(dir string) ([]fs.DirEntry, error) {
	if os.Getenv("NEURONFS_TEST_ISOLATION") == "1" {
		return os.ReadDir(dir)
	}
	if GlobalVFS == nil {
		return os.ReadDir(dir) // fallback to OS
	}
	return GlobalVFS.ReadDir(vfsRelativize(dir))
}

// vfsReadFile replaces os.ReadFile
func vfsReadFile(path string) ([]byte, error) {
	if os.Getenv("NEURONFS_TEST_ISOLATION") == "1" {
		return os.ReadFile(path)
	}
	if GlobalVFS == nil {
		return os.ReadFile(path) // fallback to OS
	}
	return GlobalVFS.ReadFile(vfsRelativize(path))
}

// vfsGlob replaces filepath.Glob
// Uses the built-in io/fs.Glob which seamlessly traverses our RouterFS.
func vfsGlob(pattern string) ([]string, error) {
	if os.Getenv("NEURONFS_TEST_ISOLATION") == "1" {
		return filepath.Glob(pattern)
	}
	if GlobalVFS == nil {
		return filepath.Glob(pattern) // fallback to OS
	}

	// Convert to VFS-relative path
	relPattern := vfsRelativize(pattern)

	matches, err := fs.Glob(GlobalVFS, relPattern)
	if err != nil {
		return nil, err
	}

	// Return absolute OS-native paths for compatibility with legacy callers
	for i, m := range matches {
		matches[i] = filepath.Join(GlobalVFSRoot, filepath.FromSlash(m))
	}

	return matches, nil
}

// vfsStat replaces os.Stat
func vfsStat(path string) (fs.FileInfo, error) {
	if os.Getenv("NEURONFS_TEST_ISOLATION") == "1" {
		return os.Stat(path)
	}
	if GlobalVFS == nil {
		return os.Stat(path) // fallback to OS
	}
	return fs.Stat(GlobalVFS, vfsRelativize(path))
}

// vfsWalkDir replaces filepath.Walk
// Note: It uses fs.WalkDir which passes fs.DirEntry instead of os.FileInfo.
func vfsWalkDir(root string, fn fs.WalkDirFunc) error {
	if os.Getenv("NEURONFS_TEST_ISOLATION") == "1" {
		return filepath.WalkDir(root, fn)
	}
	if GlobalVFS == nil {
		return fmt.Errorf("GlobalVFS not initialized")
	}
	relRoot := vfsRelativize(root)

	return fs.WalkDir(GlobalVFS, relRoot, func(path string, d fs.DirEntry, err error) error {
		// Convert VFS-relative back to absolute OS-native for legacy callers
		nativePath := filepath.Join(GlobalVFSRoot, filepath.FromSlash(path))
		return fn(nativePath, d, err)
	})
}
