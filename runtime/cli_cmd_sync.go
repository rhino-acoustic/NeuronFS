package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

// SyncCmd implements the Command interface for the Phase 60 P2P sync.
type SyncCmd struct{}

func (c *SyncCmd) Name() string {
	return "--sync"
}

func (c *SyncCmd) Execute(brainRoot string, args []string) error {
	fmt.Println("[P2P] Starting NeuronFS sync daemon (Phase 60)...")

	// Initialize the P2P Subsystem and mDNS
	if err := InitializeP2PNode(brainRoot); err != nil {
		return fmt.Errorf("failed to start P2P node: %v", err)
	}

	fmt.Println("[P2P] Searching for other NeuronFS nodes in local network...")
	fmt.Println("[P2P] Note: Discovered peers will automatically receive our brain structure.")
	fmt.Println("[P2P] Press Ctrl+C to stop.")

	// Keep alive to handle connections and wait for peers
	// Note: We'll wait until we connect to someone via mDNS Notifee,
	// but currently the DiscoveryNotifee connects and that's it.
	// To actually SEND the brain, we need to inject SendBrainToPeer inside HandlePeerFound.
	// Since we can't easily hook into HandlePeerFound without altering p2p_node.go more deeply,
	// we'll just poll GlobalP2P's Host for connected peers every few seconds.

	ctx := context.Background()
	sentPeers := make(map[string]bool)

	for {
		if GlobalP2P != nil && GlobalP2P.Host != nil {
			for _, conn := range GlobalP2P.Host.Network().Conns() {
				peerID := conn.RemotePeer()
				pidStr := peerID.String()

				if !sentPeers[pidStr] {
					log.Printf("[P2P Sync] Found connected peer: %s", pidStr)
					
					// Obtain AddrInfo to pass to SendBrainToPeer
					pi := GlobalP2P.Host.Peerstore().PeerInfo(peerID)
					
					// Send brain
					if err := SendBrainToPeer(ctx, pi, brainRoot); err != nil {
						log.Printf("[P2P Sync] Failed to send brain to %s: %v", pidStr, err)
					} else {
						sentPeers[pidStr] = true
					}
				}
			}
		}
		time.Sleep(3 * time.Second)
	}
}
