package main

import (
	"io/fs"
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

// initVFS initializes the primary router with fallback empty lower if needed.
func initVFS(upperPath string) {
	if GlobalVFS == nil {
		GlobalVFS = &RouterFS{
			Upper: nil, // If we haven't mounted jloot, just pass through.
			// Lower will be mounted via BootBrainwallet/MountCartridge.
		}
	}
}

// vfsReadDir replaces os.ReadDir
func vfsReadDir(dir string) ([]fs.DirEntry, error) {
	if GlobalVFS == nil {
		panic("GlobalVFS is nil. Call MountCartridge() early.")
	}
	return GlobalVFS.ReadDir(dir)
}

// vfsReadFile replaces os.ReadFile
func vfsReadFile(path string) ([]byte, error) {
	if GlobalVFS == nil {
		panic("GlobalVFS is nil.")
	}
	return GlobalVFS.ReadFile(path)
}

// vfsGlob replaces filepath.Glob
// Uses the built-in io/fs.Glob which seamlessly traverses our RouterFS.
func vfsGlob(pattern string) ([]string, error) {
	if GlobalVFS == nil {
		panic("GlobalVFS is nil.")
	}
	
	// fs.Glob needs forward slashes
	pattern = filepath.ToSlash(pattern)
	pattern = strings.TrimPrefix(pattern, "/")

	matches, err := fs.Glob(GlobalVFS, pattern)
	if err != nil {
		return nil, err
	}

	// For compatibility with old filepath.Glob callers, we should ensure
	// paths returned have the os-specific separator if the caller expects it?
	// Actually, fs.FS standardizes on forward slash pathing.
	// But let's return OS-native paths since the older code expects native paths.
	for i, m := range matches {
		matches[i] = filepath.FromSlash(m)
	}

	return matches, nil
}

// vfsStat replaces os.Stat
func vfsStat(path string) (fs.FileInfo, error) {
	if GlobalVFS == nil {
		panic("GlobalVFS is nil.")
	}
	path = filepath.ToSlash(path)
	path = strings.TrimPrefix(path, "/")
	return fs.Stat(GlobalVFS, path)
}

// vfsWalkDir replaces filepath.Walk
// Note: It uses fs.WalkDir which passes fs.DirEntry instead of os.FileInfo.
func vfsWalkDir(root string, fn fs.WalkDirFunc) error {
	if GlobalVFS == nil {
		panic("GlobalVFS is nil.")
	}
	root = filepath.ToSlash(root)
	root = strings.TrimPrefix(root, "/")
	
	return fs.WalkDir(GlobalVFS, root, func(path string, d fs.DirEntry, err error) error {
		// Convert virtual slashes back to OS native for legacy consumers if needed
		nativePath := filepath.FromSlash(path)
		return fn(nativePath, d, err)
	})
}
