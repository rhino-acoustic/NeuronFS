package main

import (
	"context"
	"fmt"
	"log"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

// P2PNet represents the local NeuronFS agent node
type P2PNet struct {
	Host  host.Host
	Rendezvous string
}

// DiscoveryNotifee gets notified when we find a new peer
type DiscoveryNotifee struct {
	h host.Host
}

func (n *DiscoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	if pi.ID == n.h.ID() {
		return // Do not connect to ourselves
	}
	fmt.Printf("[P2P] Peer Discovered via mDNS: %s\n", pi.ID.String())
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		fmt.Printf("[P2P] Connection failed: %v\n", err)
	} else {
		fmt.Printf("[P2P] Connected successfully to %s\n", pi.ID.String())
		// Emit event via SSE/WS
		if GlobalSSEBroker != nil {
			GlobalSSEBroker.Broadcast("info", fmt.Sprintf("[P2P] Linked with %s", pi.ID.String()))
		}
	}
}

var GlobalP2P *P2PNet

func InitializeP2PNode(brainRoot string) error {
	ctx := context.Background()

	h, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"), // Listen on a random available port
		libp2p.Ping(false),
	)
	if err != nil {
		return fmt.Errorf("failed to create libp2p node: %v", err)
	}

	log.Printf("[P2P] Local Node ID: %s", h.ID().String())
	
	for _, addr := range h.Addrs() {
		log.Printf("[P2P] Listening on: %s/p2p/%s", addr, h.ID().String())
	}

	GlobalP2P = &P2PNet{
		Host:       h,
		Rendezvous: "neuronfs-v5-mdns-rendezvous",
	}

	// Setup mDNS discovery
	notifee := &DiscoveryNotifee{h: h}
	ser := mdns.NewMdnsService(h, GlobalP2P.Rendezvous, notifee)
	err = ser.Start()
	if err != nil {
		return fmt.Errorf("failed to start mDNS: %v", err)
	}

	// Trigger Merkle Sync Worker (Phase 26)
	// if err := initMerkleSyncWorker(ctx, brainRoot); err != nil {
	// 	log.Printf("[P2P] Failed to init merkle sync: %v", err)
	// }

	// Wait asynchronously forever
	go func() {
		<-ctx.Done()
	}()

	return nil
}
