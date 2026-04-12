package main

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

// handleSyncStream receives a zip archive from a peer and saves it to _inbox/p2p/
func handleSyncStream(s network.Stream) {
	defer s.Close()
	peerID := s.Conn().RemotePeer().String()
	log.Printf("[P2P Sync] Incoming Sync Stream from: %s\n", peerID)

	// Build safe reception path in inbox
	home := os.Getenv("USERPROFILE")
	if home == "" {
		home = "."
	}
	brainRoot := filepath.Join(home, "NeuronFS", "brain_v4")
	p2pInbox := filepath.Join(brainRoot, "_inbox", "p2p")
	
	if err := os.MkdirAll(p2pInbox, 0755); err != nil {
		log.Printf("[P2P Sync] Failed to create inbox: %v", err)
		return
	}

	timestamp := time.Now().Format("20060102150405")
	fileName := fmt.Sprintf("peer_%s_%s.zip", peerID[:8], timestamp)
	outPath := filepath.Join(p2pInbox, fileName)

	outFile, err := os.Create(outPath)
	if err != nil {
		log.Printf("[P2P Sync] Failed to create file: %v", err)
		return
	}
	defer outFile.Close()

	// Read stream directly into zip file
	bytesCopied, err := io.Copy(outFile, s)
	if err != nil {
		log.Printf("[P2P Sync] Stream read error: %v", err)
		return
	}

	fmt.Printf("\033[32m[SUCCESS] Received brain from peer: %s\033[0m\n", peerID)
	fmt.Printf("          Saved isolated payload to: %s (%d bytes)\n", outPath, bytesCopied)
	
	// Notify via SSE
	if GlobalSSEBroker != nil {
		GlobalSSEBroker.Broadcast("info", fmt.Sprintf("[P2P] Received brain packet from %s", peerID[:8]))
	}
}

// SendBrainToPeer streams the local brain structure (excluding caches) directly over libp2p
func SendBrainToPeer(ctx context.Context, pi peer.AddrInfo, brainRoot string) error {
	if GlobalP2P == nil || GlobalP2P.Host == nil {
		return fmt.Errorf("p2p node not initialized")
	}

	s, err := GlobalP2P.Host.NewStream(ctx, pi.ID, "/neuronfs/sync/1.0.0")
	if err != nil {
		return fmt.Errorf("failed to open sync stream: %v", err)
	}
	defer s.Close()

	fmt.Printf("[P2P Sync] Compressing and streaming brain to peer %s...\n", pi.ID.String()[:8])
	
	archive := zip.NewWriter(s)

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

		// Marketplace rules for excluding pure execution layer overlaps
		if strings.HasPrefix(relPath, ".git") ||
			strings.HasPrefix(relPath, "_inbox") ||
			strings.HasPrefix(relPath, "_transcripts") ||
			strings.HasPrefix(relPath, "_agents") ||
			strings.HasPrefix(relPath, ".archive") ||
			strings.HasPrefix(relPath, ".neuronfs_backup") ||
			strings.HasPrefix(relPath, "scratch") {
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
		return err
	})

	if err != nil {
		return fmt.Errorf("packaging error during stream: %v", err)
	}

	if err := archive.Close(); err != nil {
		return fmt.Errorf("failed closing zip writer: %v", err)
	}

	fmt.Printf("\033[32m[SUCCESS] Transmission complete to peer: %s\033[0m\n", pi.ID.String()[:8])
	return nil
}
