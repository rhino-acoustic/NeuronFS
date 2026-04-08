package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ============================================================================
// Module: Jloot OverlayFS Core (vfs_core.go)
// Purpose: Core data structures for the VFS routing engine.
// Architecture: UnionFS pattern to merge memory (Lower) and disk (Upper).
// ============================================================================

// RouterFS implements a union filesystem over a lower (immutable RAM) 
// and an upper (mutable disk) layer.
type RouterFS struct {
	Lower fs.FS // The .jloot Cartridge (e.g., zip.Reader)
	Upper fs.FS // The local disk tree (e.g., os.DirFS)
}

// Open implements the standard fs.FS interface.
// It prioritizes the Upper layer, falling back to the Lower layer.
func (rfs *RouterFS) Open(name string) (fs.File, error) {
	name = filepath.ToSlash(name)
	name = strings.TrimPrefix(name, "/")
	if name == "" {
		name = "."
	}

	// 1. Try Upper (Disk)
	if rfs.Upper != nil {
		f, err := rfs.Upper.Open(name)
		if err == nil {
			return f, nil
		}
	}

	// 2. Try Lower (Cartridge)
	if rfs.Lower != nil {
		f, err := rfs.Lower.Open(name)
		if err == nil {
			return f, nil
		}
	}

	return nil, os.ErrNotExist
}

// ReadDir reads a directory by querying both layers and merging the results.
// Shadows Lower layer with Upper layer on name collision.
func (rfs *RouterFS) ReadDir(name string) ([]fs.DirEntry, error) {
	name = filepath.ToSlash(name)
	name = strings.TrimPrefix(name, "/")
	if name == "" {
		name = "."
	}

	entriesMap := make(map[string]fs.DirEntry)

	// 1. Read from Lower (RAM/Cartridge)
	if rfs.Lower != nil {
		lowerEntries, err := fs.ReadDir(rfs.Lower, name)
		if err == nil {
			for _, e := range lowerEntries {
				entriesMap[e.Name()] = e
			}
		}
	}

	// 2. Read from Upper (Disk/Mutation) - Overrides Lower
	if rfs.Upper != nil {
		upperEntries, err := fs.ReadDir(rfs.Upper, name)
		if err == nil {
			for _, e := range upperEntries {
				entriesMap[e.Name()] = e // Shadowing happens here
			}
		}
	}

	if len(entriesMap) == 0 {
		return nil, os.ErrNotExist
	}

	// Dump map to slice and sort for determinism
	var merged []fs.DirEntry
	for _, e := range entriesMap {
		merged = append(merged, e)
	}

	sort.Slice(merged, func(i, j int) bool {
		return merged[i].Name() < merged[j].Name()
	})

	return merged, nil
}

// ReadFile merges the file reading logic.
// In an OverlayFS, we first check the upper layer, then the lower layer.
func (rfs *RouterFS) ReadFile(name string) ([]byte, error) {
	name = filepath.ToSlash(name)
	name = strings.TrimPrefix(name, "/")
	if name == "" {
		return nil, fmt.Errorf("invalid path: %s", name)
	}

	// 1. Try Upper
	if rfs.Upper != nil {
		b, err := fs.ReadFile(rfs.Upper, name)
		if err == nil {
			return b, nil
		}
	}

	// 2. Try Lower
	if rfs.Lower != nil {
		b, err := fs.ReadFile(rfs.Lower, name)
		if err == nil {
			return b, nil
		}
	}

	return nil, os.ErrNotExist
}
