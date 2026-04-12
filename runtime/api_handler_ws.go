package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all cross-origin for dashboard
	},
}

// wsHub maintains the set of active clients and broadcasts messages to the clients.
type WSHub struct {
	clients    map[*WSClient]bool
	broadcast  chan []byte
	register   chan *WSClient
	unregister chan *WSClient
}

type WSClient struct {
	hub  *WSHub
	conn *websocket.Conn
	send chan []byte
}

var globalWSHub = &WSHub{
	broadcast:  make(chan []byte),
	register:   make(chan *WSClient),
	unregister: make(chan *WSClient),
	clients:    make(map[*WSClient]bool),
}

func init() {
	go globalWSHub.run()
}

func (h *WSHub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *WSHub) BroadcastEvent(data EventMessage) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return
	}
	h.broadcast <- jsonBytes
}

func (c *WSClient) readPump(brainRoot string) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		
		// Handle incoming Actions (Phase 21: UI Bidirectional Events)
		var payload struct {
			Action string `json:"action"`
			Source string `json:"source"`
			Target string `json:"target"`
		}
		if err := json.Unmarshal(message, &payload); err == nil {
			switch payload.Action {
			case "merge":
				fmt.Printf("[WS] Merge requested: %s -> %s\n", payload.Source, payload.Target)
				// TODO: Implement mergeNeuron(brainRoot, payload.Source, payload.Target)
			case "axon":
				fmt.Printf("[WS] Axon requested: %s -> %s\n", payload.Source, payload.Target)
				// TODO: Implement linkAxon(brainRoot, payload.Source, payload.Target)
			}
		}
	}
}

func (c *WSClient) writePump() {
	ticker := time.NewTicker(20 * time.Second) // Ping ticker
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func serveWs(brainRoot string, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &WSClient{hub: globalWSHub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in new goroutines.
	go client.writePump()
	go client.readPump(brainRoot)
}
