package main

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// SyncPayload represents the standard JSON message sent over Gossipsub
type SyncPayload struct {
	AgentID  string `json:"agent_id"`
	Region   string `json:"region"`
	RootHash string `json:"root_hash"`
	FileCnt  int    `json:"file_cnt"`
}

func initMerkleSyncWorker(ctx context.Context, brainRoot string) error {
	if GlobalP2P == nil || GlobalP2P.Host == nil {
		return fmt.Errorf("p2p node not initialized")
	}

	ps, err := pubsub.NewGossipSub(ctx, GlobalP2P.Host)
	if err != nil {
		return fmt.Errorf("failed to init gossipsub: %w", err)
	}

	topicName := "neuronfs-merkle-sync"
	topic, err := ps.Join(topicName)
	if err != nil {
		return fmt.Errorf("failed to join topic: %w", err)
	}

	sub, err := topic.Subscribe()
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	hmacKey := loadOrCreateHMACKey(brainRoot)
	cortexPath := filepath.Join(brainRoot, "cortex")

	// Listener Goroutine
	go func() {
		for {
			msg, err := sub.Next(ctx)
			if err != nil {
				return
			}
			if msg.ReceivedFrom == GlobalP2P.Host.ID() {
				continue // skip own messages
			}

			var payload SyncPayload
			if err := json.Unmarshal(msg.Data, &payload); err == nil {
				// We only sync Cortex right now to avoid huge overhead
				if payload.Region == "cortex" {
					localChain, _ := BuildChain(cortexPath, hmacKey)
					if localChain != nil && localChain.RootHash != payload.RootHash {
						msgStr := fmt.Sprintf("[P2P Sync] Merkle Diff | Remote (%s): %s... | Local: %s...", 
							payload.AgentID[len(payload.AgentID)-4:], 
							payload.RootHash[:8], 
							localChain.RootHash[:8])
						
						if GlobalSSEBroker != nil {
							GlobalSSEBroker.Broadcast("warn", msgStr)
						}
					}
				}
			}
		}
	}()

	// Broadcaster Goroutine
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				chain, err := BuildChain(cortexPath, hmacKey)
				if err == nil && chain != nil {
					payload := SyncPayload{
						AgentID:  GlobalP2P.Host.ID().String(),
						Region:   "cortex",
						RootHash: chain.RootHash,
						FileCnt:  len(chain.Nodes),
					}
					data, _ := json.Marshal(payload)
					topic.Publish(ctx, data)
				}
			}
		}
	}()

	return nil
}
