package main

import (
	"time"
)

// SyncToNAS periodically mirrors the brain structure to a network attached storage
// using robocopy (Windows native fast copy).
func SyncToNAS(brainRoot string, nasRoot string, stopCh <-chan struct{}) {
	for {
		select {
		case <-stopCh:
			return
		default:
		}
		
		// robocopy returns non-zero exit codes for successful copies (e.g. 1 means files copied),
		// so we ignore the error return.
		_ = SafeExec(ExecTimeoutSync, "robocopy", brainRoot, nasRoot, "/MIR", "/FFT", "/XO", "/MT:4", "/NDL", "/NJH", "/NJS", "/NC", "/NS", "/NP")
		
		select {
		case <-stopCh:
			return
		case <-time.After(5 * time.Second):
		}
	}
}
